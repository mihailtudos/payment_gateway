// Package models contains the data structures and validation logic for the Payment Gateway.
package models

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is a package-level singleton used to perform structural and business logic validation.
var validate = validator.New()

// init bootstraps the validation engine configuration.
func init() {
	// RegisterTagNameFunc instructs the validator to follow the JSON tag names instead of struct filed names.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}
