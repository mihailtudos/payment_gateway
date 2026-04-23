package models

import (
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/httputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	now := time.Now()

	expired := now.AddDate(0, -1, 0)
	expiredMonth := int(expired.Month())
	expiredYear := expired.Year()

	nextMonth := now.AddDate(0, 1, 0)
	year := nextMonth.Year()
	month := int(nextMonth.Month())

	testCases := []struct {
		name           string
		input          PostPaymentRequest
		expectError    bool
		expectedErrors []httputil.ValidationDetail
	}{
		{
			name: "[success] - valid payment request",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
		},
		{
			name: "[invalid] - required card number",
			input: PostPaymentRequest{
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "card_number",
					Message: "card_number is required",
				},
			},
		},
		{
			name: "[invalid] - required ExpiryMonth",
			input: PostPaymentRequest{
				CardNumber: "2222405343248871",
				ExpiryYear: year,
				Currency:   "USD",
				Amount:     100,
				Cvv:        "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "expiry_month",
					Message: "expiry_month is required",
				},
			},
		},
		{
			name: "[invalid] - required ExpiryYear",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "expiry_year",
					Message: "expiry_year is required",
				},
			},
		},
		{
			name: "[invalid] - required Currency",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "currency",
					Message: "currency is required",
				},
			},
		},
		{
			name: "[invalid] - required amount",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "USD",
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "amount",
					Message: "amount is required",
				},
			},
		},
		{
			name: "[invalid] - required Cvv",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "cvv",
					Message: "cvv is required",
				},
			},
		},
		{
			name: "[invalid] - value too long CardNumber",
			input: PostPaymentRequest{
				CardNumber:  "22224053432488711234",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "card_number",
					Message: "card_number must be at most 19",
				},
			},
		},
		{
			name: "[invalid] - value too short CardNumber",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "card_number",
					Message: "card_number must be at least 14",
				},
			},
		},
		{
			name: "[invalid] - ExpiryMonth must be at most 12",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: 13,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "expiry_month",
					Message: "expiry_month must be at most 12",
				},
			},
		},
		{
			name: "[invalid] - ExpiryMonth must be at least 1",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: -1,
				ExpiryYear:  year,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "expiry_month",
					Message: "expiry_month must be at least 1",
				},
			},
		},
		{
			name: "[invalid] - Card is expired due to year",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  expiredYear - 1,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "expiry_year",
					Message: "expiry_year: card is expired",
				},
			},
		},
		{
			name: "[invalid] - Card is expired due to month",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: expiredMonth,
				ExpiryYear:  expiredYear,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "expiry_month",
					Message: "expiry_month: card is expired",
				},
			},
		},
		{
			name: "[invalid] - Currency not in the enum",
			input: PostPaymentRequest{
				CardNumber:  "2222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "RUB",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "currency",
					Message: "currency must be one of: USD, GBP, EUR",
				},
			},
		},
		{
			name: "[invalid] - Multiple validation errors",
			input: PostPaymentRequest{
				CardNumber:  "222405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "RUB",
				Amount:      0,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "currency",
					Message: "currency must be one of: USD, GBP, EUR",
				},
				{
					Field:   "amount",
					Message: "amount is required",
				},
			},
		},
		{
			name: "[invalid] - Card number must be only digits",
			input: PostPaymentRequest{
				CardNumber:  "2a22405343248871",
				ExpiryMonth: month,
				ExpiryYear:  year,
				Currency:    "GBP",
				Amount:      100,
				Cvv:         "123",
			},
			expectError: true,
			expectedErrors: []httputil.ValidationDetail{
				{
					Field:   "card_number",
					Message: "card_number is invalid",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if !tc.expectError {
				require.NoError(t, err)
				return
			}

			assert.Error(t, err)
			humanReadableErr := httputil.TranslateValidationErrors(err)
			assert.ElementsMatch(t, tc.expectedErrors, humanReadableErr)
		})
	}
}
