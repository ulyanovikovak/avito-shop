package models

// запрос аутентификации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// токен после аутентификации
type AuthResponse struct {
	Token string `json:"token"`
}
