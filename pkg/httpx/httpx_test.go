package httpx

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jsonx"
)

type createRequest struct {
	Name string `json:"name" validate:"required"`
}

type createResponse struct {
	ID string `json:"id"`
}

func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		JSONEncoder:  jsonx.FiberEncoder,
		JSONDecoder:  jsonx.FiberDecoder,
		ErrorHandler: apperror.Handler(nil),
	})
}

func registerCreate(t *testing.T, app *fiber.App, reg *Registry) {
	t.Helper()
	Post(reg, app, "/items", func(ctx context.Context, req createRequest) (createResponse, error) {
		_ = ctx
		return createResponse{ID: req.Name}, nil
	}, Tag("items"))
}

func TestPostSuccess(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	registerCreate(t, app, reg)

	req := httptest.NewRequest(fiber.MethodPost, "/items", strings.NewReader(`{"name":"order-1"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 201 {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestResponseWrappedInResult(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	registerCreate(t, app, reg)

	req := httptest.NewRequest(fiber.MethodPost, "/items", strings.NewReader(`{"name":"order-1"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Result createResponse `json:"result"`
	}
	if err := jsonx.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result.ID != "order-1" {
		t.Fatalf("result = %+v, body = %s", envelope.Result, body)
	}
}

func TestPostRejectsUnknownMember(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	registerCreate(t, app, reg)

	req := httptest.NewRequest(fiber.MethodPost, "/items", strings.NewReader(`{"name":"x","hacked":true}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPostValidatesRequired(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	registerCreate(t, app, reg)

	req := httptest.NewRequest(fiber.MethodPost, "/items", strings.NewReader(`{"name":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmptyResponseIsNoContent(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	Delete(reg, app, "/items/:id", func(context.Context, struct{}) (Empty, error) {
		return Empty{}, nil
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodDelete, "/items/1", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 204 {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestRegistryRecordsRoute(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	registerCreate(t, app, reg)

	routes := reg.Routes()
	if len(routes) != 1 {
		t.Fatalf("routes = %d, want 1", len(routes))
	}
	route := routes[0]
	if route.Method != fiber.MethodPost || route.Path != "/items" {
		t.Fatalf("unexpected route: %+v", route)
	}
	if route.Tag != "items" {
		t.Fatalf("tag = %q", route.Tag)
	}
	if route.OperationID == "" || !strings.Contains(route.OperationID, "TestPostSuccess") {
		t.Logf("operation id = %q", route.OperationID)
	}
	if route.Request == nil || route.Response == nil {
		t.Fatal("request/response types must be recorded")
	}
}

func TestProtectedOption(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	Get(reg, app, "/me", func(context.Context, struct{}) (createResponse, error) {
		return createResponse{}, nil
	}, Protected())

	if !reg.Routes()[0].Protected {
		t.Fatal("route must be marked protected")
	}
}

func TestJoinPath(t *testing.T) {
	app := newTestApp()
	cases := []struct {
		name   string
		router fiber.Router
		path   string
		want   string
	}{
		{"plain app", app, "/items", "/items"},
		{"root group", app.Group(""), "/items", "/items"},
		{"slash group", app.Group("/"), "/items", "/items"},
		{"prefixed group", app.Group("/api/v1"), "/items", "/api/v1/items"},
		{"group with empty path", app.Group("/api/v1"), "", "/api/v1"},
		{"group without leading slash", app.Group("/api/v1"), "items", "/api/v1/items"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := joinPath(tc.router, tc.path); got != tc.want {
				t.Fatalf("joinPath = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRegistryRecordsGroupPrefix(t *testing.T) {
	app := newTestApp()
	reg := NewRegistry()
	api := app.Group("/api/v1")
	Get(reg, api, "/ping", func(context.Context, struct{}) (createResponse, error) {
		return createResponse{}, nil
	})

	if got := reg.Routes()[0].Path; got != "/api/v1/ping" {
		t.Fatalf("route path = %q, want /api/v1/ping", got)
	}
}

func TestRegistryRoutesSorted(t *testing.T) {
	reg := NewRegistry()
	reg.Add(Route{Method: fiber.MethodGet, Path: "/b"})
	reg.Add(Route{Method: fiber.MethodGet, Path: "/a"})
	reg.Add(Route{Method: fiber.MethodPost, Path: "/a"})
	routes := reg.Routes()
	if routes[0].Path != "/a" || routes[1].Path != "/a" || routes[2].Path != "/b" {
		t.Fatalf("routes are not sorted: %+v", routes)
	}
}
