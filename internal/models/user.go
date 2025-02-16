package models

// пользователь
type User struct {
	ID       int
	Username string
	Password string
	Coins    int
}

// информация
type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

// история переводов.
type CoinHistory struct {
	Received []TransferRecord `json:"received"`
	Sent     []TransferRecord `json:"sent"`
}

// купленный товар
type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}
