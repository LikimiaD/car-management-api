package main

import (
	"github.com/likimiad/car-management-api/api"
	"github.com/likimiad/car-management-api/internal/config"
	"github.com/likimiad/car-management-api/internal/database"
	"log"
)

// @title Effective Mobile Go API
// @version 0.0.1
// @description API Server for registration car plates in Effective Mobile

func main() {
	cfg := config.GetConfig()
	db := database.InitDatabase(cfg.DatabaseConfig)
	server := api.NewServer(db, cfg.HTTPServer)
	if err := server.Start(cfg.HTTPServer.Address); err != nil {
		log.Fatalf("Failed to start server %s", err.Error())
	}
}
