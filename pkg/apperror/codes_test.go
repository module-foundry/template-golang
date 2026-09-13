package apperror

import (
	"os"
	"regexp"
	"testing"
)

// TestRegistryIsValid guards the public contract of error codes.
func TestRegistryIsValid(t *testing.T) {
	codes := Codes()
	if len(codes) == 0 {
		t.Fatal("registry is empty")
	}
	seenStatus := map[int]bool{}
	for code, info := range codes {
		if code == "" {
			t.Error("empty code registered")
		}
		if info.PublicMessage == "" {
			t.Errorf("%s: public message is empty", code)
		}
		if info.Kind != KindAPI && info.Kind != KindRuntime {
			t.Errorf("%s: unknown kind %q", code, info.Kind)
		}
		if info.HTTPStatus < 400 || info.HTTPStatus > 599 {
			t.Errorf("%s: status %d out of range", code, info.HTTPStatus)
		}
		if info.Kind == KindAPI && info.HTTPStatus >= 500 {
			t.Errorf("%s: API code must be 4xx, got %d", code, info.HTTPStatus)
		}
		if info.Kind == KindRuntime && info.HTTPStatus < 500 {
			t.Errorf("%s: Runtime code must be 5xx, got %d", code, info.HTTPStatus)
		}
		_ = seenStatus
	}
}

// TestCodesFileHasNoStatusLiterals enforces "HTTP statuses come from constants".
func TestCodesFileHasNoStatusLiterals(t *testing.T) {
	source, err := os.ReadFile("codes.go")
	if err != nil {
		t.Fatalf("read codes.go: %v", err)
	}
	statusLiteral := regexp.MustCompile(`\b[1-5][0-9]{2}\b`)
	if match := statusLiteral.Find(source); match != nil {
		t.Fatalf("codes.go contains numeric HTTP status literal %q; use fiber.Status* constants", match)
	}
}

func TestInfoOfFallback(t *testing.T) {
	info := InfoOf(Code("UNKNOWN_CODE"))
	if info.Kind != KindRuntime || info.HTTPStatus != InfoOf(CodeInternal).HTTPStatus {
		t.Fatalf("unknown code must fall back to INTERNAL, got %+v", info)
	}
}
