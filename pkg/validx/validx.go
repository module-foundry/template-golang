// Package validx provides the shared validator and Fiber binding integration.
package validx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"template-golang/pkg/apperror"
)

// Validator wraps go-playground/validator with project defaults.
type Validator struct {
	validate *validator.Validate
}

var defaultValidator = New()

// Default returns the shared validator instance.
func Default() *Validator { return defaultValidator }

// New builds a Validator with JSON field names in error messages.
func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return &Validator{validate: v}
}

// ValidateStruct validates a DTO and converts failures into API errors.
func (v *Validator) ValidateStruct(out any) error {
	if out == nil {
		return nil
	}
	if err := v.validate.Struct(out); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return apperror.API(apperror.CodeValidationFailed, humanize(validationErrors))
		}
		return apperror.API(apperror.CodeValidationFailed, "invalid request")
	}
	return nil
}

// Validate implements fiber's StructValidator interface.
func (v *Validator) Validate(out any) error {
	return v.ValidateStruct(out)
}

func humanize(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "invalid request"
	}
	first := errs[0]
	field := first.Field()
	switch first.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "uuid":
		return field + " must be a valid uuid"
	case "min":
		return field + " is too short"
	case "max":
		return field + " is too long"
	default:
		return field + " is invalid"
	}
}
