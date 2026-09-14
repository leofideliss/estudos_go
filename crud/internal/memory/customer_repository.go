package memory

import (
	"sync"

	"crud/internal/model"
	"crud/internal/repository"
)

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

func (cr *CustomerRepository) List() ([]model.Customer, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	list := make([]model.Customer, 0, len(cr.customers))

	for _, c := range cr.customers {
		list = append(list, c)
	}

	return list, nil
}

func (cr *CustomerRepository) GetById(id int) (model.Customer, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	c, ok := cr.customers[id]
	if !ok {
		return model.Customer{}, repository.ErrNotFound
	}

	return c, nil
}

func (cr *CustomerRepository) Create(c model.Customer) (model.Customer, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	c.ID = cr.nextID
	cr.customers[c.ID] = c
	cr.nextID++

	return c, nil
}

func (cr *CustomerRepository) Update(c model.Customer, id int) (bool, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, ok := cr.customers[id]; !ok {
		return false, repository.ErrNotFound
	}
	c.ID = id
	cr.customers[id] = c
	return true, nil
}

func (cr *CustomerRepository) Delete(id int) (bool, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, ok := cr.customers[id]; !ok {
		return false, repository.ErrNotFound
	}

	delete(cr.customers, id)
	return true, nil
}
