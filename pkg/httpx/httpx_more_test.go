package httpx

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
)

func TestGetAndPatch(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	Get(reg, app, "/ping", func(context.Context, struct{}) (map[string]string, error) {
		return map[string]string{"status": "ok"}, nil
	})
	Patch(reg, app, "/items/:id", func(context.Context, createRequest) (createResponse, error) {
		return createResponse{ID: "patched"}, nil
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("get status = %d", resp.StatusCode)
	}

	req := httptest.NewRequest(fiber.MethodPatch, "/items/1", strings.NewReader(`{"name":"n"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp2, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp2.Body.Close() }()
	if resp2.StatusCode != 200 {
		t.Fatalf("patch status = %d", resp2.StatusCode)
	}
}

func TestPostMalformedJSON(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	registerCreate(t, app, reg)

	req := httptest.NewRequest(fiber.MethodPost, "/items", strings.NewReader(`{`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestHandlerErrorPropagates(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	Get(reg, app, "/missing", func(context.Context, struct{}) (createResponse, error) {
		return createResponse{}, apperror.API(apperror.CodeNotFound, "missing")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOperationIDAndTagOptions(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	Get(reg, app, "/custom", func(context.Context, struct{}) (createResponse, error) {
		return createResponse{}, nil
	}, OperationID("customop"), Tag("custom"))

	route := reg.Routes()[0]
	if route.OperationID != "customop" || route.Tag != "custom" {
		t.Fatalf("route = %+v", route)
	}
}

func TestDeriveTagFallback(t *testing.T) {
	if got := deriveTag("nodot"); got != "default" {
		t.Fatalf("tag = %q", got)
	}
}

func TestRegisteredMethods(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	Get(reg, app, "/a", func(context.Context, struct{}) (Empty, error) { return Empty{}, nil })
	Post(reg, app, "/a", func(context.Context, struct{}) (Empty, error) { return Empty{}, nil })
	Patch(reg, app, "/a", func(context.Context, struct{}) (Empty, error) { return Empty{}, nil })
	Delete(reg, app, "/a", func(context.Context, struct{}) (Empty, error) { return Empty{}, nil })
	if len(reg.Routes()) != 4 {
		t.Fatalf("routes = %d", len(reg.Routes()))
	}
}

func TestHandlerWithCtxError(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	PostWithCtx(reg, app, "/ctx", func(c fiber.Ctx, req createRequest) (createResponse, error) {
		_ = c
		_ = req
		return createResponse{}, apperror.API(apperror.CodeConflict, "conflict")
	})
	req := httptest.NewRequest(fiber.MethodPost, "/ctx", strings.NewReader(`{"name":"x"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 409 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

var _ = errors.Is
