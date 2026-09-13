package app

import (
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"template-golang/internal/config"
	"template-golang/pkg/apperror"
	"template-golang/pkg/httpx"
	"template-golang/pkg/jsonx"
	"template-golang/pkg/jwtx"
	"template-golang/pkg/openapi"
)

func testConfig() *config.Config {
	return &config.Config{
		AppEnv:      config.EnvDevelopment,
		AppName:     "template-golang",
		JWTSecret:   "test-secret",
		JWTTTL:      time.Hour,
		CookieName:  "access_token",
		APIBasePath: "/api/v1",
	}
}

// TestRouteSnapshot freezes the public route table. Any accidental change to
// routes, methods or operation ids fails here before it reaches clients.
func TestRouteSnapshot(t *testing.T) {
	cfg := testConfig()
	routes := BuildRegistry(cfg).Routes()
	type snapshot struct {
		Method      string
		Path        string
		OperationID string
		Protected   bool
	}
	got := make([]snapshot, 0, len(routes))
	for _, route := range routes {
		got = append(got, snapshot{route.Method, route.Path, route.OperationID, route.Protected})
	}

	want := []snapshot{
		{Method: "POST", Path: "/api/v1/auth/mini-apps/telegram", OperationID: "auth.Handler.Telegram", Protected: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("route contract changed:\n got: %+v\nwant: %+v", got, want)
	}
}

// TestRuntimeMounting guards the Fiber group ordering: docs and public routes
// live under the API prefix without auth, while the old paths are gone.
func TestRuntimeMounting(t *testing.T) {
	cfg := testConfig()
	cfg.FeatureOpenAPIEnabled = true
	server := fiber.New(fiber.Config{
		JSONEncoder:  jsonx.FiberEncoder,
		JSONDecoder:  jsonx.FiberDecoder,
		ErrorHandler: apperror.Handler(nil),
	})
	RegisterRoutes(server, RouteDeps{
		Config:   cfg,
		Registry: httpx.NewRegistry(),
		Signer:   jwtx.New(cfg.JWTSecret, cfg.JWTTTL, cfg.CookieName),
	})

	cases := []struct {
		method string
		path   string
		want   int
	}{
		{fiber.MethodGet, "/api/v1/docs", fiber.StatusOK},
		{fiber.MethodGet, "/api/v1/openapi.json", fiber.StatusOK},
		{fiber.MethodPost, "/api/v1/auth/mini-apps/telegram", fiber.StatusCreated},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		resp, err := server.Test(req)
		if err != nil {
			t.Fatalf("%s %s: %v", tc.method, tc.path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, resp.StatusCode, tc.want)
		}
	}
}

// TestOpenAPIGolden keeps docs/openapi.json in sync with the code.
func TestOpenAPIGolden(t *testing.T) {
	cfg := testConfig()
	registry := BuildRegistry(cfg)
	spec := openapi.Build(registry, openapi.Config{
		Title:       cfg.AppName,
		Version:     "dev",
		Description: "Runtime-generated API reference",
		CookieName:  cfg.CookieName,
	})
	got, err := jsonx.MarshalDeterministic(spec)
	if err != nil {
		t.Fatal(err)
	}

	goldenPath := "../../docs/openapi.json"
	stored, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden file is missing, run `task openapi:dump`: %v", err)
	}

	var gotDoc, wantDoc any
	if err := jsonx.Unmarshal(got, &gotDoc); err != nil {
		t.Fatal(err)
	}
	if err := jsonx.Unmarshal(stored, &wantDoc); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotDoc, wantDoc) {
		t.Fatalf("%s is out of date: run `task openapi:dump` and review the diff", goldenPath)
	}
}
