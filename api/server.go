package api

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/likimiad/car-management-api/docs"
	"github.com/likimiad/car-management-api/internal/config"
	"github.com/likimiad/car-management-api/internal/database"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"time"
)

type Server struct {
	DB               *database.Database
	Router           *mux.Router
	Timeout          time.Duration
	IdleTimeout      time.Duration
	MaxWorkers       int
	ThirdPartyAPIURL string
	DebugMode        bool
}

func getServer(db *database.Database, cfg config.HTTPServer) *Server {
	server := &Server{
		DB:               db,
		Router:           mux.NewRouter(),
		Timeout:          cfg.Timeout,
		IdleTimeout:      cfg.IdleTimeout,
		MaxWorkers:       cfg.MaxWorkers,
		ThirdPartyAPIURL: cfg.ThirdPartyAPIURL,
		DebugMode:        cfg.DebugMode,
	}
	server.routes()
	return server
}

func NewServer(db *database.Database, cfg config.HTTPServer) *Server {
	defer func(start time.Time) {
		fmt.Printf("%s [%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), "START", "create server and routes", time.Since(start))
	}(time.Now())
	return getServer(db, cfg)
}

func (s *Server) Start(address string) error {
	fmt.Printf("%s [%s] %s http://%s\n", time.Now().Format("2006-01-02 15:04:05"), "START", "starting server", address)
	return http.ListenAndServe(address, s.Router)
}

func (s *Server) routes() {
	s.Router.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)

	s.Router.Handle("/api/cars", s.logger(s.handleGetCars())).Methods("GET")
	s.Router.Handle("/api/cars/{id}", s.logger(s.handleGetCar())).Methods("GET")
	s.Router.Handle("/api/cars", s.logger(s.handlePostCar())).Methods("POST")
	s.Router.Handle("/api/cars/{id}", s.logger(s.handleDeleteCar())).Methods("DELETE")
	s.Router.Handle("/api/cars/{id}", s.logger(s.handleUpdateCar())).Methods("PUT")
}

func (s *Server) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			fmt.Printf("%s [%s] %s %s %s\n", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
		}(time.Now())
		next.ServeHTTP(w, r)
	})
}
