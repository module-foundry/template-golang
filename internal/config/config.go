// Package config loads and validates the application configuration.
//
// Rules: every environment variable lives here, has a typed field and is
// validated before the server starts. .env is loaded for local development
// only; real environment variables always win.
package config

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Environment names.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"

	// DefaultJWTSecret is used in .env.example; rejected in production.
	DefaultJWTSecret = "dev-secret-change-me"
)

// Config is the typed application configuration.
type Config struct {
	AppEnv  string `env:"APP_ENV,default=development"`
	AppName string `env:"APP_NAME,default=template-golang"`

	LogLevel               string `env:"LOG_LEVEL,default=info"`
	LogFormat              string `env:"LOG_FORMAT,default=auto"`
	ErrorTraceEnabled      bool   `env:"ERROR_TRACE_ENABLED,default=true"`
	ErrorStacktraceEnabled bool   `env:"ERROR_STACKTRACE_ENABLED,default=false"`

	HTTPHost         string        `env:"HTTP_HOST,default=0.0.0.0"`
	HTTPPort         int           `env:"HTTP_PORT,default=3001"`
	HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT,default=10s"`
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT,default=10s"`
	HTTPIdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT,default=60s"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT,default=10s"`

	// APIBasePath is the prefix every API route (including /docs) is mounted
	// under. Empty mounts the API at the server root.
	APIBasePath string `env:"API_BASE_PATH,default=/api/v1"`

	PostgresDSN      string `env:"POSTGRES_DSN,required"`
	PostgresMaxConns int32  `env:"POSTGRES_MAX_CONNS,default=10"`
	PostgresMinConns int32  `env:"POSTGRES_MIN_CONNS,default=1"`
	DBAutoMigrate    bool   `env:"DB_AUTO_MIGRATE,default=true"`

	RedisAddr     string `env:"REDIS_ADDR,default=localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD,default="`
	RedisDB       int    `env:"REDIS_DB,default=0"`

	JWTSecret    string        `env:"JWT_SECRET,required"`
	JWTTTL       time.Duration `env:"JWT_TTL,default=24h"`
	CookieName   string        `env:"COOKIE_NAME,default=access_token"`
	CookieDomain string        `env:"COOKIE_DOMAIN,default="`
	CookieSecure bool          `env:"COOKIE_SECURE,default=false"`

	CORSAllowedOriginsRaw string `env:"CORS_ALLOWED_ORIGINS,default=http://localhost:3000"`

	FeatureOpenAPIEnabled bool `env:"FEATURE_OPENAPI_ENABLED,default=true"`
	FeaturePprofEnabled   bool `env:"FEATURE_PPROF_ENABLED,default=false"`
}

// IsProduction reports whether the app runs in production.
func (c *Config) IsProduction() bool { return c.AppEnv == EnvProduction }

// Addr is the listen address.
func (c *Config) Addr() string { return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort) }

// BasePath returns the API prefix normalized: a leading slash and no trailing
// slash. An empty value mounts the API at the server root.
func (c *Config) BasePath() string { return normalizeBasePath(c.APIBasePath) }

func normalizeBasePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(path, "/")
}

// CORSOrigins returns the comma-separated allow-list, trimmed and deduplicated.
func (c *Config) CORSOrigins() []string {
	seen := map[string]struct{}{}
	var origins []string
	for _, raw := range strings.Split(c.CORSAllowedOriginsRaw, ",") {
		origin := strings.TrimSpace(raw)
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins
}

// Load reads .env (best effort), decodes environment variables and validates.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := envconfig.Process(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate aggregates all configuration problems into one error.
func (c *Config) Validate() error {
	var problems []error

	switch c.AppEnv {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		problems = append(problems, fmt.Errorf("APP_ENV must be one of %s|%s|%s", EnvDevelopment, EnvStaging, EnvProduction))
	}
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		problems = append(problems, fmt.Errorf("HTTP_PORT must be in range 1..65535"))
	}
	if strings.ContainsAny(c.APIBasePath, " \t\r\n?#") {
		problems = append(problems, fmt.Errorf("API_BASE_PATH must be a URL path without spaces, query or fragment"))
	}
	if strings.TrimSpace(c.JWTSecret) == "" {
		problems = append(problems, fmt.Errorf("JWT_SECRET is required"))
	}
	if c.JWTTTL <= 0 {
		problems = append(problems, fmt.Errorf("JWT_TTL must be positive"))
	}
	if c.PostgresMaxConns < c.PostgresMinConns {
		problems = append(problems, fmt.Errorf("POSTGRES_MAX_CONNS must be >= POSTGRES_MIN_CONNS"))
	}

	for _, origin := range c.CORSOrigins() {
		if origin == "*" {
			problems = append(problems, fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain *"))
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			problems = append(problems, fmt.Errorf("CORS_ALLOWED_ORIGINS contains invalid origin %q", origin))
		}
	}

	if c.IsProduction() {
		if c.JWTSecret == DefaultJWTSecret || len(c.JWTSecret) < 32 {
			problems = append(problems, fmt.Errorf("JWT_SECRET must be a strong non-default value in production"))
		}
		if !c.CookieSecure {
			problems = append(problems, fmt.Errorf("COOKIE_SECURE must be true in production"))
		}
	}

	return errors.Join(problems...)
}

// IsNotExist reports whether the error means "file not found".
func IsNotExist(err error) bool { return errors.Is(err, os.ErrNotExist) }
