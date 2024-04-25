package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Car struct {
	ID     int    `json:"id"`
	RegNum string `json:"regNum"`
	Mark   string `json:"mark"`
	Model  string `json:"model"`
	Year   int    `json:"year"`
	Owner  Owner  `json:"owner"`
}

type Owner struct {
	ID         int     `json:"ownerId"`
	Name       string  `json:"name"`
	Surname    string  `json:"surname"`
	Patronymic *string `json:"patronymic"`
}

var ErrCarExists = errors.New("a car with that plate is already in the database")

func (db *Database) GridCarInfo(ctx context.Context, mark, model string, year, limit, offset int) ([]Car, error) {
	markParam := "%" + mark + "%"
	modelParam := "%" + model + "%"
	if mark == "" {
		markParam = "%"
	}
	if model == "" {
		modelParam = "%"
	}
	yearParam := year
	if year == 0 {
		yearParam = 0
	}

	rows, err := db.QueryContext(ctx, GridCarInfo, markParam, modelParam, yearParam, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying cars: %v", err)
	}
	defer rows.Close()

	var cars []Car
	for rows.Next() {
		var car Car
		if err := rows.Scan(&car.ID, &car.RegNum, &car.Mark, &car.Model, &car.Year, &car.Owner.ID, &car.Owner.Name, &car.Owner.Surname, &car.Owner.Patronymic); err != nil {
			continue
		}
		cars = append(cars, car)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}
	return cars, nil
}

func (db *Database) GetCar(ctx context.Context, id int) (*Car, error) {
	var car Car
	var owner Owner
	car.Owner = owner
	fmt.Println(id)
	err := db.QueryRowContext(ctx, GridOneCarInfo, id).Scan(&car.ID, &car.RegNum, &car.Mark, &car.Model, &car.Year, &car.Owner.ID, &car.Owner.Name, &car.Owner.Surname, &car.Owner.Patronymic)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("Error querying GetCar: %v", err)
	}
	return &car, nil
}

func (db *Database) DeleteCar(ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, DeleteCar, id)
	if err != nil {
		return fmt.Errorf("error deleting car: %v", err)
	}
	return err
}

func (db *Database) UpdateCarInfo(ctx context.Context, id int, mark, model string, year, ownerId int) error {
	_, err := db.ExecContext(ctx, UpdateCarInfo, mark, model, year, id, ownerId)
	if err != nil {
		return fmt.Errorf("error updating car info: %v", err)
	}
	return err
}

func (db *Database) OwnerExists(ctx context.Context, ownerId int) bool {
	var exists bool
	err := db.QueryRowContext(ctx, OwnerExists, ownerId).Scan(&exists)
	if err != nil {
		log.Printf("error checking if owner exists: %v", err)
		return false
	}
	return exists
}

func (db *Database) AddNewCar(ctx context.Context, regNum, mark, model string, year int, ownerId int64) (int64, error) {
	var existingId, newId int64
	err := db.QueryRowContext(ctx, CheckCarExists, regNum).Scan(&existingId)
	if err == nil {
		return existingId, ErrCarExists
	} else if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error checking car existence: %v", err)
	}

	err = db.QueryRowContext(ctx, AddNewCar, regNum, mark, model, year, ownerId).Scan(&newId)
	if err != nil {
		return 0, fmt.Errorf("error adding new car: %v", err)
	}
	return newId, nil
}

func (db *Database) GetOrCreateOwner(ctx context.Context, owner Owner) (int64, error) {
	var ownerId int64
	err := db.QueryRowContext(ctx, СheckPerson, owner.Name, owner.Surname).Scan(&ownerId)
	if err != nil {
		if err == sql.ErrNoRows {
			err = db.QueryRowContext(ctx, AddPerson, owner.Name, owner.Surname).Scan(&ownerId)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}
	return ownerId, nil
}
