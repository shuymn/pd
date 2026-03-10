package log

import (
	"context"
	"io"
	"log/slog"

	"github.com/shuymn/pd/internal/diagnostic"
	"github.com/shuymn/pd/internal/discovery"
	"github.com/shuymn/pd/internal/output"
)

// NewDiagnosticLogger returns a logger configured for diagnostic JSONL output.
func NewDiagnosticLogger(w io.Writer) *slog.Logger {
	return slog.New(diagnostic.NewHandler(w))
}

// NewOutputLogger returns a logger configured for success JSONL output.
func NewOutputLogger(w io.Writer) *slog.Logger {
	return slog.New(output.NewHandler(w))
}

// WriteOutput writes a single payload through the logger contract.
func WriteOutput(ctx context.Context, logger *slog.Logger, value any) {
	logger.LogAttrs(ctx, slog.LevelInfo, "", slog.Any("value", value))
}

// WriteDiagnostic writes a single diagnostic through the diagnostic logger contract.
func WriteDiagnostic(ctx context.Context, logger *slog.Logger, diagnosticErr *discovery.DiagnosticError) {
	logger.WarnContext(ctx, "invalid document", "path", diagnosticErr.Path, "reason", diagnosticErr.Reason)
}
