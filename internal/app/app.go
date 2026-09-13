// Package app wires all dependencies together. It is the only place where
// concrete implementations are bound to interfaces.
package app

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"template-golang/internal/config"
	"template-golang/internal/middleware"
	"template-golang/migrations"
	"template-golang/pkg/apperror"
	"template-golang/pkg/httpx"
	"template-golang/pkg/jsonx"
	"template-golang/pkg/jwtx"
	"template-golang/pkg/logger"
	"template-golang/pkg/postgres"
	"template-golang/pkg/redisx"
)

// App owns the lifecycle of every dependency.
type App struct {
	cfg      *config.Config
	log      *slog.Logger
	fiber    *fiber.App
	pool     *pgxpool.Pool
	redis    *redis.Client
	registry *httpx.Registry
}

// New builds the application: logger, databases, migrations, HTTP server.
func New(ctx context.Context, cfg *config.Config, version string) (*App, error) {
	log := logger.New(logger.Config{
		Level:   cfg.LogLevel,
		Format:  cfg.LogFormat,
		Env:     cfg.AppEnv,
		Service: cfg.AppName,
		Version: version,
	})
	apperror.Configure(cfg.ErrorTraceEnabled, cfg.ErrorStacktraceEnabled)

	pool, err := postgres.New(ctx, postgres.Config{
		DSN:      cfg.PostgresDSN,
		MaxConns: cfg.PostgresMaxConns,
		MinConns: cfg.PostgresMinConns,
	})
	if err != nil {
		return nil, apperror.Runtime(apperror.CodeDBConnectFailed, err)
	}

	if cfg.DBAutoMigrate {
		if err := migrations.Run(ctx, pool, log); err != nil {
			pool.Close()
			return nil, apperror.Runtime(apperror.CodeDBMigrationFailed, err)
		}
	}

	rdb, err := redisx.New(ctx, redisx.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		pool.Close()
		return nil, apperror.Runtime(apperror.CodeRedisFailed, err)
	}

	signer := jwtx.New(cfg.JWTSecret, cfg.JWTTTL, cfg.CookieName)
	registry := httpx.NewRegistry()

	server := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		JSONEncoder:  jsonx.FiberEncoder,
		JSONDecoder:  jsonx.FiberDecoder,
		ErrorHandler: apperror.Handler(log),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
		BodyLimit:    4 * 1024 * 1024,
	})

	server.Use(
		middleware.RequestID(),
		middleware.RequestLogger(log),
		recover.New(),
		cors.New(cors.Config{
			AllowOrigins:     cfg.CORSOrigins(),
			AllowCredentials: true,
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		}),
		helmet.New(),
	)
	if cfg.FeaturePprofEnabled {
		server.Use(pprof.New())
	}

	RegisterRoutes(server, RouteDeps{
		Config:   cfg,
		Log:      log,
		Registry: registry,
		Signer:   signer,
		Pool:     pool,
		Redis:    rdb,
		Version:  version,
	})

	return &App{cfg: cfg, log: log, fiber: server, pool: pool, redis: rdb, registry: registry}, nil
}

// Run starts the HTTP server and blocks until the context is canceled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.log.Info("http server started", slog.String("addr", a.cfg.Addr()), slog.String("env", a.cfg.AppEnv))
		errCh <- a.fiber.Listen(a.cfg.Addr())
	}()

	select {
	case err := <-errCh:
		a.close()
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()
		a.log.Info("shutting down")
		err := a.fiber.ShutdownWithContext(shutdownCtx)
		a.close()
		return err
	}
}

func (a *App) close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.pool != nil {
		a.pool.Close()
	}
}
