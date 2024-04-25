package database

const (
	CreateСheckTablePeoples = `
		CREATE TABLE IF NOT EXISTS peoples (
			id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			surname VARCHAR(255) NOT NULL,
		    patronymic VARCHAR(255)   
		);`
	CreateCheckTableCars = `
		CREATE TABLE IF NOT EXISTS cars (
			id INT GENERATED ALWAYS AS IDENTITY,
			reg_num VARCHAR(255) NOT NULL UNIQUE,
			mark VARCHAR(255) NOT NULL,
			model VARCHAR(255) NOT NULL,
			year INT,
			owner_id INT NOT NULL,
			FOREIGN KEY (owner_id) REFERENCES peoples(id)
		);`
	GridCarInfo = `
		SELECT cars.id, cars.reg_num, cars.mark, cars.model, cars.year, peoples.id, peoples.name, peoples.surname, peoples.patronymic
		FROM cars
		JOIN peoples ON cars.owner_id = peoples.id
		WHERE ($1 = '%' OR cars.mark LIKE $1) AND ($2 = '%' OR cars.model LIKE $2) AND ($3 = 0 OR cars.year = $3)
		ORDER BY cars.id
		LIMIT $4 OFFSET $5;`
	GridOneCarInfo = `
		SELECT cars.id, cars.reg_num, cars.mark, cars.model, cars.year, peoples.id, peoples.name, peoples.surname, peoples.patronymic
		FROM cars
		JOIN peoples ON cars.owner_id = peoples.id
		WHERE cars.id = $1;`
	OwnerExists    = `SELECT EXISTS(SELECT 1 FROM peoples WHERE id = $1)`
	DeleteCar      = `DELETE FROM cars WHERE id = $1;`
	CheckCarExists = `
		SELECT id
		FROM cars
		WHERE reg_num = $1;`
	UpdateCarInfo = `
		UPDATE cars
		SET mark = COALESCE(NULLIF($1, ''), mark),
			model = COALESCE(NULLIF($2, ''), model),
			year = COALESCE(NULLIF($3, 0), year),
			owner_id = COALESCE($5, owner_id)
		WHERE id = $4;`
	AddNewCar = `
		INSERT INTO cars (reg_num, mark, model, year, owner_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;`
	СheckPerson = `
		SELECT id
		FROM peoples
		WHERE name = $1 AND surname = $2;`
	AddPerson = `
		INSERT INTO peoples (name, surname)
		VALUES ($1, $2)
		RETURNING id`
)
