//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envelope mirrors httputil.Envelope for decoding responses.
type envelope[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

func decodeData[T any](t *testing.T, body io.Reader) T {
	t.Helper()
	var env envelope[T]
	require.NoError(t, json.NewDecoder(body).Decode(&env))
	return env.Data
}

// baseRequest returns a valid payment body. Callers override fields to steer
// the bank simulator: odd last digit → authorized, even → declined, 0 → 503.
func baseRequest() map[string]any {
	return map[string]any{
		"card_number":  "2222405343248871", // ends in 1 → authorized
		"expiry_month": 12,
		"expiry_year":  2030,
		"currency":     "GBP",
		"amount":       100,
		"cvv":          "123",
	}
}

func postPayment(t *testing.T, body map[string]any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(fmt.Sprintf("%s/api/payments", gatewayURL), "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	return resp
}

func TestPostPayment_Authorized(t *testing.T) {
	resp := postPayment(t, baseRequest())
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	payment := decodeData[models.PaymentResponse](t, resp.Body)
	assert.Equal(t, models.Authorized, payment.PaymentStatus)
	assert.NotEmpty(t, payment.ID)
	assert.Equal(t, "8871", payment.CardNumberLastFour)
	assert.Equal(t, "GBP", payment.Currency)
	assert.Equal(t, 100, payment.Amount)
}

func TestPostPayment_Declined(t *testing.T) {
	body := baseRequest()
	body["card_number"] = "2222405343248872" // ends in 2 → declined

	resp := postPayment(t, body)
	defer resp.Body.Close()

	require.Equal(t, http.StatusPaymentRequired, resp.StatusCode)
	payment := decodeData[models.PaymentResponse](t, resp.Body)
	assert.Equal(t, models.Declined, payment.PaymentStatus)
	assert.NotEmpty(t, payment.ID)
	assert.Equal(t, "8872", payment.CardNumberLastFour)
}

func TestPostPayment_BankUnavailable(t *testing.T) {
	body := baseRequest()
	body["card_number"] = "2222405343248870" // ends in 0 → bank 503

	resp := postPayment(t, body)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestPostPayment_ValidationFailed_ExpiredCard(t *testing.T) {
	body := baseRequest()
	body["expiry_year"] = 2020

	resp := postPayment(t, body)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestPostPayment_ValidationFailed_MissingField(t *testing.T) {
	body := baseRequest()
	delete(body, "cvv")

	resp := postPayment(t, body)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestPostPayment_ValidationFailed_UnsupportedCurrency(t *testing.T) {
	body := baseRequest()
	body["currency"] = "JPY"

	resp := postPayment(t, body)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestGetPayment_Found(t *testing.T) {
	postResp := postPayment(t, baseRequest())
	defer postResp.Body.Close()
	require.Equal(t, http.StatusCreated, postResp.StatusCode)

	created := decodeData[models.PaymentResponse](t, postResp.Body)
	require.NotEmpty(t, created.ID)

	resp, err := http.Get(fmt.Sprintf("%s/api/payments/%s", gatewayURL, created.ID))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	payment := decodeData[models.PaymentResponse](t, resp.Body)
	assert.Equal(t, created.ID, payment.ID)
	assert.Equal(t, models.Authorized, payment.PaymentStatus)
	assert.Equal(t, "8871", payment.CardNumberLastFour)
}

func TestGetPayment_NotFound(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/api/payments/a0000000-0000-0000-0000-000000000000", gatewayURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetPayment_InvalidID(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/api/payments/not-a-uuid", gatewayURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
