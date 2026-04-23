package httputil

import "fmt"

// FieldError represents a single business rule violation tied to a specific field.
type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
