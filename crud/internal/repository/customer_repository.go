package repository

import (
	"errors"

	"crud/internal/model"
)

var ErrNotFound = errors.New("cliente não encontrado")

type CustomerRepository interface {
	List() ([]model.Customer, error)
	GetById(id int) (model.Customer, error)
	Create(c model.Customer) (model.Customer, error)
	Update(c model.Customer, id int) (bool, error)
	Delete(id int) (bool, error)
}
