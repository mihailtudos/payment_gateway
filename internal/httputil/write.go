package httputil

import (
	"encoding/json"
	"net/http"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload) //nolint:errcheck
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// OK writes a 200 response with data.
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Envelope{
		Success:   true,
		Data:      data,
		Timestamp: now(),
	})
}

// Created writes a 201 response with data.
func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, Envelope{
		Success:   true,
		Data:      data,
		Timestamp: now(),
	})
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 with a message.
func BadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "BAD_REQUEST",
			Message: message,
		},
		Timestamp: now(),
	})
}

// ValidationFailed writes a 422 with per-field validation details.
func ValidationFailed(w http.ResponseWriter, details []ValidationDetail) {
	writeJSON(w, http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_FAILED",
			Message: "one or more fields failed validation",
			Details: details,
		},
		Timestamp: now(),
	})
}

// NotFound writes a 404.
func NotFound(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusNotFound, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "NOT_FOUND",
			Message: message,
		},
		Timestamp: now(),
	})
}

// Unauthorized writes a 401.
func Unauthorized(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
		Timestamp: now(),
	})
}

// Forbidden writes a 403.
func Forbidden(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusForbidden, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "FORBIDDEN",
			Message: message,
		},
		Timestamp: now(),
	})
}

// InternalServerError writes a 500, masking internal details from the client.
func InternalServerError(w http.ResponseWriter) {
	writeJSON(w, http.StatusInternalServerError, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "an unexpected error occurred, please try again later",
		},
		Timestamp: now(),
	})
}

func ServiceUnavailable(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusServiceUnavailable, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "SERVICE_UNAVAILABLE",
			Message: message,
		},
		Timestamp: now(),
	})
}

// PaymentDeclined writes a 402 for downstream bank declines.
// The declined payment record is included in data so the caller has the ID and status.
func PaymentDeclined(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusPaymentRequired, Envelope{
		Success: false,
		Data:    data,
		Error: &APIError{
			Code:    "PAYMENT_DECLINED",
			Message: "bank declined payment",
		},
		Timestamp: now(),
	})
}
