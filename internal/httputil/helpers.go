package httputil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// TranslateValidationErrors converts validator.ValidationErrors into
// a slice of ValidationDetail ready to pass straight into ValidationFailed.
func TranslateValidationErrors(err error) []ValidationDetail {
	var fe *FieldError
	if errors.As(err, &fe) {
		return []ValidationDetail{
			{
				Field:   fe.Field,
				Message: fmt.Sprintf("%s: %s", fe.Field, fe.Message),
			},
		}
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []ValidationDetail{
			{Field: "request", Message: err.Error()},
		}
	}
	details := make([]ValidationDetail, 0, len(ve))
	for _, fe := range ve {
		details = append(details, ValidationDetail{
			Field:   fe.Field(),
			Message: humanReadableMessage(fe),
		})
	}
	return details
}

func humanReadableMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return field + " is invalid"
	}
}
