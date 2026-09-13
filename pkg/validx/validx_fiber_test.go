package validx

import "testing"

type fiberBoundDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// TestValidateFiberInterface ensures validx satisfies fiber's StructValidator.
func TestValidateFiberInterface(t *testing.T) {
	v := Default()
	if err := v.Validate(fiberBoundDTO{Email: "a@b.c"}); err != nil {
		t.Fatalf("valid dto rejected: %v", err)
	}
	if err := v.Validate(fiberBoundDTO{}); err == nil {
		t.Fatal("invalid dto accepted")
	}
}

// TestValidateNewInstance checks that New() returns an independent validator.
func TestValidateNewInstance(t *testing.T) {
	v := New()
	if err := v.ValidateStruct(fiberBoundDTO{Email: "a@b.c"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
