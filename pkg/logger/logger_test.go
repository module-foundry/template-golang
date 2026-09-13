package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"template-golang/pkg/jsonx"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"other": slog.LevelInfo,
		"":      slog.LevelInfo,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNewJSONInProduction(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "info", Format: "auto", Env: "production", Service: "svc", Version: "1", Output: &buf})
	log.Info("hello", slog.String("k", "v"))

	var entry map[string]any
	if err := jsonx.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not valid json: %v (%s)", err, buf.String())
	}
	if entry["msg"] != "hello" || entry["service"] != "svc" {
		t.Fatalf("unexpected entry: %v", entry)
	}
}

func TestNewTextInDevelopment(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "debug", Format: "auto", Env: "development", Output: &buf})
	log.Debug("dev line")
	if !strings.Contains(buf.String(), "dev line") {
		t.Fatalf("missing log line: %s", buf.String())
	}
}

func TestFromCtxFallback(t *testing.T) {
	if FromCtx(nil) == nil {
		t.Fatal("nil context must fall back to default logger")
	}
	log := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	ctx := WithContext(context.Background(), log)
	if FromCtx(ctx) != log {
		t.Fatal("context logger must be returned")
	}
}
