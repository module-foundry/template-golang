// Package httpx provides typed route helpers and a route registry.
//
// Every public route is registered through these helpers: method, path,
// request/response types and auth flag are recorded and later used to
// generate OpenAPI without any annotations in handlers.
package httpx

import (
	"bytes"
	"context"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jsonx"
	"template-golang/pkg/validx"
)

// Handler is a typed handler that never touches Fiber types.
type Handler[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)

// HandlerWithCtx is a typed handler that needs the Fiber context (cookies, headers).
type HandlerWithCtx[Req, Resp any] func(c fiber.Ctx, req Req) (Resp, error)

// Empty is the response type for endpoints that return no body (204).
type Empty struct{}

// Route describes a registered route for OpenAPI generation.
type Route struct {
	Method      string
	Path        string
	OperationID string
	Tag         string
	Request     reflect.Type
	Response    reflect.Type
	Protected   bool
}

// Registry collects routes. It is built once at startup.
type Registry struct {
	mu     sync.RWMutex
	routes []Route
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{} }

// Add records a route.
func (r *Registry) Add(route Route) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, route)
}

// Routes returns a copy of the registered routes sorted by path and method.
func (r *Registry) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Route, len(r.routes))
	copy(out, r.routes)
	sortRoutes(out)
	return out
}

type config struct {
	protected   bool
	tag         string
	operationID string
}

// Option customises a route registration.
type Option func(*config)

// Protected marks the route as requiring authentication.
func Protected() Option { return func(c *config) { c.protected = true } }

// Tag sets the OpenAPI tag (defaults to the module package name).
func Tag(name string) Option { return func(c *config) { c.tag = name } }

// OperationID overrides the generated operation id.
func OperationID(id string) Option { return func(c *config) { c.operationID = id } }

// Get registers a GET route.
func Get[Req, Resp any](reg *Registry, router fiber.Router, path string, fn Handler[Req, Resp], opts ...Option) {
	register(reg, router, fiber.MethodGet, path, fn, nil, opts...)
}

// Post registers a POST route.
func Post[Req, Resp any](reg *Registry, router fiber.Router, path string, fn Handler[Req, Resp], opts ...Option) {
	register(reg, router, fiber.MethodPost, path, fn, nil, opts...)
}

// PostWithCtx registers a POST route that needs the Fiber context.
func PostWithCtx[Req, Resp any](reg *Registry, router fiber.Router, path string, fn HandlerWithCtx[Req, Resp], opts ...Option) {
	register(reg, router, fiber.MethodPost, path, nil, fn, opts...)
}

// Patch registers a PATCH route.
func Patch[Req, Resp any](reg *Registry, router fiber.Router, path string, fn Handler[Req, Resp], opts ...Option) {
	register(reg, router, fiber.MethodPatch, path, fn, nil, opts...)
}

// Delete registers a DELETE route.
func Delete[Req, Resp any](reg *Registry, router fiber.Router, path string, fn Handler[Req, Resp], opts ...Option) {
	register(reg, router, fiber.MethodDelete, path, fn, nil, opts...)
}

func register[Req, Resp any](
	reg *Registry,
	router fiber.Router,
	method, path string,
	fn Handler[Req, Resp],
	fnCtx HandlerWithCtx[Req, Resp],
	opts ...Option,
) {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	sample := any(fn)
	if fnCtx != nil {
		sample = any(fnCtx)
	}
	operationID := cfg.operationID
	if operationID == "" {
		operationID = deriveOperationID(sample)
	}
	tag := cfg.tag
	if tag == "" {
		tag = deriveTag(operationID)
	}

	var req Req
	route := Route{
		Method:      method,
		Path:        joinPath(router, path),
		OperationID: operationID,
		Tag:         tag,
		Request:     reflect.TypeOf(&req).Elem(),
		Response:    reflect.TypeOf((*Resp)(nil)).Elem(),
		Protected:   cfg.protected,
	}
	reg.Add(route)

	handler := func(c fiber.Ctx) error {
		var out Req
		if method != fiber.MethodGet && method != fiber.MethodDelete {
			body := bytes.TrimSpace(c.Body())
			if len(body) > 0 {
				if err := jsonx.UnmarshalStrict(body, &out); err != nil {
					return apperror.API(apperror.CodeValidationFailed, "malformed json body")
				}
			}
		}
		if err := validx.Default().ValidateStruct(out); err != nil {
			return err
		}

		var (
			resp Resp
			err  error
		)
		if fnCtx != nil {
			resp, err = fnCtx(c, out)
		} else {
			resp, err = fn(c.Context(), out)
		}
		if err != nil {
			return err
		}
		if _, ok := any(resp).(Empty); ok {
			return c.SendStatus(fiber.StatusNoContent)
		}
		status := fiber.StatusOK
		if method == fiber.MethodPost {
			status = fiber.StatusCreated
		}
		return c.Status(status).JSON(resp)
	}

	switch method {
	case fiber.MethodGet:
		router.Get(path, handler)
	case fiber.MethodPost:
		router.Post(path, handler)
	case fiber.MethodPatch:
		router.Patch(path, handler)
	case fiber.MethodDelete:
		router.Delete(path, handler)
	}
}

// joinPath returns the path as mounted by the router: the group prefix plus
// the route path. The registry stores full paths so OpenAPI matches the real
// routes. Routers without a group prefix (apps) keep the path unchanged.
func joinPath(router fiber.Router, path string) string {
	group, ok := router.(*fiber.Group)
	if !ok {
		return path
	}
	prefix := strings.TrimRight(group.Prefix, "/")
	if prefix == "" {
		return path
	}
	if path == "" {
		return prefix
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return prefix + path
}

func deriveOperationID(fn any) string {
	if fn == nil {
		return "unknown"
	}
	full := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if idx := strings.LastIndex(full, "/"); idx >= 0 {
		full = full[idx+1:]
	}
	full = strings.TrimSuffix(full, "-fm")
	full = strings.ReplaceAll(full, ".(*", ".")
	full = strings.ReplaceAll(full, ").", ".")
	full = strings.TrimSuffix(full, ")")
	return full
}

func deriveTag(operationID string) string {
	if idx := strings.Index(operationID, "."); idx > 0 {
		return operationID[:idx]
	}
	return "default"
}

func sortRoutes(routes []Route) {
	slices.SortFunc(routes, func(a, b Route) int {
		if a.Path != b.Path {
			return strings.Compare(a.Path, b.Path)
		}
		return strings.Compare(a.Method, b.Method)
	})
}
