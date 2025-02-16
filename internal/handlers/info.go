package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"avito-shop/internal/constants"
	"avito-shop/internal/models"
)

// возвращает информацию о пользователе
func InfoHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(constants.UserIDKey).(int)
		if !ok {
			http.Error(w, "Неавторизован.", http.StatusUnauthorized)
			return
		}

		var coins int
		if err := database.QueryRow("SELECT coins FROM users WHERE id = $1", userID).Scan(&coins); err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}

		// список покупок
		rows, err := database.Query("SELECT item, COUNT(*) FROM purchases WHERE user_id = $1 GROUP BY item", userID)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var inventory []models.InventoryItem
		for rows.Next() {
			var item string
			var quantity int
			if err := rows.Scan(&item, &quantity); err != nil {
				http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
				return
			}
			inventory = append(inventory, models.InventoryItem{Type: item, Quantity: quantity})
		}

		// история переводово
		receivedRows, err := database.Query(`
			SELECT u.username, ct.amount FROM coin_transfers ct
			JOIN users u ON ct.from_user = u.id
			WHERE ct.to_user = $1
		`, userID)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		defer receivedRows.Close()

		var received []models.TransferRecord
		for receivedRows.Next() {
			var fromUsername string
			var amount int
			if err := receivedRows.Scan(&fromUsername, &amount); err != nil {
				http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
				return
			}
			received = append(received, models.TransferRecord{FromUser: fromUsername, Amount: amount})
		}

		sentRows, err := database.Query(`
			SELECT u.username, ct.amount FROM coin_transfers ct
			JOIN users u ON ct.to_user = u.id
			WHERE ct.from_user = $1
		`, userID)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
		defer sentRows.Close()

		var sent []models.TransferRecord
		for sentRows.Next() {
			var toUsername string
			var amount int
			if err := sentRows.Scan(&toUsername, &amount); err != nil {
				http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
				return
			}
			sent = append(sent, models.TransferRecord{ToUser: toUsername, Amount: amount})
		}

		info := models.InfoResponse{
			Coins:       coins,
			Inventory:   inventory,
			CoinHistory: models.CoinHistory{Received: received, Sent: sent},
		}

		response := map[string]interface{}{
			"description": "Успешный ответ.",
			"data":        info,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}
	}
}
