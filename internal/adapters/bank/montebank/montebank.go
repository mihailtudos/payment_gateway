package montebank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/httputil"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

type Config struct {
	BaseURL         string
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

func DefaultConfig(baseURL string) Config {
	return Config{
		BaseURL:         baseURL,
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}
}

type HTTPBankAdapter struct {
	client  *http.Client
	baseURL string
}

func NewHTTPBankAdapter(cfg Config) *HTTPBankAdapter {
	transport := &http.Transport{
		MaxIdleConns:       cfg.MaxIdleConns,
		IdleConnTimeout:    cfg.IdleConnTimeout,
		DisableCompression: false,
	}

	return &HTTPBankAdapter{
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		baseURL: cfg.BaseURL,
	}
}

func (a *HTTPBankAdapter) ProcessPayment(ctx context.Context, req models.BankRequest) (models.BankResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return models.BankResponse{}, fmt.Errorf("bank adapter: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return models.BankResponse{}, fmt.Errorf("bank adapter: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return models.BankResponse{}, fmt.Errorf("bank adapter: call bank: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var bankResp models.BankResponse
		if err := httputil.DecodeJSON(resp.Body, &bankResp); err != nil {
			return models.BankResponse{}, fmt.Errorf("bank adapter: decode response: %w", err)
		}
		return bankResp, nil

	case http.StatusServiceUnavailable:
		return models.BankResponse{}, bank.ErrBankUnavailable

	case http.StatusBadRequest:
		return models.BankResponse{}, bank.ErrBankBadRequest

	default:
		return models.BankResponse{}, fmt.Errorf("bank adapter: unexpected status %d", resp.StatusCode)
	}
}
