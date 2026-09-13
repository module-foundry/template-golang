package validx

import (
	"errors"
	"testing"

	"template-golang/pkg/apperror"
)

type userDTO struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

func TestValidateStructSuccess(t *testing.T) {
	dto := userDTO{Email: "ada@example.com", Name: "ada"}
	if err := Default().ValidateStruct(dto); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateStructReturnsAPIError(t *testing.T) {
	dto := userDTO{Email: "not-an-email"}
	err := Default().ValidateStruct(dto)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want *apperror.Error, got %T", err)
	}
	if appErr.Code != apperror.CodeValidationFailed {
		t.Fatalf("code = %s", appErr.Code)
	}
	if appErr.Kind() != apperror.KindAPI {
		t.Fatalf("kind = %s", appErr.Kind())
	}
}

func TestValidateNil(t *testing.T) {
	if err := Default().ValidateStruct(nil); err != nil {
		t.Fatalf("nil must be valid, got %v", err)
	}
}
