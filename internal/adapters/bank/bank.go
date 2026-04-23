package bank

import (
	"context"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

type Adapter interface {
	ProcessPayment(ctx context.Context, req models.BankRequest) (models.BankResponse, error)
}
