package openapi

import (
	"reflect"
	"testing"
	"time"

	"uuid"

	"template-golang/pkg/httpx"
)

type kitchenSink struct {
	Flag      bool              `json:"flag"`
	Count     int               `json:"count"`
	Big       int64             `json:"big,omitempty"`
	Ratio     float64           `json:"ratio"`
	Title     string            `json:"title" enum:"a,b,c" example:"a"`
	Optional  *string           `json:"optional"`
	Moment    time.Time         `json:"moment"`
	ID        uuid.UUID         `json:"id"`
	Tags      []string          `json:"tags"`
	Labels    map[string]string `json:"labels"`
	Hidden    string            `json:"-"`
	private   string            //nolint:unused
	Nested    nested            `json:"nested"`
	Formatted string            `json:"formatted" format:"email"`
}

type nested struct {
	Value string `json:"value"`
}

func TestSchemaKitchenSink(t *testing.T) {
	schemas := map[string]any{}
	schema := schemaFor(reflect.TypeOf(kitchenSink{}), schemas)

	if schema["$ref"] == nil {
		t.Fatalf("expected a component ref, got %v", schema)
	}
	component, ok := schemas["kitchenSink"].(map[string]any)
	if !ok {
		t.Fatalf("component missing: %v", schemas)
	}
	props := component["properties"].(map[string]any)

	checks := map[string]string{
		"flag":      "boolean",
		"count":     "integer",
		"ratio":     "number",
		"moment":    "string",
		"id":        "string",
		"tags":      "array",
		"labels":    "object",
		"nested":    "object",
		"formatted": "string",
	}
	for field, kind := range checks {
		prop, ok := props[field].(map[string]any)
		if !ok {
			t.Fatalf("property %s missing", field)
		}
		if prop["type"] != kind && prop["$ref"] == nil {
			t.Errorf("%s type = %v, want %s (%v)", field, prop["type"], kind, prop)
		}
	}
	if _, ok := props["Hidden"]; ok {
		t.Error("json:\"-\" field must be skipped")
	}

	if props["moment"].(map[string]any)["format"] != "date-time" {
		t.Error("time.Time must map to date-time")
	}
	if props["id"].(map[string]any)["format"] != "uuid" {
		t.Error("uuid.UUID must map to uuid format")
	}
	if props["formatted"].(map[string]any)["format"] != "email" {
		t.Error("format tag must be honoured")
	}
	if _, ok := props["title"].(map[string]any)["enum"]; !ok {
		t.Error("enum tag must be honoured")
	}

	required := component["required"].([]any)
	requiredSet := map[string]bool{}
	for _, name := range required {
		requiredSet[name.(string)] = true
	}
	if !requiredSet["flag"] || requiredSet["optional"] || requiredSet["big"] {
		t.Errorf("required fields wrong: %v", required)
	}
}

func TestSchemaForPointerAndCollections(t *testing.T) {
	schemas := map[string]any{}
	ptr := schemaFor(reflect.TypeOf((*string)(nil)), schemas)
	if ptr["nullable"] != true {
		t.Fatalf("pointer must be nullable: %v", ptr)
	}
	arr := schemaFor(reflect.TypeOf([]kitchenSink{}), schemas)
	if arr["type"] != "array" {
		t.Fatalf("slice must be array: %v", arr)
	}
	m := schemaFor(reflect.TypeOf(map[string]int{}), schemas)
	if m["additionalProperties"].(map[string]any)["type"] != "integer" {
		t.Fatalf("map values must map: %v", m)
	}
}

func TestFiberPathToOpenAPI(t *testing.T) {
	if got := fiberPathToOpenAPI("/users/:id/posts/:post"); got != "/users/{id}/posts/{post}" {
		t.Fatalf("got %q", got)
	}
	if params := pathParameters("/users/:id"); len(params) != 1 {
		t.Fatalf("params = %v", params)
	}
}

func TestEmptyResponseMapping(t *testing.T) {
	reg := httpx.NewRegistry()
	reg.Add(httpx.Route{Method: "DELETE", Path: "/x", OperationID: "x.Delete", Response: reflect.TypeOf(httpx.Empty{})})
	spec := Build(reg, Config{Title: "t", Version: "1", CookieName: "c"})
	paths := spec["paths"].(map[string]any)
	op := paths["/x"].(map[string]any)["delete"].(map[string]any)
	responses := op["responses"].(map[string]any)
	if _, ok := responses["204"]; !ok {
		t.Fatalf("204 response expected: %v", responses)
	}
}
