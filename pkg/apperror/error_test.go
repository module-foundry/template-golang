package apperror

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/jsonx"
)

func TestAPIErrorShape(t *testing.T) {
	Configure(true, false)
	err := API(CodeNotFound, "user not found")
	if err.Code != CodeNotFound {
		t.Fatalf("code = %s", err.Code)
	}
	if err.HTTPStatus() != 404 {
		t.Fatalf("status = %d", err.HTTPStatus())
	}
	if err.Kind() != KindAPI {
		t.Fatalf("kind = %s", err.Kind())
	}
	if err.Trace() == "" {
		t.Fatal("trace must be captured when enabled")
	}
	if err.PublicMessage() != "user not found" {
		t.Fatalf("public message = %q", err.PublicMessage())
	}
}

func TestRuntimeErrorHidesDetails(t *testing.T) {
	Configure(false, false)
	cause := errors.New("pq: relation does not exist")
	err := Runtime(CodeDBQueryFailed, cause)
	if err.Trace() != "" {
		t.Fatal("trace must be empty when disabled")
	}
	if err.PublicMessage() != "internal server error" {
		t.Fatalf("public message leaked: %q", err.PublicMessage())
	}
	if !errors.Is(err, cause) {
		t.Fatal("cause must be unwrappable")
	}
	var target *Error
	if !errors.As(err, &target) || target.Code != CodeDBQueryFailed {
		t.Fatal("errors.As must find *Error")
	}
}

func TestMessageTruncatedToLimit(t *testing.T) {
	long := strings.Repeat("я", MaxMessageLength+50)
	err := API(CodeValidationFailed, long)
	if got := len([]rune(err.Message)); got != MaxMessageLength {
		t.Fatalf("message length = %d, want %d", got, MaxMessageLength)
	}
}

func TestHandlerWritesEnvelope(t *testing.T) {
	log := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	app := fiber.New(fiber.Config{ErrorHandler: Handler(log)})
	app.Get("/boom", func(fiber.Ctx) error {
		return API(CodeValidationFailed, "bad field")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}

	var envelope Envelope
	if err := jsonx.Unmarshal(readBody(t, resp), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error.Type != CodeValidationFailed {
		t.Fatalf("type = %s", envelope.Error.Type)
	}
	if envelope.Error.Message != "bad field" {
		t.Fatalf("message = %s", envelope.Error.Message)
	}
}

func TestHandlerWrapsPlainErrors(t *testing.T) {
	log := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	app := fiber.New(fiber.Config{ErrorHandler: Handler(log)})
	app.Get("/boom", func(fiber.Ctx) error {
		return fmt.Errorf("unexpected: %w", errors.New("dial tcp"))
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 500 {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func BenchmarkErrorAPI(b *testing.B) {
	Configure(true, false)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = API(CodeValidationFailed, "bad field")
	}
}

func BenchmarkErrorRuntime(b *testing.B) {
	Configure(false, false)
	cause := errors.New("db is down")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Runtime(CodeDBQueryFailed, cause)
	}
}
