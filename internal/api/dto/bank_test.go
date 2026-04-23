package dto

import (
	"fmt"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestFromPaymentRequestToBankReq(t *testing.T) {
	req := models.PostPaymentRequest{
		CardNumber:  "2222405343248871",
		ExpiryMonth: 4,
		ExpiryYear:  2026,
		Currency:    "GBP",
		Amount:      1050,
		Cvv:         "123",
	}

	got := FromPaymentRequestToBankReq(req)

	assert.Equal(t, req.CardNumber, got.CardNumber)
	assert.Equal(t, fmt.Sprintf("%d/%d", req.ExpiryMonth, req.ExpiryYear), got.ExpiryDate)
	assert.Equal(t, "GBP", got.Currency)
	assert.Equal(t, 1050, got.Amount)
	assert.Equal(t, "123", got.Cvv)
}

func TestFromBankResToPaymentResp_Authorized(t *testing.T) {
	req := models.PostPaymentRequest{
		CardNumber:  "2222405343248871",
		ExpiryMonth: 4,
		ExpiryYear:  2026,
		Currency:    "GBP",
		Amount:      1050,
	}
	bankRes := models.BankResponse{Authorized: true, AuthorizationCode: "abc-123"}

	got := FromBankResToPaymentResp(req, bankRes)

	assert.NotEmpty(t, got.ID)
	assert.Equal(t, models.Authorized, got.PaymentStatus)
	assert.Equal(t, "8871", got.CardNumberLastFour)
	assert.Equal(t, 4, got.ExpiryMonth)
	assert.Equal(t, 2026, got.ExpiryYear)
	assert.Equal(t, "GBP", got.Currency)
	assert.Equal(t, 1050, got.Amount)
}

func TestFromBankResToPaymentResp_Declined(t *testing.T) {
	req := models.PostPaymentRequest{
		CardNumber:  "2222405343248872",
		ExpiryMonth: 12,
		ExpiryYear:  2030,
		Currency:    "USD",
		Amount:      100,
	}
	bankRes := models.BankResponse{Authorized: false, AuthorizationCode: ""}

	got := FromBankResToPaymentResp(req, bankRes)

	assert.NotEmpty(t, got.ID)
	assert.Equal(t, models.Declined, got.PaymentStatus)
	assert.Equal(t, "8872", got.CardNumberLastFour)
}

func TestFromBankResToPaymentResp_UniqueIDs(t *testing.T) {
	req := models.PostPaymentRequest{
		CardNumber:  "2222405343248871",
		ExpiryMonth: 1,
		ExpiryYear:  2030,
		Currency:    "EUR",
		Amount:      500,
	}
	bankRes := models.BankResponse{Authorized: true}

	a := FromBankResToPaymentResp(req, bankRes)
	b := FromBankResToPaymentResp(req, bankRes)

	assert.NotEqual(t, a.ID, b.ID, "each payment must get a unique ID")
}
