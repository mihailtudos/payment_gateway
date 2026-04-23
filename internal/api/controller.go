package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/handlers"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @Summary Health check endpoint Ping-Pong
// @Produce json
// @Success 200 {object} models.Pong
// @Router /ping [get]
// PingHandler returns an http.HandlerFunc that handles HTTP Ping GET requests.
func (a *API) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(models.Pong{Message: "pong"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// SwaggerHandler returns an http.HandlerFunc that handles HTTP Swagger related requests.
func (a *API) SwaggerHandler() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", docs.SwaggerInfo.Host)),
	)
}

// @Summary      Retrieve a payment by ID
// @Description  Returns the stored details of a previously processed payment (authorized or declined).
// @Produce      json
// @Param        id   path      string                true  "Payment UUID"  format(uuid)
// @Success      200  {object}  PaymentResponseEnvelope          "Payment found"
// @Failure      404  {object}  ErrorResponseEnvelope            "Payment not found"
// @Failure      400  {object}  ErrorResponseEnvelope            "Invalid UUID"
// @Failure      500  {object}  ErrorResponseEnvelope            "Internal server error"
// @Router       /api/payments/{id} [get]
// GetPaymentHandler returns an http.HandlerFunc that handles Payments GET requests.
func (a *API) GetPaymentHandler() http.HandlerFunc {
	h := handlers.NewPaymentsHandler(a.paymentsRepo, nil, a.telemetry)

	return h.GetHandler()
}

// @Summary      Process a payment
// @Description  Validates the request, forwards to the acquiring bank, stores the result, and returns it.
// @Description  Three outcomes are possible:
// @Description  - **Authorized (201)** — bank approved; payment stored and returned.
// @Description  - **Declined (402)**   — bank declined; payment still stored and returned with decline error.
// @Description  - **Rejected (422)**   — validation failed; bank never called, nothing stored.
// @Accept       json
// @Produce      json
// @Param        payment  body      models.PostPaymentRequest           true   "Payment request body"
// @Success      201      {object}  PaymentResponseEnvelope             "Payment authorized"
// @Failure      400      {object}  ErrorResponseEnvelope               "Malformed request body"
// @Failure      402      {object}  DeclinedPaymentResponseEnvelope     "Payment declined — record saved, ID returned"
// @Failure      422      {object}  ErrorResponseEnvelope               "Validation failed"
// @Failure      500      {object}  ErrorResponseEnvelope               "Internal server error"
// @Failure      503      {object}  ErrorResponseEnvelope               "Acquiring bank unavailable"
// @Router       /api/payments [post]
// PostPaymentHandler returns an http.HandlerFunc that handles Payments POST requests.
func (a *API) PostPaymentHandler() http.HandlerFunc {
	h := handlers.NewPaymentsHandler(a.paymentsRepo, a.acquiringBank, a.telemetry)

	return h.PostHandler()
}
