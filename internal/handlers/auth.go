package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"avito-shop/internal/db"
	"avito-shop/internal/models"
	"avito-shop/internal/utils"
)

// обрабатывает запросы на вход и регистрацию
func AuthHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}
		user, err := db.GetUserByUsername(database, req.Username)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}

		if user == nil {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
				return
			}
			user, err = db.CreateUser(database, req.Username, string(hashedPassword))
			if err != nil {
				http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
				return
			}
		} else {
			// Если пользователь существует, проверяем пароль
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
				http.Error(w, "Неавторизован.", http.StatusUnauthorized)
				return
			}
		}

		token, err := utils.GenerateJWT(*user)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"description": "Успешный ответ.",
			"token":       token,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
	}
}
