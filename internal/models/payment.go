package models

import (
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/httputil"
)

type PostPaymentRequest struct {
	CardNumber  CardNumber `json:"card_number" validate:"required,numeric,min=14,max=19" example:"2222405343248871"`
	ExpiryMonth int        `json:"expiry_month" validate:"required,min=1,max=12" example:"4"`
	ExpiryYear  int        `json:"expiry_year" validate:"required,min=1" example:"2026"`
	Currency    string     `json:"currency" validate:"required,oneof=USD GBP EUR" example:"GBP"`
	Amount      int        `json:"amount" validate:"required,min=1" example:"100"`
	Cvv         string     `json:"cvv" validate:"required,numeric,min=3,max=4" example:"123"`
}

func (r *PostPaymentRequest) Validate() error {
	if err := validate.Struct(r); err != nil {
		return err
	}

	now := time.Now()

	if r.ExpiryYear < now.Year() {
		return &httputil.FieldError{Field: "expiry_year", Message: "card is expired"}
	}

	if r.ExpiryYear == now.Year() && r.ExpiryMonth < int(now.Month()) {
		return &httputil.FieldError{Field: "expiry_month", Message: "card is expired"}
	}

	return nil
}

type PaymentResponse struct {
	ID                 string        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PaymentStatus      PaymentStatus `json:"payment_status" enums:"Authorized,Declined" example:"Authorized"`
	CardNumberLastFour string        `json:"card_number_last_four" example:"8877"`
	ExpiryMonth        int           `json:"expiry_month"          example:"4"`
	ExpiryYear         int           `json:"expiry_year"           example:"2025"`
	Currency           string        `json:"currency"              example:"GBP"`
	Amount             int           `json:"amount"                example:"100"`
}
