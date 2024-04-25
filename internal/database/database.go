package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/likimiad/car-management-api/internal/config"
	"log"
	"time"
)

type Database struct {
	*sql.DB
}

func InitDatabase(cfg config.DatabaseConfig) *Database {
	defer func(start time.Time) {
		fmt.Printf("%s [%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), "START", "make connection with database", time.Since(start))
	}(time.Now())

	db := makeConnection(cfg)
	db.initTables()
	return db
}

func makeConnection(cfg config.DatabaseConfig) *Database {
	dbLink := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := sql.Open("postgres", dbLink)
	if err != nil {
		log.Fatalf("error connection to database: %s", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("error during connection check: %s", err.Error())
	}

	return &Database{db}
}

func (db *Database) initTables() {
	defer func(start time.Time) {
		fmt.Printf("%s [%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), "START", "checking the availability of the database", time.Since(start))
	}(time.Now())
	db.checkTables()
}

func (db *Database) checkTables() {
	_, err := db.Exec(CreateСheckTablePeoples)
	if err != nil {
		log.Fatalf("error database during people table initialization: %s", err.Error())
	}
	_, err = db.Exec(CreateCheckTableCars)
	if err != nil {
		log.Fatalf("error database during cars table initialization: %s", err.Error())
	}
}
