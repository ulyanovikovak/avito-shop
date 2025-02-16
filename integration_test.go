package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"avito-shop/internal/app"
	"avito-shop/internal/models"
)

// Для ответов
type AuthResponseWrapper struct {
	Description string `json:"description"`
	Token       string `json:"token"`
}

type InfoResponseWrapper struct {
	Description string              `json:"description"`
	Data        models.InfoResponse `json:"data"`
}

// Запуск приложения
func setupApp(t *testing.T) *app.App {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/avito?sslmode=disable"
	}
	a, err := app.InitializeApp(dbURL)
	if err != nil {
		t.Fatalf("Проблема с запуском: %v", err)
	}
	return a
}

// регистрация
func registerUser(t *testing.T, a *app.App, username, password string) string {
	t.Helper()
	authReq := models.AuthRequest{
		Username: username,
		Password: password,
	}
	body, err := json.Marshal(authReq)
	if err != nil {
		t.Fatalf("Ошибка при маршалинге запроса: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Ошибка при регистрации: статус %d", rr.Code)
	}
	var resp AuthResponseWrapper
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка при разборе ответа: %v", err)
	}
	return resp.Token
}

// Покупка
func TestBuyMerch(t *testing.T) {
	a := setupApp(t)

	username := "test_buy_user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	token := registerUser(t, a, username, "password123")

	req := httptest.NewRequest("GET", "/api/buy/cup", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Ошибка при покупке: статус %d", rr.Code)
	}

	// Получение проверка что покупка отображается у пользователя
	req = httptest.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Ошибка при получении информации: статус %d", rr.Code)
	}
	var infoResp InfoResponseWrapper
	if err := json.Unmarshal(rr.Body.Bytes(), &infoResp); err != nil {
		t.Fatalf("Ошибка при разборе информации: %v", err)
	}

	found := false
	for _, item := range infoResp.Data.Inventory {
		if strings.ToLower(item.Type) == "cup" {
			found = true
			if item.Quantity < 1 {
				t.Fatalf("Ожидается минимум 1 'cup', получено %d", item.Quantity)
			}
		}
	}
	if !found {
		t.Fatalf("Товар 'cup' не найден в инвентаре")
	}
}

// Перевод монет
func TestSendCoin(t *testing.T) {
	a := setupApp(t)

	senderUsername := "test_sender_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	receiverUsername := "test_receiver_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	senderToken := registerUser(t, a, senderUsername, "password123")
	receiverToken := registerUser(t, a, receiverUsername, "password123")

	transferReq := models.SendCoinRequest{
		ToUser: receiverUsername,
		Amount: 100,
	}
	body, err := json.Marshal(transferReq)
	if err != nil {
		t.Fatalf("Ошибка при маршалинге запроса передачи: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/sendCoin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+senderToken)
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Ошибка при передаче монет: статус %d", rr.Code)
	}

	// Проверка наличия перевода в истории у отправителя
	req = httptest.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+senderToken)
	rr = httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Ошибка при получении информации отправителя: статус %d", rr.Code)
	}
	var senderInfo InfoResponseWrapper
	if err := json.Unmarshal(rr.Body.Bytes(), &senderInfo); err != nil {
		t.Fatalf("Ошибка при разборе информации отправителя: %v", err)
	}
	foundSent := false
	for _, record := range senderInfo.Data.CoinHistory.Sent {
		if record.ToUser == receiverUsername && record.Amount == 100 {
			foundSent = true
			break
		}
	}
	if !foundSent {
		t.Fatalf("Запись о переводе не найдена в истории отправленных переводов")
	}

	// Проверка наличия перевода в истории у получателяы
	req = httptest.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+receiverToken)
	rr = httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Ошибка при получении информации получателя: статус %d", rr.Code)
	}
	var receiverInfo InfoResponseWrapper
	if err := json.Unmarshal(rr.Body.Bytes(), &receiverInfo); err != nil {
		t.Fatalf("Ошибка при разборе информации получателя: %v", err)
	}
	foundReceived := false
	for _, record := range receiverInfo.Data.CoinHistory.Received {
		if record.FromUser == senderUsername && record.Amount == 100 {
			foundReceived = true
			break
		}
	}
	if !foundReceived {
		t.Fatalf("Запись о переводе не найдена в истории полученных переводов")
	}
}
