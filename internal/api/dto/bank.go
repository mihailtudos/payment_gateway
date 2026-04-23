package dto

import (
	"fmt"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/google/uuid"
)

func FromPaymentRequestToBankReq(payment models.PostPaymentRequest) models.BankRequest {
	return models.BankRequest{
		CardNumber: payment.CardNumber,
		ExpiryDate: fmt.Sprintf("%d/%d", payment.ExpiryMonth, payment.ExpiryYear),
		Currency:   payment.Currency,
		Amount:     payment.Amount,
		Cvv:        payment.Cvv,
	}
}

func FromBankResToPaymentResp(payment models.PostPaymentRequest, res models.BankResponse) models.PaymentResponse {
	var status models.PaymentStatus
	if res.Authorized {
		status = models.Authorized
	} else {
		status = models.Declined
	}

	return models.PaymentResponse{
		ID:                 uuid.NewString(),
		PaymentStatus:      status,
		CardNumberLastFour: payment.CardNumber.LastFour(),
		ExpiryMonth:        payment.ExpiryMonth,
		ExpiryYear:         payment.ExpiryYear,
		Currency:           payment.Currency,
		Amount:             payment.Amount,
	}
}
