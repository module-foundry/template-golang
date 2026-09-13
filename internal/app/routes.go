package app

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"template-golang/internal/config"
	"template-golang/internal/middleware"
	"template-golang/internal/modules/auth"
	"template-golang/pkg/apperror"
	"template-golang/pkg/httpx"
	"template-golang/pkg/jwtx"
	"template-golang/pkg/openapi"
	"template-golang/pkg/postgres"
	"template-golang/pkg/redisx"
)

// RouteDeps carries everything route registration needs.
type RouteDeps struct {
	Config   *config.Config
	Log      *slog.Logger
	Registry *httpx.Registry
	Signer   *jwtx.Signer
	Pool     *pgxpool.Pool
	Redis    *redis.Client
	Version  string
}

// RegisterRoutes is the single place where modules are mounted.
//
// Public and protected routers are handed to module Register functions; each
// module is its own composition root (router.go) and only receives the
// infrastructure it needs.
func RegisterRoutes(root fiber.Router, deps RouteDeps) {
	registerHealth(root, deps)

	public := root.Group("")
	protected := root.Group("/", middleware.Auth(deps.Signer))

	auth.Register(public, deps.Registry, auth.Deps{
		Config: deps.Config,
		Signer: deps.Signer,
	})

	_ = protected // new protected modules are registered here, e.g. user.Register(protected, ...)

	if deps.Config.FeatureOpenAPIEnabled {
		openapi.Register(root, deps.Registry, openapi.Config{
			Title:       deps.Config.AppName,
			Version:     deps.Version,
			Description: "Runtime-generated API reference",
			CookieName:  deps.Config.CookieName,
		}, deps.Log)
	}
}

func registerHealth(root fiber.Router, deps RouteDeps) {
	root.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})
	root.Get("/readyz", func(c fiber.Ctx) error {
		if err := postgres.Healthcheck(c.Context(), deps.Pool); err != nil {
			return apperror.Runtime(apperror.CodeDBConnectFailed, err)
		}
		if err := redisx.Healthcheck(c.Context(), deps.Redis); err != nil {
			return apperror.Runtime(apperror.CodeRedisFailed, err)
		}
		return c.JSON(map[string]string{"status": "ready"})
	})
}

// BuildRegistry builds the route registry without touching databases.
// It is used to dump docs/openapi.json (task openapi:dump).
func BuildRegistry(cfg *config.Config) *httpx.Registry {
	registry := httpx.NewRegistry()
	probe := fiber.New(fiber.Config{AppName: cfg.AppName})
	auth.Register(probe.Group(""), registry, auth.Deps{
		Config: cfg,
		Signer: jwtx.New(cfg.JWTSecret, cfg.JWTTTL, cfg.CookieName),
	})
	return registry
}
