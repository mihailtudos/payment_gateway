package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank/mocks"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/pkg/tel"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// envelope mirrors httputil.Envelope for decoding test responses.
type envelope[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

func decodeData[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var env envelope[T]
	require.NoError(t, json.NewDecoder(w.Body).Decode(&env))
	return env.Data
}

// newRouter wires a PaymentsHandler onto a chi router for use in table tests.
func newRouter(ps *repository.PaymentsRepository, bankAdapter bank.Adapter) *chi.Mux {
	h := NewPaymentsHandler(ps, bankAdapter, tel.NewNoopTelemetry())
	r := chi.NewRouter()
	r.Get("/api/payments/{id}", h.GetHandler())
	r.Post("/api/payments", h.PostHandler())
	return r
}

func serve(r *chi.Mux, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// validPostBody returns a JSON-encoded POST body whose card ends in the given digit.
func validPostBody(t *testing.T, lastDigit int) *bytes.Buffer {
	t.Helper()
	// Base card is 15 digits; append lastDigit to produce a 16-digit number.
	body := map[string]any{
		"card_number":  "222240534324887" + string(rune('0'+lastDigit)),
		"expiry_month": 12,
		"expiry_year":  2030,
		"currency":     "GBP",
		"amount":       100,
		"cvv":          "123",
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

// ── GET /api/payments/{id} ────────────────────────────────────────────────────

func TestGetPaymentHandler_Found(t *testing.T) {
	stored := models.PaymentResponse{
		ID:                 "a0000000-0000-0000-0000-000000000001",
		PaymentStatus:      models.Authorized,
		CardNumberLastFour: "8871",
		ExpiryMonth:        12,
		ExpiryYear:         2030,
		Currency:           "GBP",
		Amount:             100,
	}
	ps := repository.NewPaymentsRepository()
	assert.NoError(t, ps.AddPayment(stored))

	r := newRouter(ps, nil)
	req, _ := http.NewRequest(http.MethodGet, "/api/payments/a0000000-0000-0000-0000-000000000001", nil)
	w := serve(r, req)

	assert.Equal(t, http.StatusOK, w.Code)
	got := decodeData[models.PaymentResponse](t, w)
	assert.Equal(t, stored.ID, got.ID)
	assert.Equal(t, models.Authorized, got.PaymentStatus)
}

func TestGetPaymentHandler_NotFound(t *testing.T) {
	r := newRouter(repository.NewPaymentsRepository(), nil)
	req, _ := http.NewRequest(http.MethodGet, "/api/payments/b0000000-0000-0000-0000-000000000002", nil)
	w := serve(r, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPaymentHandler_InvalidID(t *testing.T) {
	r := newRouter(repository.NewPaymentsRepository(), nil)
	req, _ := http.NewRequest(http.MethodGet, "/api/payments/not-a-uuid", nil)
	w := serve(r, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPaymentHandler_NilUUID(t *testing.T) {
	r := newRouter(repository.NewPaymentsRepository(), nil)
	req, _ := http.NewRequest(http.MethodGet, "/api/payments/00000000-0000-0000-0000-000000000000", nil)
	w := serve(r, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── POST /api/payments ────────────────────────────────────────────────────────

func TestPostPaymentHandler_Authorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBank := mocks.NewMockAdapter(ctrl)
	mockBank.EXPECT().
		ProcessPayment(gomock.Any(), gomock.Any()).
		Return(models.BankResponse{Authorized: true, AuthorizationCode: "auth-code-123"}, nil)

	r := newRouter(repository.NewPaymentsRepository(), mockBank)
	req, _ := http.NewRequest(http.MethodPost, "/api/payments", validPostBody(t, 1))
	req.Header.Set("Content-Type", "application/json")
	w := serve(r, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	got := decodeData[models.PaymentResponse](t, w)
	assert.Equal(t, models.Authorized, got.PaymentStatus)
	assert.NotEmpty(t, got.ID)
	assert.Equal(t, "GBP", got.Currency)
	assert.Equal(t, 100, got.Amount)
}

func TestPostPaymentHandler_Declined(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBank := mocks.NewMockAdapter(ctrl)
	mockBank.EXPECT().
		ProcessPayment(gomock.Any(), gomock.Any()).
		Return(models.BankResponse{Authorized: false, AuthorizationCode: ""}, nil)

	r := newRouter(repository.NewPaymentsRepository(), mockBank)
	req, _ := http.NewRequest(http.MethodPost, "/api/payments", validPostBody(t, 2))
	req.Header.Set("Content-Type", "application/json")
	w := serve(r, req)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	got := decodeData[models.PaymentResponse](t, w)
	assert.Equal(t, models.Declined, got.PaymentStatus)
	assert.NotEmpty(t, got.ID)
}

func TestPostPaymentHandler_Rejected_ValidationFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBank := mocks.NewMockAdapter(ctrl)
	// Bank must NOT be called when validation fails.
	mockBank.EXPECT().ProcessPayment(gomock.Any(), gomock.Any()).Times(0)

	r := newRouter(repository.NewPaymentsRepository(), mockBank)

	body, _ := json.Marshal(map[string]any{
		"card_number":  "2222405343248871",
		"expiry_month": 12,
		"expiry_year":  2020, // expired
		"currency":     "GBP",
		"amount":       100,
		"cvv":          "123",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := serve(r, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestPostPaymentHandler_BankUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBank := mocks.NewMockAdapter(ctrl)
	mockBank.EXPECT().
		ProcessPayment(gomock.Any(), gomock.Any()).
		Return(models.BankResponse{}, bank.ErrBankUnavailable)

	r := newRouter(repository.NewPaymentsRepository(), mockBank)
	req, _ := http.NewRequest(http.MethodPost, "/api/payments", validPostBody(t, 1))
	req.Header.Set("Content-Type", "application/json")
	w := serve(r, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestPostPaymentHandler_BankBadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBank := mocks.NewMockAdapter(ctrl)
	mockBank.EXPECT().
		ProcessPayment(gomock.Any(), gomock.Any()).
		Return(models.BankResponse{}, errors.New("unexpected bank error"))

	r := newRouter(repository.NewPaymentsRepository(), mockBank)
	req, _ := http.NewRequest(http.MethodPost, "/api/payments", validPostBody(t, 1))
	req.Header.Set("Content-Type", "application/json")
	w := serve(r, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
