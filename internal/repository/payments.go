package repository

import (
	"fmt"
	"sync"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/google/uuid"
)

type SafeMap struct {
	mu       *sync.RWMutex
	payments map[uuid.UUID]models.PaymentResponse
}

type PaymentsRepository struct {
	SafeMap
}

func NewPaymentsRepository() *PaymentsRepository {
	store := &PaymentsRepository{}
	store.payments = map[uuid.UUID]models.PaymentResponse{}
	store.mu = &sync.RWMutex{}
	return store
}

func (ps *PaymentsRepository) GetPayment(id uuid.UUID) *models.PaymentResponse {
	ps.mu.RLock()
	p, ok := ps.payments[id]
	ps.mu.RUnlock()

	if !ok {
		return nil
	}

	return &p
}

func (ps *PaymentsRepository) AddPayment(payment models.PaymentResponse) error {
	pID, err := uuid.Parse(payment.ID)
	if err != nil {
		return fmt.Errorf("repository: invalid payment id %q: %w", payment.ID, err)
	}

	ps.mu.Lock()
	ps.payments[pID] = payment
	ps.mu.Unlock()

	return nil
}
