package openapi

import (
	"reflect"
	"testing"

	"template-golang/pkg/httpx"
	"template-golang/pkg/jsonx"
)

type miniAppRequest struct {
	InitData string `json:"init_data,omitempty" example:"query_id=1"`
}

type miniAppResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id" format:"uuid"`
}

func testRegistry() *httpx.Registry {
	reg := httpx.NewRegistry()
	reg.Add(httpx.Route{
		Method:      "POST",
		Path:        "/auth/mini-apps/telegram",
		OperationID: "auth.Telegram",
		Tag:         "auth",
		Request:     reflect.TypeOf(miniAppRequest{}),
		Response:    reflect.TypeOf(miniAppResponse{}),
	})
	reg.Add(httpx.Route{
		Method:      "GET",
		Path:        "/users/:id",
		OperationID: "user.Get",
		Tag:         "user",
		Request:     reflect.TypeOf(struct{}{}),
		Response:    reflect.TypeOf(map[string]string{}),
		Protected:   true,
	})
	return reg
}

func TestBuildContainsPathsAndSecurity(t *testing.T) {
	spec := Build(testRegistry(), Config{Title: "t", Version: "1", CookieName: "access_token"})

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("paths missing")
	}
	if _, ok := paths["/auth/mini-apps/telegram"]; !ok {
		t.Fatal("auth path missing")
	}
	usersPath, ok := paths["/users/{id}"].(map[string]any)
	if !ok {
		t.Fatal("path parameter must be converted to {id}")
	}
	get, _ := usersPath["get"].(map[string]any)
	if _, ok := get["security"]; !ok {
		t.Fatal("protected route must have security")
	}

	components, _ := spec["components"].(map[string]any)
	schemes, _ := components["securitySchemes"].(map[string]any)
	if _, ok := schemes["cookieAuth"]; !ok {
		t.Fatal("cookieAuth scheme missing")
	}
	if _, ok := schemes["bearerAuth"]; !ok {
		t.Fatal("bearerAuth scheme missing")
	}
	schemas, _ := components["schemas"].(map[string]any)
	if _, ok := schemas["ErrorEnvelope"]; !ok {
		t.Fatal("ErrorEnvelope schema missing")
	}
	if _, ok := schemas["miniAppResponse"]; !ok {
		t.Fatal("response schema missing")
	}
}

func TestSuccessEnvelope(t *testing.T) {
	spec := Build(testRegistry(), Config{Title: "t", Version: "1", CookieName: "c"})
	paths := spec["paths"].(map[string]any)
	post := paths["/auth/mini-apps/telegram"].(map[string]any)["post"].(map[string]any)
	responses := post["responses"].(map[string]any)
	created := responses["201"].(map[string]any)
	schema := created["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["result"]; !ok {
		t.Fatalf("success schema must expose result: %v", schema)
	}
	required := schema["required"].([]any)
	if len(required) != 1 || required[0] != "result" {
		t.Fatalf("result must be required: %v", required)
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	cfg := Config{Title: "t", Version: "1", CookieName: "access_token"}
	first, err := jsonx.MarshalDeterministic(Build(testRegistry(), cfg))
	if err != nil {
		t.Fatal(err)
	}
	second, err := jsonx.MarshalDeterministic(Build(testRegistry(), cfg))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("openapi build is not deterministic")
	}
}

func TestEnvelopeCodesMatchRegistry(t *testing.T) {
	spec := Build(testRegistry(), Config{Title: "t", Version: "1", CookieName: "c"})
	components := spec["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	envelope := schemas["ErrorEnvelope"].(map[string]any)
	props := envelope["properties"].(map[string]any)
	errObj := props["error"].(map[string]any)
	errProps := errObj["properties"].(map[string]any)
	enum := errProps["type"].(map[string]any)["enum"].([]any)
	if len(enum) < 10 {
		t.Fatalf("error code enum looks too small: %d", len(enum))
	}
}
