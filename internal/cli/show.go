package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shuymn/pd/internal/discovery"
	"github.com/shuymn/pd/internal/log"
)

// ShowCmd implements the `pd show` command.
type ShowCmd struct {
	Path string `arg:"" help:"Document path relative to discovery root."`
	JSON bool   `help:"Output results as JSON object." name:"json" default:"true"`
	Body bool   `help:"Include document body in the JSON output." name:"body"`
}

// Run executes the show command.
func (sc *ShowCmd) Run(ctx context.Context, root *Root) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	normalizedRoot, err := normalizePath("--root", cwd, root.Root)
	if err != nil {
		return err
	}

	normalizedPath, err := normalizePath("<path>", filepath.Join(cwd, normalizedRoot), sc.Path)
	if err != nil {
		return err
	}

	s := discovery.Scanner{
		Root: filepath.Join(cwd, normalizedRoot),
	}

	result, err := s.Show(normalizedPath, sc.Body)
	if err != nil {
		var diagnosticErr *discovery.DiagnosticError
		if errors.As(err, &diagnosticErr) {
			log.WriteDiagnostic(ctx, root.DiagnosticLogger, diagnosticErr)
			return ErrDiagnostics
		}

		return fmt.Errorf("show document: %w", err)
	}

	log.WriteOutput(ctx, root.OutputLogger, result)

	return nil
}
