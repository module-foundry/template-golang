package apperror

import (
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestWrapNilCause(t *testing.T) {
	err := Wrap(nil, CodeInternal)
	if err.Code != CodeInternal {
		t.Fatalf("code = %s", err.Code)
	}
	if err.Unwrap() != nil {
		t.Fatal("nil cause must stay nil")
	}
}

func TestWrapWithCause(t *testing.T) {
	cause := errors.New("dial tcp: refused")
	err := Wrap(cause, CodeDBConnectFailed)
	if !errors.Is(err, cause) {
		t.Fatal("cause must be unwrappable")
	}
	if err.Error() == "" {
		t.Fatal("Error() must not be empty")
	}
}

func TestPublicMessageFallsBackToRegistry(t *testing.T) {
	err := API(CodeForbidden, "")
	if err.PublicMessage() != "forbidden" {
		t.Fatalf("public message = %q", err.PublicMessage())
	}
}

func TestIsComparesCodes(t *testing.T) {
	err := API(CodeNotFound, "x")
	if !errors.Is(err, API(CodeNotFound, "other")) {
		t.Fatal("errors.Is must compare codes")
	}
	if errors.Is(err, API(CodeConflict, "x")) {
		t.Fatal("different codes must not match")
	}
}

func TestHandlerRuntimeLogsStack(t *testing.T) {
	Configure(true, true)
	t.Cleanup(func() { Configure(true, false) })

	var buf strings.Builder
	log := slog.New(slog.NewTextHandler(&buf, nil))
	app := fiber.New(fiber.Config{ErrorHandler: Handler(log)})
	app.Get("/db", func(fiber.Ctx) error {
		return Runtime(CodeDBQueryFailed, errors.New("relation does not exist"))
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/db", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 500 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	line := buf.String()
	if !strings.Contains(line, "DB_QUERY_FAILED") || !strings.Contains(line, "stack=") {
		t.Fatalf("runtime log must contain code and stack: %s", line)
	}
}

func TestHandlerNilLogger(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: Handler(nil)})
	app.Get("/x", func(fiber.Ctx) error { return API(CodeNotFound, "nope") })
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/x", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}
