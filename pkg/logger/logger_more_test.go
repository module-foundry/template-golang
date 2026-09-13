package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
)

type fakeReporter struct{ calls int }

func (f *fakeReporter) Report(context.Context, error) { f.calls++ }

func TestNoopReporter(t *testing.T) {
	rep := NoopReporter()
	rep.Report(context.Background(), errors.New("x"))
}

func TestReporterInterface(t *testing.T) {
	var rep Reporter = &fakeReporter{}
	rep.Report(context.Background(), errors.New("x"))
}

func TestNewExplicitJSON(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "debug", Format: "json", Env: "development", Output: &buf})
	log.Warn("careful")
	if !bytes.Contains(buf.Bytes(), []byte(`"level":"WARN"`)) {
		t.Fatalf("unexpected log: %s", buf.String())
	}
}

func TestNewExplicitText(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "info", Format: "text", Env: "production", Output: &buf})
	log.Info("plain")
	if !bytes.Contains(buf.Bytes(), []byte("plain")) {
		t.Fatalf("unexpected log: %s", buf.String())
	}
}

func TestFromCtxWithoutLogger(t *testing.T) {
	if FromCtx(context.Background()) == nil {
		t.Fatal("must return default logger")
	}
	var nilCtx context.Context
	if FromCtx(nilCtx) == nil {
		t.Fatal("nil context must be handled")
	}
}

func TestWithContextOverrides(t *testing.T) {
	var buf bytes.Buffer
	custom := slog.New(slog.NewTextHandler(&buf, nil))
	ctx := WithContext(context.Background(), custom)
	FromCtx(ctx).Info("from custom")
	if !bytes.Contains(buf.Bytes(), []byte("from custom")) {
		t.Fatalf("custom logger was not used: %s", buf.String())
	}
}
