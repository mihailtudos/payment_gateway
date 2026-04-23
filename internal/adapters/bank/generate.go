package bank

//go:generate go run go.uber.org/mock/mockgen -destination mocks/mock_adapter.go -package mocks github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank Adapter
