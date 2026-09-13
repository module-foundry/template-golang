package postgres

import (
	"context"
	"testing"
)

func TestNewInvalidDSN(t *testing.T) {
	if _, err := New(context.Background(), Config{DSN: "://not-a-dsn"}); err == nil {
		t.Fatal("expected error for invalid dsn")
	}
}

func TestHealthcheckNilPool(t *testing.T) {
	if err := Healthcheck(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil pool")
	}
}
