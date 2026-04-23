package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardNumber_LastFour(t *testing.T) {
	tests := []struct {
		card CardNumber
		want string
	}{
		{"2222405343248871", "8871"},
		{"12345678901234", "1234"},   // 14-digit minimum
		{"1234567890123456789", "6789"}, // 19-digit maximum
		{"1234", "1234"},             // exactly 4 digits
		{"123", "123"},              // shorter than 4 — return as-is
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.card.LastFour(), "card=%s", tt.card)
	}
}

func TestCardNumber_Masked(t *testing.T) {
	card := CardNumber("2222405343248871")

	assert.Equal(t, "************8871", card.Masked(4))
	assert.Equal(t, "**********248871", card.Masked(6))
	assert.Equal(t, "2222405343248871", card.Masked(20)) // visible >= len: return unchanged
}

func TestCardNumber_Validate(t *testing.T) {
	tests := []struct {
		name    string
		card    CardNumber
		wantErr bool
	}{
		{"valid 16-digit", "2222405343248871", false},
		{"valid 14-digit min", "22224053432488", false},
		{"valid 19-digit max", "2222405343248871234", false},
		{"too short 13-digit", "2222405343248", true},
		{"too long 20-digit", "22224053432488712345", true},
		{"non-numeric", "222240534324ABCD", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardNumber_MarshalJSON(t *testing.T) {
	card := CardNumber("2222405343248871")
	b, err := json.Marshal(card)
	require.NoError(t, err)
	assert.Equal(t, `"2222405343248871"`, string(b))
}

func TestCardNumber_UnmarshalJSON(t *testing.T) {
	var card CardNumber
	err := json.Unmarshal([]byte(`"2222405343248871"`), &card)
	require.NoError(t, err)
	assert.Equal(t, CardNumber("2222405343248871"), card)
}

func TestCardNumber_UnmarshalJSON_Invalid(t *testing.T) {
	var card CardNumber
	err := json.Unmarshal([]byte(`12345`), &card)
	assert.Error(t, err)
}
