package openapi

import (
	"reflect"
	"testing"

	"template-golang/pkg/httpx"
)

func TestSchemaForNilAndUnknown(t *testing.T) {
	schemas := map[string]any{}
	if got := schemaFor(nil, schemas); got["type"] != "object" {
		t.Fatalf("nil type = %v", got)
	}
	if got := schemaFor(reflect.TypeOf(make(chan int)), schemas); got["type"] != "object" {
		t.Fatalf("unknown kind = %v", got)
	}
	if got := schemaFor(reflect.TypeOf(0), schemas); got["type"] != "integer" {
		t.Fatalf("int kind = %v", got)
	}
}

func TestSchemaForAnonymousStruct(t *testing.T) {
	schemas := map[string]any{}
	schema := schemaFor(reflect.TypeOf(struct {
		Value string `json:"value"`
	}{}), schemas)
	if schema["type"] != "object" {
		t.Fatalf("anonymous struct must be inline: %v", schema)
	}
	props := schema["properties"].(map[string]any)
	if props["value"] == nil {
		t.Fatalf("value property missing: %v", schema)
	}
	if len(schemas) != 0 {
		t.Fatalf("anonymous struct must not create components: %v", schemas)
	}
}

func TestBuildWithoutRequestBody(t *testing.T) {
	reg := httpx.NewRegistry()
	reg.Add(httpx.Route{Method: "GET", Path: "/empty", OperationID: "empty.Get"})
	spec := Build(reg, Config{Title: "t", Version: "1", CookieName: "c"})
	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/empty"]; !ok {
		t.Fatalf("path missing: %v", paths)
	}
}

func TestResponsesForProtected(t *testing.T) {
	reg := httpx.NewRegistry()
	reg.Add(httpx.Route{
		Method: "GET", Path: "/private", OperationID: "p.Get",
		Response:  reflect.TypeOf(map[string]string{}),
		Protected: true,
	})
	spec := Build(reg, Config{Title: "t", Version: "1", CookieName: "c"})
	paths := spec["paths"].(map[string]any)
	op := paths["/private"].(map[string]any)["get"].(map[string]any)
	responses := op["responses"].(map[string]any)
	if _, ok := responses["401"]; !ok {
		t.Fatalf("401 must be documented: %v", responses)
	}
	if _, ok := responses["default"]; !ok {
		t.Fatalf("default error must be documented: %v", responses)
	}
}
