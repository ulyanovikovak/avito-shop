package main

import (
	"log"
	"net/http"
	"os"

	"avito-shop/internal/app"
)

func main() {
	// если база не задана, будет база по умолчанию
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@db:5432/avito?sslmode=disable"
	}

	// Запуск сервера
	a, err := app.InitializeApp(dbURL)
	if err != nil {
		log.Fatalf("Ошибка инициализации приложения: %v", err)
	}

	log.Println("Сервис запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", a.Router))
}
