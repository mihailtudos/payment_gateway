package httputil

// Envelope is the standard API response wrapper for all responses.
// @Description Standard JSON envelope returned by all endpoints.
type Envelope struct {
	Success   bool      `json:"success"`
	Data      any       `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// APIError describes an error returned in the envelope.
type APIError struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
}

// ValidationDetail describes a single field-level failure.
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
