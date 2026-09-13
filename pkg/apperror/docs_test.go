package apperror

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestErrorCodesDocsAreSynced keeps docs/error-codes/{en,ru}.md in lockstep
// with the registry: every registered code must be documented in both languages.
func TestErrorCodesDocsAreSynced(t *testing.T) {
	codes := Codes()
	for _, lang := range []string{"en", "ru"} {
		path := filepath.Join("..", "..", "docs", "error-codes", lang+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(data)
		for code := range codes {
			if !strings.Contains(content, string(code)) {
				t.Errorf("%s: code %s is not documented", path, code)
			}
		}
	}
}
