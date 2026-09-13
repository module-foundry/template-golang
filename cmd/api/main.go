// Command api is the single entry point of the application.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"template-golang/internal/app"
	"template-golang/internal/config"
	"template-golang/pkg/jsonx"
	"template-golang/pkg/openapi"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	checkEnv := flag.Bool("check-env", false, "validate configuration and exit")
	dumpOpenAPI := flag.Bool("dump-openapi", false, "print the OpenAPI document and exit")
	healthcheck := flag.Bool("healthcheck", false, "probe the local /healthz endpoint and exit")
	flag.Parse()

	switch {
	case *checkEnv:
		runCheckEnv()
		return
	case *dumpOpenAPI:
		runDumpOpenAPI()
		return
	case *healthcheck:
		runHealthcheck()
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration is invalid", slog.String("error", err.Error()))
		os.Exit(1)
	}

	application, err := app.New(ctx, cfg, version)
	if err != nil {
		slog.Error("application failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := application.Run(ctx); err != nil {
		slog.Error("application stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func runCheckEnv() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("configuration OK (env=%s, addr=%s)\n", cfg.AppEnv, cfg.Addr())
}

func runDumpOpenAPI() {
	cfg := docsConfig()
	registry := app.BuildRegistry(cfg)
	spec := openapi.Build(registry, openapi.Config{
		Title:       cfg.AppName,
		Version:     version,
		Description: "Runtime-generated API reference",
		CookieName:  cfg.CookieName,
	})
	data, err := jsonx.MarshalDeterministic(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal openapi:", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func runHealthcheck() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "3001"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck failed:", err)
		os.Exit(1)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "healthcheck status:", resp.StatusCode)
		os.Exit(1)
	}
}

func docsConfig() *config.Config {
	cfg := &config.Config{
		AppEnv:          config.EnvDevelopment,
		AppName:         envOr("APP_NAME", "template-golang"),
		JWTSecret:       "docs-only",
		JWTTTL:          time.Hour,
		CookieName:      envOr("COOKIE_NAME", "access_token"),
		RedisAddr:       "localhost:6379",
		HTTPPort:        3001,
		ShutdownTimeout: 10 * time.Second,
	}
	if loaded, err := config.Load(); err == nil {
		return loaded
	}
	return cfg
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
