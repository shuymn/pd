package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shuymn/pd/internal/discovery"
	"github.com/shuymn/pd/internal/log"
	"github.com/shuymn/pd/internal/metadata"
)

// ListCmd implements the `pd list` command.
type ListCmd struct {
	JSON bool    `help:"Output results as JSON array."  name:"json" default:"true"`
	Kind *string `help:"Filter results by kind."        name:"kind" optional:""`
}

// Run executes the list command.
func (lc *ListCmd) Run(ctx context.Context, root *Root) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	normalizedRoot, err := normalizePath("--root", cwd, root.Root)
	if err != nil {
		return err
	}

	var kindFilter *metadata.Kind

	if lc.Kind != nil {
		var k metadata.Kind
		k, err = metadata.ParseKind(*lc.Kind)
		if err != nil {
			return fmt.Errorf("invalid kind %q: %w", *lc.Kind, err)
		}

		kindFilter = &k
	}

	s := discovery.Scanner{
		Root: filepath.Join(cwd, normalizedRoot),
	}

	results, err := s.Scan(ctx, kindFilter)
	var diagnosticErrs discovery.DiagnosticErrors
	if err != nil && !errors.As(err, &diagnosticErrs) {
		return fmt.Errorf("scan: %w", err)
	}

	if results == nil {
		results = []metadata.Result{}
	}

	log.WriteOutput(ctx, root.OutputLogger, results)

	if diagnosticErrs != nil && root.Verbose {
		for _, diagnosticErr := range diagnosticErrs {
			log.WriteDiagnostic(ctx, root.DiagnosticLogger, diagnosticErr)
		}
	}

	return nil
}
