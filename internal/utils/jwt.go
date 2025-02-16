package utils

import (
	"time"

	"avito-shop/internal/constants"
	"avito-shop/internal/models"

	"github.com/golang-jwt/jwt/v4"
)

// генерирует JWT токен для пользователя
func GenerateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(constants.JWTSecret))
}
