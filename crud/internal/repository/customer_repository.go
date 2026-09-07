package repository

import (
	"errors"
	"sync"

	"crud/internal/model"
)

var ErrorNotFound = errors.New("Cliente não encontrado")

type CustomerRepository struct {
	mu        sync.Mutex
	customers map[int]model.Customer
	nextID    int
}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		customers: make(map[int]model.Customer),
		nextID:    1,
	}
}

func (cr *CustomerRepository) findAll() ([]model.Customer, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	list := make([]model.Customer, 0, len(cr.customers))

	for _, c := range cr.customers {
		list = append(list, c)
	}

	return list, nil
}

func (cr *CustomerRepository) findById(id int) (model.Customer, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	c, ok := cr.customers[id]
	if !ok {
		return model.Customer{}, ErrorNotFound
	}

	return c, nil
}
