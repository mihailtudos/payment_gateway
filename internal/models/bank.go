package models

// BankResponse
type BankResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code" example:"0bb07405-6d44-4b50-a14f-7ae0beff13ad"`
}

// BankRequest
type BankRequest struct {
	CardNumber CardNumber `json:"card_number"`
	ExpiryDate string     `json:"expiry_date"`
	Currency   string     `json:"currency"`
	Amount     int        `json:"amount"`
	Cvv        string     `json:"cvv"`
}
