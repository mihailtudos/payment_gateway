package api

import (
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/httputil"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

// PaymentResponseEnvelope is a swaggo-only type for documenting the POST /payments success response.
type PaymentResponseEnvelope struct {
	Success   bool                   `json:"success"`
	Data      models.PaymentResponse `json:"data,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// ErrorResponseEnvelope wraps an API error in the standard envelope.
type ErrorResponseEnvelope struct {
	Success   bool               `json:"success"`
	Error     *httputil.APIError `json:"error,omitempty"`
	Timestamp string             `json:"timestamp"`
}

// DeclinedPaymentResponseEnvelope is the 402 response for bank-declined payments.
// It includes the stored payment record alongside the decline error so callers
// have the payment ID and status for their own records.
type DeclinedPaymentResponseEnvelope struct {
	Success   bool                   `json:"success"`
	Data      models.PaymentResponse `json:"data,omitempty"`
	Error     *httputil.APIError     `json:"error,omitempty"`
	Timestamp string                 `json:"timestamp"`
}
