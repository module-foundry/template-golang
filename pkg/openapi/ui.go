package openapi

import (
	_ "embed"
	"log/slog"
	"sync"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/httpx"
	"template-golang/pkg/jsonx"
)

//go:embed assets/scalar.js
var scalarJS []byte

var (
	buildOnce sync.Once
	cached    []byte
	buildErr  error
)

// Document renders (and caches) the OpenAPI document as JSON.
func Document(reg *httpx.Registry, cfg Config) ([]byte, error) {
	buildOnce.Do(func() {
		cached, buildErr = jsonx.MarshalDeterministic(Build(reg, cfg))
	})
	return cached, buildErr
}

// Register mounts /openapi.json and the Scalar UI at /docs.
// Scalar assets are embedded: no CDN dependency, no build step.
func Register(router fiber.Router, reg *httpx.Registry, cfg Config, log *slog.Logger) {
	router.Get("/openapi.json", func(c fiber.Ctx) error {
		doc, err := Document(reg, cfg)
		if err != nil {
			return err
		}
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(doc)
	})

	router.Get("/docs/scalar.js", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "text/javascript; charset=utf-8")
		return c.Send(scalarJS)
	})

	router.Get("/docs", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(docsHTML)
	})

	if log != nil {
		log.Info("openapi enabled", slog.String("docs", "/docs"), slog.String("spec", "/openapi.json"))
	}
}

const docsHTML = `<!doctype html>
<html>
  <head>
    <title>API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.json"></script>
    <script src="/docs/scalar.js"></script>
  </body>
</html>`
