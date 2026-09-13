package response

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jsonx"
)

func newApp() *fiber.App {
	return fiber.New(fiber.Config{
		JSONEncoder: jsonx.FiberEncoder,
		JSONDecoder: jsonx.FiberDecoder,
	})
}

func TestOK(t *testing.T) {
	app := newApp()
	app.Get("/ok", func(c fiber.Ctx) error {
		return OK(c, map[string]string{"status": "ok"})
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ok", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestCreated(t *testing.T) {
	app := newApp()
	app.Post("/items", func(c fiber.Ctx) error {
		return Created(c, map[string]string{"id": "1"})
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/items", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 201 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestNoContent(t *testing.T) {
	app := newApp()
	app.Delete("/items", func(c fiber.Ctx) error {
		return NoContent(c)
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodDelete, "/items", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 204 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestFailUsesRegistry(t *testing.T) {
	app := newApp()
	app.Get("/fail", func(c fiber.Ctx) error {
		return Fail(c, apperror.API(apperror.CodeNotFound, "missing"))
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/fail", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}
