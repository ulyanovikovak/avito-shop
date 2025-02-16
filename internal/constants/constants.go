package constants

// для подписи JWT-токенов
const JWTSecret = "my_super_secret_key_which_is_long_enough_123456"

// ключ
type contextKey string

const UserIDKey contextKey = "userID"
