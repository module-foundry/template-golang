package redisx

import (
	"context"
	"testing"
)

func TestHealthcheckNilClient(t *testing.T) {
	if err := Healthcheck(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil client")
	}
}
