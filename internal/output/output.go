package output

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
)

// Handler is a slog.Handler that writes output values as JSONL.
type Handler struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// NewHandler creates a new Handler that writes to w.
func NewHandler(w io.Writer) *Handler {
	return &Handler{
		enc: json.NewEncoder(w),
	}
}

// Enabled reports whether the handler handles the given level.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

// Handle writes the "value" attribute as a single JSON line.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	var value any

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "value" {
			value = a.Value.Any()
			return false
		}

		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.enc.Encode(value)
}

// WithAttrs returns h unchanged; attribute pre-population is not supported.
func (h *Handler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns h unchanged; grouping is not supported.
func (h *Handler) WithGroup(_ string) slog.Handler {
	return h
}
