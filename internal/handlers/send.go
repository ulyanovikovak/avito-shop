package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"avito-shop/internal/constants"
	"avito-shop/internal/db"
	"avito-shop/internal/models"
)

// перевод коинов
func SendCoinHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(constants.UserIDKey).(int)
		if !ok {
			http.Error(w, "Неавторизован.", http.StatusUnauthorized)
			return
		}
		var req models.SendCoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}

		receiver, err := db.GetUserByUsername(database, req.ToUser)
		if err != nil || receiver == nil {
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

		res, err := tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2 AND coins >= $1", req.Amount, userID)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		affected, err := res.RowsAffected()
		if err != nil || affected != 1 {
			http.Error(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec("UPDATE users SET coins = coins + $1 WHERE id = $2", req.Amount, receiver.ID)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("INSERT INTO coin_transfers (from_user, to_user, amount) VALUES ($1, $2, $3)", userID, receiver.ID, req.Amount)
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
		response := map[string]string{
			"description": "Успешный ответ.",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
	}
}
