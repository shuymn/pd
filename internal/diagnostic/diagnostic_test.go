package diagnostic_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/shuymn/pd/internal/diagnostic"
)

func TestHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("warn produces jsonl with path and reason", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		h := diagnostic.NewHandler(&buf)
		logger := slog.New(h)

		logger.Warn("msg", "path", "docs/x.md", "reason", "error text")

		output := buf.String()
		if output == "" {
			t.Fatal("Handle() produced no output")
		}

		if output[len(output)-1] != '\n' {
			t.Error("Handle() output does not end with newline")
		}

		var got struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &got); err != nil {
			t.Fatalf("Handle() output is not valid JSON: %v", err)
		}

		if got.Path != "docs/x.md" {
			t.Errorf("Path = %q, want %q", got.Path, "docs/x.md")
		}

		if got.Reason != "error text" {
			t.Errorf("Reason = %q, want %q", got.Reason, "error text")
		}
	})

	t.Run("multiple calls produce jsonl", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		h := diagnostic.NewHandler(&buf)
		logger := slog.New(h)

		logger.Warn("msg", "path", "docs/a.md", "reason", "reason A")
		logger.Warn("msg", "path", "docs/b.md", "reason", "reason B")

		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
		}

		for i, line := range lines {
			var got struct {
				Path   string `json:"path"`
				Reason string `json:"reason"`
			}
			if err := json.Unmarshal([]byte(line), &got); err != nil {
				t.Fatalf("line %d is not valid JSON: %v", i, err)
			}
			if got.Path == "" {
				t.Errorf("line %d: Path is empty", i)
			}
			if got.Reason == "" {
				t.Errorf("line %d: Reason is empty", i)
			}
		}
	})

	t.Run("info level is not output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		h := diagnostic.NewHandler(&buf)
		logger := slog.New(h)

		logger.Info("should be ignored", "path", "docs/x.md", "reason", "whatever")

		if buf.Len() != 0 {
			t.Errorf("expected no output for Info, got %q", buf.String())
		}
	})
}
