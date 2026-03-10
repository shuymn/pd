package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/shuymn/pd/internal/diagnostic"
	"github.com/shuymn/pd/internal/discovery"
	"github.com/shuymn/pd/internal/metadata"
)

// ListCmd implements the `pd list` command.
type ListCmd struct {
	JSON bool    `help:"Output results as JSON array."  name:"json" default:"true"`
	Kind *string `help:"Filter results by kind."        name:"kind" optional:""`
}

// validateRoot rejects absolute paths and paths that traverse above the base directory.
func validateRoot(root string) error {
	if filepath.IsAbs(root) {
		return fmt.Errorf("--root must be a relative path, got %q", root)
	}

	cleaned := filepath.Clean(root)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("--root must not traverse above the base directory, got %q", root)
	}

	return nil
}

// Run executes the list command.
func (lc *ListCmd) Run(ctx context.Context, root *Root) error {
	if err := validateRoot(root.Root); err != nil {
		return err
	}

	gitRoot, err := findGitRoot(ctx)
	if err != nil {
		return fmt.Errorf("find git root: %w", err)
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
		Root:   filepath.Join(gitRoot, root.Root),
		Logger: slog.New(diagnostic.NewHandler(os.Stderr)),
	}

	results, err := s.Scan(ctx, kindFilter)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	if results == nil {
		results = []metadata.Result{}
	}

	data, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}

	_, err = fmt.Fprintf(os.Stdout, "%s\n", data)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
