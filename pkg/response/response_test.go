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
	var body Envelope
	if err := jsonx.Unmarshal(readAll(t, resp), &body); err != nil {
		t.Fatal(err)
	}
	result, ok := body.Result.(map[string]any)
	if !ok || result["status"] != "ok" {
		t.Fatalf("payload not wrapped in result: %s", body.Result)
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
	var body Envelope
	if err := jsonx.Unmarshal(readAll(t, resp), &body); err != nil {
		t.Fatal(err)
	}
	if result, ok := body.Result.(map[string]any); !ok || result["id"] != "1" {
		t.Fatalf("payload not wrapped in result: %s", body.Result)
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
