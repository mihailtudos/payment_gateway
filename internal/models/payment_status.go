package models

import "encoding/json"

type PaymentStatus string

const (
	Authorized PaymentStatus = "Authorized"
	Declined   PaymentStatus = "Declined"
	Rejected   PaymentStatus = "Rejected"
	NoStatus   PaymentStatus = ""
)

func (s PaymentStatus) String() string {
	return string(s)
}

func (s PaymentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *PaymentStatus) UnmarshalJSON(data []byte) error {
	var displayString string
	if err := json.Unmarshal(data, &displayString); err != nil {
		return err
	}

	switch displayString {
	case "Authorized":
		*s = Authorized
	case "Declined":
		*s = Declined
	case "Rejected":
		*s = Rejected
	default:
		*s = ""
	}
	return nil
}
