package db

import (
	"database/sql"
	"errors"

	"avito-shop/internal/models"
)

// получения пользователя по имени
func GetUserByUsername(database *sql.DB, username string) (*models.User, error) {
	row := database.QueryRow("SELECT id, username, password, coins FROM users WHERE username = $1", username)
	u := &models.User{}
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Coins); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// создание нового пользователя
func CreateUser(database *sql.DB, username, password string) (*models.User, error) {
	var id int
	err := database.QueryRow("INSERT INTO users (username, password, coins) VALUES ($1, $2, 1000) RETURNING id", username, password).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:       id,
		Username: username,
		Password: password,
		Coins:    1000,
	}, nil
}
