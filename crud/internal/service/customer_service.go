package service

import (
	"crud/internal/model"
	"errors"
)

type CustomerRepository interface {
	findAll() ([]model.Customer, error)
	findById(id int) (model.Customer, error)
	addCustomer(c model.Customer) (model.Customer, error)
	updateCustomer(c model.Customer, id int) (bool, error)
	deleteCustomer(id int) (bool, error)
}

var ErrInputNameEmail = errors.New("Name e email são obrigatórios")

type CustomerService struct {
	repo CustomerRepository
}

func NewCustomerService(repo CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List() ([]model.Customer, error) {
	return s.repo.findAll()
}

func (s *CustomerService) FindById(id int) (model.Customer, error) {
	return s.repo.findById(id)
}

func (s *CustomerService) Create(c model.Customer) (model.Customer, error) {
	if c.Email == "" || c.Name == "" {
		return model.Customer{}, ErrInputNameEmail
	}
	return s.repo.addCustomer(c)
}

func (s *CustomerService) Update(c model.Customer, id int) (bool, error) {
	if c.Email == "" || c.Name == "" {
		return false, ErrInputNameEmail
	}

	return s.repo.updateCustomer(c, id)
}

func (s *CustomerService) Delete(id int) (bool, error) {
	return s.repo.deleteCustomer(id)
}
