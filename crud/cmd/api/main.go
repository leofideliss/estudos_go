package main

import (
	"crud/internal/handler"
	"crud/internal/repository"
	"crud/internal/service"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	customerRepo := repository.NewCustomerRepository()
	custoemerService := service.NewCustomerService(customerRepo)
	customerHandler := handler.NewCustomeHandler(custoemerService)

	customerHandler.RegisterRoutes(mux)

	log.Println("rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
