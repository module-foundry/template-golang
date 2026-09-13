package validx

import (
	"errors"
	"strings"
	"testing"

	"template-golang/pkg/apperror"
)

func TestHumanizeBranches(t *testing.T) {
	cases := []struct {
		name string
		dto  any
		want string
	}{
		{"required", struct {
			Name string `json:"name" validate:"required"`
		}{}, "name is required"},
		{"uuid", struct {
			ID string `json:"id" validate:"uuid"`
		}{ID: "nope"}, "id must be a valid uuid"},
		{"min", struct {
			Name string `json:"name" validate:"min=5"`
		}{Name: "ab"}, "name is too short"},
		{"max", struct {
			Name string `json:"name" validate:"max=2"`
		}{Name: "abcd"}, "name is too long"},
		{"default", struct {
			Count int `json:"count" validate:"gt=10"`
		}{Count: 1}, "count is invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Default().ValidateStruct(tc.dto)
			if err == nil {
				t.Fatal("expected error")
			}
			var appErr *apperror.Error
			if !errors.As(err, &appErr) {
				t.Fatalf("want *apperror.Error, got %T", err)
			}
			if appErr.Message != tc.want {
				t.Fatalf("message = %q, want %q", appErr.Message, tc.want)
			}
		})
	}
}

func TestValidateBrokenStruct(t *testing.T) {
	err := Default().ValidateStruct(struct{}{})
	if err != nil {
		t.Fatalf("empty struct must validate: %v", err)
	}
}

func TestHumanizeEmpty(t *testing.T) {
	if got := humanize(nil); got != "invalid request" {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(humanize(nil), "invalid") {
		t.Fatal("sanity")
	}
}
