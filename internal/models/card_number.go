package models

import (
	"encoding/json"
	"strings"
)

// CardNumber is a custom type based on string.
type CardNumber string

func (c CardNumber) String() string {
	return string(c)
}

func (c CardNumber) Validate() error {
	return validate.Var(c, "required,numeric,min=14,max=19")
}

// LastFour returns the last 4 digits of the card.
func (c CardNumber) LastFour() string {
	if len(c) < 4 {
		return string(c)
	}
	return string(c[len(c)-4:])
}

// Masked returns a masked version for logs, responses, UI — explicit opt-in.
func (c CardNumber) Masked(visibleCount int) string {
	if len(c) <= visibleCount {
		return string(c)
	}
	maskLen := len(c) - visibleCount
	return strings.Repeat("*", maskLen) + string(c[len(c)-visibleCount:])
}

// MarshalJSON marshals the raw value — needed for bank transmission in request bodies.
func (c CardNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

// UnmarshalJSON unmarshals from the incoming request body.
func (c *CardNumber) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*c = CardNumber(s)
	return nil
}
