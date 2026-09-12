package handler

import (
	"crud/internal/model"
	"crud/internal/repository"
	"crud/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type CustomerService interface {
	List() ([]model.Customer, error)
	GetById(id int) (model.Customer, error)
	Create(c model.Customer) (model.Customer, error)
	Update(c model.Customer, id int) (bool, error)
	Delete(id int) (bool, error)
}

type CustomerHanlder struct {
	service CustomerService
}

func NewCustomeHandler(s CustomerService) *CustomerHanlder {
	return &CustomerHanlder{service: s}
}

func (h *CustomerHanlder) RegisterRoutes(mx *http.ServeMux) {
	mx.HandleFunc("GET /customers", h.List)
	mx.HandleFunc("POST /customers", h.Create)
	mx.HandleFunc("GET /customers/{id}", h.GetById)
	mx.HandleFunc("PUT /customers/{id}", h.Update)
	mx.HandleFunc("DELETE /customers/{id}", h.Delete)
}

func (h *CustomerHanlder) List(w http.ResponseWriter, r *http.Request) {
	customers, error := h.service.List()
	if error != nil {
		WriteError(w, http.StatusInternalServerError, "erro ao listar")
		return
	}

	WriteJSON(w, http.StatusCreated, customers)
}
func (h *CustomerHanlder) Create(w http.ResponseWriter, r *http.Request) {
	var input model.Customer
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadGateway, "JSON Inválido")
	}

	create, error := h.service.Create(input)
	if error != nil {
		if errors.Is(error, service.ErrInputNameEmail) {
			WriteError(w, http.StatusBadGateway, error.Error())
			return
		}
		WriteError(w, http.StatusBadGateway, "Erro ao criar")
		return
	}

	WriteJSON(w, http.StatusCreated, create)
}
func (h *CustomerHanlder) GetById(w http.ResponseWriter, r *http.Request) {
	id, error := idFromPath(r)
	if error != nil {
		WriteError(w, http.StatusBadGateway, "id inválido")
		return
	}

	customer, err := h.service.GetById(id)
	if err != nil {
		if errors.Is(err, repository.ErrorNotFound) {
			WriteError(w, http.StatusBadGateway, err.Error())
			return
		}
		WriteError(w, http.StatusBadGateway, "Erro ao consultar")

		return
	}

	WriteJSON(w, http.StatusCreated, customer)
}
func (h *CustomerHanlder) Update(w http.ResponseWriter, r *http.Request) {
	var input model.Customer
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadGateway, "JSON Inválido")
		return
	}

	id, error := idFromPath(r)
	if error != nil {
		WriteError(w, http.StatusBadGateway, "id inválido")
		return
	}
	updated, error := h.service.Update(input, id)

	if error != nil {
		if errors.Is(error, service.ErrInputNameEmail) {
			WriteError(w, http.StatusBadGateway, error.Error())
			return
		}
		WriteError(w, http.StatusBadGateway, "Erro ao atualizar")
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}

func (h *CustomerHanlder) Delete(w http.ResponseWriter, r *http.Request) {
	id, error := idFromPath(r)
	if error != nil {
		WriteError(w, http.StatusBadGateway, "id inválido")
		return
	}

	_, err := h.service.Delete(id)

	if err != nil {
		WriteError(w, http.StatusBadGateway, "Erro ao deletar")
		return
	}

	WriteJSON(w, http.StatusNoContent, nil)
}

func idFromPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}
