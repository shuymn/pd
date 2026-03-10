package output_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/shuymn/pd/internal/output"
)

func TestHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("info produces jsonl payload", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		h := output.NewHandler(&buf)
		logger := slog.New(h)

		logger.Info("msg", "value", map[string]string{"kind": "adr"})

		output := buf.String()
		if output == "" {
			t.Fatal("Handle() produced no output")
		}

		var got map[string]string
		if err := json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &got); err != nil {
			t.Fatalf("Handle() output is not valid JSON: %v", err)
		}

		if got["kind"] != "adr" {
			t.Errorf("kind = %q, want %q", got["kind"], "adr")
		}
	})

	t.Run("multiple calls produce jsonl", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		h := output.NewHandler(&buf)
		logger := slog.New(h)

		logger.Info("msg", "value", map[string]string{"path": "a"})
		logger.Info("msg", "value", map[string]string{"path": "b"})

		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
		}
	})
}
