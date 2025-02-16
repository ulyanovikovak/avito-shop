package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"avito-shop/internal/constants"
)

// запросы на покупку мерча
func BuyHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(constants.UserIDKey).(int)
		if !ok {
			http.Error(w, "Неавторизован.", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		item := vars["item"]
		var price int
		if err := database.QueryRow("SELECT price FROM merch WHERE name = $1", item).Scan(&price); err != nil {
			http.Error(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}

		tx, err := database.Begin()
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
				return
			}
		}()

		res, err := tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2 AND coins >= $1", price, userID)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		affected, err := res.RowsAffected()
		if err != nil || affected != 1 {
			http.Error(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec("INSERT INTO purchases (user_id, item, price) VALUES ($1, $2, $3)", userID, item, price)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		if err = tx.Commit(); err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"description": "Успешный ответ.",
		}); err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
		}
	}
}
