package app

import (
	"database/sql"
	"log"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"avito-shop/internal/handlers"
	"avito-shop/internal/middleware"
)

// роутер и база данных
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// подключение к базе
func InitializeApp(dbURL string) (*App, error) {
	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err = database.Ping(); err != nil {
		return nil, err
	}

	a := &App{
		Router: mux.NewRouter(),
		DB:     database,
	}

	a.initializeRoutes()
	log.Println("Приложение инициализировано")
	return a, nil
}

// распределение запросов для API
func (a *App) initializeRoutes() {
	//  аутентификации
	a.Router.HandleFunc("/api/auth", handlers.AuthHandler(a.DB)).Methods("POST")

	// маршруты с проверкой авторизации через middleware
	api := a.Router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JWTMiddleware)
	api.HandleFunc("/info", handlers.InfoHandler(a.DB)).Methods("GET")
	api.HandleFunc("/sendCoin", handlers.SendCoinHandler(a.DB)).Methods("POST")
	api.HandleFunc("/buy/{item}", handlers.BuyHandler(a.DB)).Methods("GET")
}
