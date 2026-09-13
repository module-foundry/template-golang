// Package apperror provides a small, stable error registry and constructors.
//
// Every error that leaves the application carries a Code from this file.
// HTTP status values MUST use fiber/http constants - numeric literals are
// forbidden and enforced by a test.
package apperror

import "github.com/gofiber/fiber/v3"

// Kind splits errors into two zones.
type Kind string

const (
	// KindAPI marks client errors (4xx). They are safe to expose.
	KindAPI Kind = "API"
	// KindRuntime marks server/infrastructure errors (5xx). Details stay in logs.
	KindRuntime Kind = "Runtime"
)

// Code is a stable, frontend-facing error identifier.
type Code string

// Minimal, stable set of codes. Reuse before adding a new one:
// specifics belong to message/details, not to a new code.
const (
	// API kind.
	CodeValidationFailed Code = "VALIDATION_FAILED"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeTokenExpired     Code = "TOKEN_EXPIRED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeNotFound         Code = "NOT_FOUND"
	CodeConflict         Code = "CONFLICT"
	CodeRateLimited      Code = "RATE_LIMITED"
	CodePayloadTooLarge  Code = "PAYLOAD_TOO_LARGE"

	// Runtime kind.
	CodeInternal          Code = "INTERNAL"
	CodeDBQueryFailed     Code = "DB_QUERY_FAILED"
	CodeDBConnectFailed   Code = "DB_CONNECT_FAILED"
	CodeDBMigrationFailed Code = "DB_MIGRATION_FAILED"
	CodeRedisFailed       Code = "REDIS_FAILED"
	CodeHTTPClientFailed  Code = "HTTP_CLIENT_FAILED"
	CodePanic             Code = "PANIC"
	CodeConfigInvalid     Code = "CONFIG_INVALID"
)

// Info describes a code: its zone, HTTP status and the message shown to clients.
type Info struct {
	Kind          Kind
	HTTPStatus    int
	PublicMessage string
}

// registry is the single source of truth for error metadata.
var registry = map[Code]Info{
	CodeValidationFailed: {KindAPI, fiber.StatusUnprocessableEntity, "validation failed"},
	CodeUnauthorized:     {KindAPI, fiber.StatusUnauthorized, "unauthorized"},
	CodeTokenExpired:     {KindAPI, fiber.StatusUnauthorized, "token expired"},
	CodeForbidden:        {KindAPI, fiber.StatusForbidden, "forbidden"},
	CodeNotFound:         {KindAPI, fiber.StatusNotFound, "not found"},
	CodeConflict:         {KindAPI, fiber.StatusConflict, "conflict"},
	CodeRateLimited:      {KindAPI, fiber.StatusTooManyRequests, "too many requests"},
	CodePayloadTooLarge:  {KindAPI, fiber.StatusRequestEntityTooLarge, "payload too large"},

	CodeInternal:          {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodeDBQueryFailed:     {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodeDBConnectFailed:   {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodeDBMigrationFailed: {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodeRedisFailed:       {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodeHTTPClientFailed:  {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodePanic:             {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
	CodeConfigInvalid:     {KindRuntime, fiber.StatusInternalServerError, "internal server error"},
}

// InfoOf returns metadata for a code. Unknown codes fall back to CodeInternal.
func InfoOf(code Code) Info {
	info, ok := registry[code]
	if !ok {
		return registry[CodeInternal]
	}
	return info
}

// Codes returns every registered code (used by tests and OpenAPI generation).
func Codes() map[Code]Info {
	out := make(map[Code]Info, len(registry))
	for code, info := range registry {
		out[code] = info
	}
	return out
}
