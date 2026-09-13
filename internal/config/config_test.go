package config

import (
	"strings"
	"testing"
	"time"
)

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("APP_ENV", EnvDevelopment)
	t.Setenv("COOKIE_SECURE", "false")
}

func TestLoadDefaults(t *testing.T) {
	setRequired(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.HTTPPort != 3001 {
		t.Fatalf("http port = %d", cfg.HTTPPort)
	}
	if cfg.JWTTTL != 24*time.Hour {
		t.Fatalf("jwt ttl = %s", cfg.JWTTTL)
	}
	if !cfg.DBAutoMigrate {
		t.Fatal("auto migrate must default to true")
	}
	if !cfg.ErrorTraceEnabled {
		t.Fatal("error trace must default to true")
	}
}

func TestCORSOriginsParsing(t *testing.T) {
	setRequired(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", " http://a.example ,https://b.example,,http://a.example ")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	origins := cfg.CORSOrigins()
	want := []string{"http://a.example", "https://b.example"}
	if len(origins) != len(want) {
		t.Fatalf("origins = %v", origins)
	}
	for i := range want {
		if origins[i] != want[i] {
			t.Fatalf("origins = %v", origins)
		}
	}
}

func TestValidateProductionInvariants(t *testing.T) {
	setRequired(t)
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("JWT_SECRET", DefaultJWTSecret)

	_, err := Load()
	if err == nil {
		t.Fatal("expected production validation error")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") || !strings.Contains(err.Error(), "COOKIE_SECURE") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateInvalidOrigin(t *testing.T) {
	setRequired(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "not-a-url")
	if _, err := Load(); err == nil {
		t.Fatal("expected origin validation error")
	}
}

func TestValidateWildcardForbidden(t *testing.T) {
	setRequired(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")
	if _, err := Load(); err == nil {
		t.Fatal("wildcard origin must be rejected")
	}
}

func TestValidateMaxConns(t *testing.T) {
	setRequired(t)
	t.Setenv("POSTGRES_MAX_CONNS", "1")
	t.Setenv("POSTGRES_MIN_CONNS", "5")
	if _, err := Load(); err == nil {
		t.Fatal("expected max/min conns validation error")
	}
}
