package bank

import "errors"

var (
	// ErrBankUnavailable is returned when the bank returns 503.
	// The payment should be rejected — no authorization decision was made.
	ErrBankUnavailable = errors.New("bank adapter: acquiring bank unavailable")

	// ErrBankBadRequest means our gateway sent a malformed request — this is our bug.
	ErrBankBadRequest = errors.New("bank adapter: malformed request sent to bank")
)
