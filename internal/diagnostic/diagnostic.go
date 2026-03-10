package diagnostic

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
)

// Handler is a slog.Handler that writes diagnostic records as JSONL to w.
// Only records at WARN level or above are written.
// Each record is written as {"path":"...","reason":"..."}.
type Handler struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// NewHandler creates a new Handler that writes to w.
func NewHandler(w io.Writer) *Handler {
	return &Handler{
		mu:  sync.Mutex{},
		enc: json.NewEncoder(w),
	}
}

// Enabled reports whether the handler handles records at the given level.
// Only WARN and above are handled.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelWarn
}

// Handle writes the record as a single JSON line {"path":"...","reason":"..."}.
// It extracts "path" and "reason" from the record's attributes.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	var path, reason string

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "path":
			path = a.Value.String()
		case "reason":
			reason = a.Value.String()
		}
		return path == "" || reason == ""
	})

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.enc.Encode(struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}{
		Path:   path,
		Reason: reason,
	})
}

// WithAttrs returns h unchanged; attribute pre-population is not supported.
func (h *Handler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns h unchanged; grouping is not supported.
func (h *Handler) WithGroup(_ string) slog.Handler {
	return h
}
