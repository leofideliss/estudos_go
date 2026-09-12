package service

import (
	"crud/internal/model"
	"errors"
)

type CustomerRepository interface {
	List() ([]model.Customer, error)
	GetById(id int) (model.Customer, error)
	Create(c model.Customer) (model.Customer, error)
	Update(c model.Customer, id int) (bool, error)
	Delete(id int) (bool, error)
}

var ErrInputNameEmail = errors.New("Name e email são obrigatórios")

type CustomerService struct {
	repo CustomerRepository
}

func NewCustomerService(repo CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List() ([]model.Customer, error) {
	return s.repo.List()
}

func (s *CustomerService) GetById(id int) (model.Customer, error) {
	return s.repo.GetById(id)
}

func (s *CustomerService) Create(c model.Customer) (model.Customer, error) {
	if c.Email == "" || c.Name == "" {
		return model.Customer{}, ErrInputNameEmail
	}
	return s.repo.Create(c)
}

func (s *CustomerService) Update(c model.Customer, id int) (bool, error) {
	if c.Email == "" || c.Name == "" {
		return false, ErrInputNameEmail
	}

	return s.repo.Update(c, id)
}

func (s *CustomerService) Delete(id int) (bool, error) {
	return s.repo.Delete(id)
}
