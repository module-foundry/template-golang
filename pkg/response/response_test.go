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

// TestEnvelopeGolden locks the success shape to exactly one "result" key.
func TestEnvelopeGolden(t *testing.T) {
	got, err := jsonx.MarshalDeterministic(Envelope{Result: map[string]string{"status": "ok"}})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"result":{"status":"ok"}}`
	if string(got) != want {
		t.Fatalf("envelope = %s, want %s", got, want)
	}
}

// TestPaginationShape locks the list convention: result.pagination.<params>.
func TestPaginationShape(t *testing.T) {
	type item struct {
		ID string `json:"id"`
	}
	type listResult struct {
		Items      []item     `json:"items"`
		Pagination Pagination `json:"pagination"`
	}
	got, err := jsonx.MarshalDeterministic(Envelope{Result: listResult{
		Items:      []item{{ID: "1"}},
		Pagination: Pagination{Page: 2, PerPage: 20, Total: 57},
	}})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"result":{"items":[{"id":"1"}],"pagination":{"page":2,"per_page":20,"total":57}}}`
	if string(got) != want {
		t.Fatalf("list envelope = %s, want %s", got, want)
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
