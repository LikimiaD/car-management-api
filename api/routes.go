package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/likimiad/car-management-api/internal/database"
	"net/http"
	"strconv"
	"sync"
)

type plateResult struct {
	InputPlate string `json:"inputPlate"`
	ID         *int64 `json:"id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// @Summary Get list of cars
// @Description Get cars with optional filtering by mark, model, and year, with pagination
// @Tags cars
// @Accept  json
// @Produce  json
// @Param   mark    query     string     false  "Filter by car mark"
// @Param   model   query     string     false  "Filter by car model"
// @Param   year    query     int        false  "Filter by car year"
// @Param   limit   query     int        false  "Limit number of cars returned"
// @Param   offset  query     int        false  "Offset where to start fetching cars"
// @Success 200 {array} database.Car
// @Router /api/cars [get]
func (s *Server) handleGetCars() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		getParam := func(key string, defaultVal int) int {
			if valStr := r.URL.Query().Get(key); valStr != "" {
				if val, err := strconv.Atoi(valStr); err == nil {
					return val
				}
			}
			return defaultVal
		}

		ctx, cancel := context.WithTimeout(r.Context(), s.Timeout)
		defer cancel()

		mark := r.URL.Query().Get("mark")
		model := r.URL.Query().Get("model")
		year := getParam("year", 0)
		limit := getParam("limit", 10)
		offset := getParam("offset", 0)
		if limit < 0 {
			s.respondWithError(w, http.StatusBadRequest, "limit cannot be negative")
			return
		}
		if offset < 0 {
			s.respondWithError(w, http.StatusBadRequest, "offset cannot be negative")
			return
		}

		cars, err := s.DB.GridCarInfo(ctx, mark, model, year, limit, offset)
		if err != nil {
			if s.DebugMode {
				s.debugErrorMessage(err)
			}
			s.respondWithError(w, http.StatusInternalServerError, "Server error")
			return
		}

		s.respondAny(w, http.StatusOK, cars)
	}
}

// @Summary Get a car
// @Description Get details of a car by its ID
// @Tags cars
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Success 200 {object} database.Car "Successfully retrieved the car"
// @Failure 400 {string} string "Invalid car ID"
// @Failure 404 {string} string "Car not found"
// @Failure 500 {string} string "Server error"
// @Router /api/cars/{id} [get]
func (s *Server) handleGetCar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr, ok := vars["id"]
		if !ok {
			s.respondWithError(w, http.StatusBadRequest, "car ID is missing")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			s.respondWithError(w, http.StatusBadRequest, "invalid car ID")
			return
		}

		car, err := s.DB.GetCar(r.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				s.respondWithError(w, http.StatusNotFound, "car not found")
			} else {
				s.respondWithError(w, http.StatusInternalServerError, "server error")
			}
			return
		}

		s.respondAny(w, http.StatusOK, car)
	}
}

// @Summary Add new cars
// @Description Add new cars using registration numbers
// @Tags cars
// @Accept json
// @Produce json
// @Param regNums body []string true "Array of registration numbers"
// @Success 201 {array} plateResult "Successfully added cars with results for each plate"
// @Failure 400 {string} string "Bad request due to malformed JSON input"
// @Failure 500 {string} string "Server error"
// @Router /api/cars [post]
func (s *Server) handlePostCar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var regNums struct {
			Plates []string `json:"regNums"`
		}
		if err := json.NewDecoder(r.Body).Decode(&regNums); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results := make([]plateResult, len(regNums.Plates))
		ch := make(chan int, s.MaxWorkers)
		resultCh := make(chan plateResult)

		for i := 0; i < s.MaxWorkers; i++ {
			ch <- i
		}

		wg := sync.WaitGroup{}
		for idx, plate := range regNums.Plates {
			wg.Add(1)
			go func(index int, plate string) {
				defer wg.Done()
				workerID := <-ch
				if s.DebugMode {
					s.workerMessage(workerID, "start working")
				}

				car, err := s.fetchCarInfoWithContext(r.Context(), plate, s.IdleTimeout)
				if errors.Is(err, ErrCarNotFound) {
					resultCh <- plateResult{InputPlate: plate, Error: "car with registration number not found"}
					ch <- workerID
					return
				} else if err != nil {
					if s.DebugMode {
						s.debugErrorMessage(err)
					}
					resultCh <- plateResult{InputPlate: plate, Error: "error with getting data from third party api"}
					ch <- workerID
					return
				}

				ownerId, err := s.DB.GetOrCreateOwner(r.Context(), car.Owner)
				if err != nil {
					if s.DebugMode {
						s.debugErrorMessage(err)
					}
					resultCh <- plateResult{InputPlate: plate, Error: err.Error()}
					ch <- workerID
					return
				}

				carID, err := s.DB.AddNewCar(r.Context(), car.RegNum, car.Mark, car.Model, car.Year, ownerId)
				if errors.Is(err, database.ErrCarExists) {
					resultCh <- plateResult{InputPlate: plate, ID: &carID}
				} else if err != nil {
					if s.DebugMode {
						s.debugErrorMessage(err)
					}
					resultCh <- plateResult{InputPlate: plate, Error: "error while adding car in database"}
				} else {
					resultCh <- plateResult{InputPlate: plate, ID: &carID}
				}
				ch <- workerID
				if s.DebugMode {
					s.workerMessage(workerID, "successfully completed the job")
				}
			}(idx, plate)
		}

		go func() {
			wg.Wait()
			close(ch)
			close(resultCh)
		}()

		idx := 0
		for res := range resultCh {
			results[idx] = res
			idx++
		}

		s.respondAny(w, http.StatusCreated, results)
	}
}

// @Summary Delete a car
// @Description Delete a car by its ID
// @Tags cars
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Success 204 {object} nil
// @Failure 400 {string} string "Invalid car ID"
// @Failure 404 {string} string "Car not found"
// @Failure 500 {string} string "Error while deleting the car"
// @Router /api/cars/{id} [delete]
func (s *Server) handleDeleteCar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		idStr, ok := vars["id"]
		if !ok {
			s.respondWithError(w, http.StatusBadRequest, "missing car ID")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			s.respondWithError(w, http.StatusBadRequest, "invalid car ID")
			return
		}

		if err := s.DB.DeleteCar(r.Context(), id); err != nil {
			if s.DebugMode {
				s.debugErrorMessage(err)
			}
			s.respondWithError(w, http.StatusInternalServerError, "error while deleting car")
			return
		}

		s.respondNoContent(w, http.StatusNoContent)
	}
}

// @Summary Update a car
// @Description Update car's information by its ID
// @Tags cars
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Param car body database.Car true "Car data"
// @Success 204 {object} nil
// @Failure 400 {object} nil "Bad Request"
// @Failure 404 {object} nil "Car not found"
// @Router /api/cars/{id} [put]
func (s *Server) handleUpdateCar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr, ok := vars["id"]
		if !ok {
			s.respondWithError(w, http.StatusBadRequest, "missing car ID")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			s.respondWithError(w, http.StatusBadRequest, "invalid car ID")
			return
		}

		var car database.Car
		if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
			s.respondWithError(w, http.StatusBadRequest, "error parsing car data")
			return
		}

		if !s.DB.OwnerExists(r.Context(), car.Owner.ID) {
			s.respondWithError(w, http.StatusBadRequest, "owner does not exist")
			return
		}

		if err := s.DB.UpdateCarInfo(r.Context(), id, car.Mark, car.Model, car.Year, car.Owner.ID); err != nil {
			if s.DebugMode {
				s.debugErrorMessage(err)
			}
			s.respondWithError(w, http.StatusInternalServerError, "error while updating car information")
			return
		}

		s.respondNoContent(w, http.StatusNoContent)
	}
}
