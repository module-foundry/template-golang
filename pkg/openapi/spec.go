// Package openapi builds an OpenAPI 3.0.3 document from the typed route
// registry and serves it together with the Scalar UI at runtime.
package openapi

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"template-golang/pkg/apperror"
	"template-golang/pkg/httpx"
)

// Config describes the generated document.
type Config struct {
	Title       string
	Version     string
	Description string
	CookieName  string
	// BasePath is the API prefix the UI assets are served under, e.g. /api/v1.
	BasePath string
}

// Build generates an OpenAPI 3.0.3 document. It is deterministic: routes are
// sorted, so the result is stable and can be golden-tested.
func Build(reg *httpx.Registry, cfg Config) map[string]any {
	components := map[string]any{}
	schemas := map[string]any{}

	paths := map[string]any{}
	for _, route := range reg.Routes() {
		path := fiberPathToOpenAPI(route.Path)
		item, _ := paths[path].(map[string]any)
		if item == nil {
			item = map[string]any{}
			paths[path] = item
		}

		operation := map[string]any{
			"operationId": route.OperationID,
			"tags":        []any{route.Tag},
			"responses":   responsesFor(route, schemas),
		}
		if route.Protected {
			operation["security"] = []any{
				map[string]any{"cookieAuth": []any{}},
				map[string]any{"bearerAuth": []any{}},
			}
		}
		if params := pathParameters(route.Path); len(params) > 0 {
			operation["parameters"] = params
		}
		if route.Method != http.MethodGet && route.Method != http.MethodDelete {
			body := map[string]any{"required": true}
			if schema := schemaForType(route.Request, schemas); schema != nil {
				body["content"] = map[string]any{
					"application/json": map[string]any{"schema": schema},
				}
			}
			operation["requestBody"] = body
		}
		item[strings.ToLower(route.Method)] = operation
	}

	schemas["ErrorEnvelope"] = errorEnvelopeSchema()
	components["schemas"] = schemas
	components["securitySchemes"] = map[string]any{
		"cookieAuth": map[string]any{
			"type": "apiKey",
			"in":   "cookie",
			"name": cfg.CookieName,
		},
		"bearerAuth": map[string]any{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		},
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       cfg.Title,
			"version":     cfg.Version,
			"description": cfg.Description,
		},
		"paths":      paths,
		"components": components,
	}
}

func responsesFor(route httpx.Route, schemas map[string]any) map[string]any {
	responses := map[string]any{}
	status := http.StatusOK
	if route.Method == http.MethodPost {
		status = http.StatusCreated
	}
	if isEmptyType(route.Response) {
		responses[fmt.Sprintf("%d", http.StatusNoContent)] = map[string]any{"description": "No Content"}
	} else {
		schema := schemaForType(route.Response, schemas)
		responses[fmt.Sprintf("%d", status)] = map[string]any{
			"description": "Success",
			"content": map[string]any{
				"application/json": map[string]any{"schema": schema},
			},
		}
	}
	if route.Protected {
		responses[fmt.Sprintf("%d", http.StatusUnauthorized)] = errorResponse("Unauthorized")
	}
	responses["default"] = errorResponse("Error")
	return responses
}

func errorResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/ErrorEnvelope"},
			},
		},
	}
}

func errorEnvelopeSchema() map[string]any {
	codes := apperror.Codes()
	codeList := make([]string, 0, len(codes))
	for code := range codes {
		codeList = append(codeList, string(code))
	}
	sort.Strings(codeList)
	codeEnum := make([]any, 0, len(codeList))
	for _, code := range codeList {
		codeEnum = append(codeEnum, code)
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"error": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type":    map[string]any{"type": "string", "enum": codeEnum},
					"message": map[string]any{"type": "string"},
				},
				"required": []any{"type", "message"},
			},
		},
		"required": []any{"error"},
	}
}

func fiberPathToOpenAPI(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

func pathParameters(path string) []any {
	var params []any
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, ":") {
			params = append(params, map[string]any{
				"name":     strings.TrimPrefix(part, ":"),
				"in":       "path",
				"required": true,
				"schema":   map[string]any{"type": "string"},
			})
		}
	}
	return params
}

func isEmptyType(t reflect.Type) bool {
	return t == reflect.TypeOf(httpx.Empty{})
}

func schemaForType(t reflect.Type, schemas map[string]any) map[string]any {
	schema := schemaFor(t, schemas)
	return schema
}

func schemaFor(t reflect.Type, schemas map[string]any) map[string]any {
	if t == nil {
		return map[string]any{"type": "object"}
	}
	if t == reflect.TypeOf(time.Time{}) {
		return map[string]any{"type": "string", "format": "date-time"}
	}
	if t.Name() == "UUID" {
		return map[string]any{"type": "string", "format": "uuid"}
	}

	switch t.Kind() {
	case reflect.Pointer:
		inner := schemaFor(t.Elem(), schemas)
		inner["nullable"] = true
		return inner
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if t.PkgPath() == "time" && t.Name() == "Duration" {
			return map[string]any{"type": "string", "format": "duration"}
		}
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.String:
		if t.Name() == "UUID" {
			return map[string]any{"type": "string", "format": "uuid"}
		}
		return map[string]any{"type": "string"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": schemaFor(t.Elem(), schemas)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": schemaFor(t.Elem(), schemas)}
	case reflect.Struct:
		return structSchema(t, schemas)
	default:
		return map[string]any{"type": "object"}
	}
}

func structSchema(t reflect.Type, schemas map[string]any) map[string]any {
	name := t.Name()
	if name != "" {
		if _, exists := schemas[name]; exists {
			return map[string]any{"$ref": "#/components/schemas/" + name}
		}
		schemas[name] = map[string]any{} // placeholder to break cycles
	}

	properties := map[string]any{}
	required := []any{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag := field.Tag.Get("json")
		fieldName := strings.SplitN(jsonTag, ",", 2)[0]
		if fieldName == "-" {
			continue
		}
		if fieldName == "" {
			fieldName = field.Name
		}
		prop := schemaFor(field.Type, schemas)
		if format := field.Tag.Get("format"); format != "" {
			prop["format"] = format
		}
		if example := field.Tag.Get("example"); example != "" {
			prop["example"] = example
		}
		if enum := field.Tag.Get("enum"); enum != "" {
			values := strings.Split(enum, ",")
			options := make([]any, 0, len(values))
			for _, value := range values {
				options = append(options, value)
			}
			prop["enum"] = options
		}
		properties[fieldName] = prop

		omitempty := strings.Contains(jsonTag, "omitempty")
		if !omitempty && field.Type.Kind() != reflect.Pointer {
			required = append(required, fieldName)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	if name != "" {
		schemas[name] = schema
		return map[string]any{"$ref": "#/components/schemas/" + name}
	}
	return schema
}
