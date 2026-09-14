package postgre

import (
	"crud/internal/model"
	"crud/internal/repository"
	"database/sql"
	"errors"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (cr *CustomerRepository) List() ([]model.Customer, error) {
	rows, err := cr.db.Query(`SELECT id , name , email FROM customers ORDER BY id`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []model.Customer

	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email); err != nil {
			return nil, err
		}

		list = append(list, c)
	}

	return list, rows.Err()
}

func (cr *CustomerRepository) GetById(id int) (model.Customer, error) {
	var c model.Customer
	err := cr.db.QueryRow(`SELECT id , name , email FROM customers WHERE id = $1 `, id).Scan(&c.ID, &c.Name, &c.Email)

	if errors.Is(err, sql.ErrNoRows) {
		return model.Customer{}, repository.ErrNotFound
	}

	if err != nil {
		return model.Customer{}, err
	}

	return c, nil
}
func (cr *CustomerRepository) Create(c model.Customer) (model.Customer, error) {
	err := cr.db.QueryRow(`INSERT INTO customers (name , email) VALUES ($1 , $2) RETURNING id`, c.Name, c.Email).Scan(&c.ID)
	if err != nil {
		return model.Customer{}, err
	}
	return c, nil
}
func (cr *CustomerRepository) Update(c model.Customer, id int) (bool, error) {
	res, err := cr.db.Exec(`UPDATE customers SET name = $1 , email = $2 WHERE id = $3`, c.Name, c.Email, id)
	if err != nil {
		return false, err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return false, repository.ErrNotFound
	}

	return true, nil
}

func (cr *CustomerRepository) Delete(id int) (bool, error) {
	res, err := cr.db.Exec(`DELETE FROM customers WHERE id = $1`, id)

	if err != nil {
		return false, err
	}

	n, _ := res.RowsAffected()

	if n == 0 {
		return false, repository.ErrNotFound
	}

	return true, nil
}
