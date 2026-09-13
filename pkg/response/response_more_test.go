package response

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jsonx"
)

func TestFailPlainError(t *testing.T) {
	app := newApp()
	app.Get("/fail", func(c fiber.Ctx) error {
		return Fail(c, errors.New("db exploded"))
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/fail", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 500 {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var envelope apperror.Envelope
	if err := jsonx.Unmarshal(readAll(t, resp), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Type != apperror.CodeInternal {
		t.Fatalf("type = %s", envelope.Error.Type)
	}
	if envelope.Error.Message != "internal server error" {
		t.Fatalf("message leaked: %s", envelope.Error.Message)
	}
}
