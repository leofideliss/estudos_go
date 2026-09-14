package postgre

import (
	"crud/internal/model"
	"database/sql"
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

func (cr *CustomerRepository) GetById(id int) (model.Customer, error)          {}
func (cr *CustomerRepository) Create(c model.Customer) (model.Customer, error) {}
func (cr *CustomerRepository) Update(c model.Customer, id int) (bool, error)   {}
func (cr *CustomerRepository) Delete(id int) (bool, error)                     {}
