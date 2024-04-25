package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/likimiad/car-management-api/internal/database"
	"net/http"
	"runtime"
	"time"
)

var ErrCarNotFound = errors.New("car with registration number not found")

func (s *Server) fetchCarInfoWithContext(ctx context.Context, regNum string, timeout time.Duration) (database.Car, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, "GET", fmt.Sprintf("%s/info?regNum=%s", s.ThirdPartyAPIURL, regNum), nil)
	if err != nil {
		return database.Car{}, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return database.Car{}, fmt.Errorf("failed to fetch car info: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var car database.Car
		if err := json.NewDecoder(resp.Body).Decode(&car); err != nil {
			return database.Car{}, fmt.Errorf("failed to decode car info: %v", err)
		}
		return car, nil
	case http.StatusNotFound:
		return database.Car{}, ErrCarNotFound
	default:
		return database.Car{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}
}

func (s *Server) debugErrorMessage(err error) {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		fmt.Printf("\t[%s] Occurred in: %s\n\t\tFile: %s, Line: %d\n", "Error", funcName, file, line)
	} else {
		fmt.Printf("\t[%s] %s\n", "Error", err.Error())
	}
}

func (s *Server) workerMessage(id int, message string) {
	fmt.Printf("\t[%s] id: %d %s\n", "WORKER", id, message)
}
