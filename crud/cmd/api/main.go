package main

import (
	"crud/internal/handler"
	"crud/internal/postgre"
	"crud/internal/service"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL não definida")
	}
	db, err := connectDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	customerRepo := postgre.NewCustomerRepository(db)
	//	customerRepo := memory.NewCustomerRepository()
	custoemerService := service.NewCustomerService(customerRepo)
	customerHandler := handler.NewCustomeHandler(custoemerService)

	customerHandler.RegisterRoutes(mux)

	log.Println("rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func connectDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrindo conexão: %w", err)
	}

	for i := 1; i <= 10; i++ {
		if err = db.Ping(); err == nil {
			log.Println("conectado ao banco")
			return db, nil
		}
		log.Printf("banco ainda não respondeu (tentativa %d/10): %v", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("banco não respondeu após 10 tentativas: %w", err)
}
