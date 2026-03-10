package discovery

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/shuymn/pd/internal/frontmatter"
	"github.com/shuymn/pd/internal/heading"
	"github.com/shuymn/pd/internal/metadata"
)

var discardLogger = slog.New(slog.DiscardHandler)

// Scanner walks a directory tree and extracts discovery metadata from Markdown files.
//
// NOTE: value receivers are used because all methods are read-only and the struct
// holds only a string and a pointer (cheap to copy). Switch to pointer receivers if
// a mutating method is added, the struct grows beyond ~4 fields, or it appears in
// hot copy paths.
type Scanner struct {
	Root   string
	Logger *slog.Logger
}

// Scan walks the docs directory, extracts frontmatter, validates, and applies optional kind filter.
// Valid documents are returned as Results (sorted by path).
// Invalid documents are reported via the Scanner's Logger at WARN level.
func (s Scanner) Scan(ctx context.Context, kind *metadata.Kind) ([]metadata.Result, error) {
	logger := s.Logger
	if logger == nil {
		logger = discardLogger
	}

	var results []metadata.Result

	walk := s.newWalkFunc(ctx, kind, logger, &results)

	parent := filepath.Dir(s.Root)
	dir := filepath.Base(s.Root)

	err := fs.WalkDir(os.DirFS(parent), dir, walk)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	slices.SortFunc(results, func(a, b metadata.Result) int {
		return strings.Compare(a.Path, b.Path)
	})

	return results, nil
}

func (s Scanner) newWalkFunc(
	ctx context.Context,
	kind *metadata.Kind,
	logger *slog.Logger,
	results *[]metadata.Result,
) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, walkErr error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		return s.handleFile(ctx, path, kind, logger, results)
	}
}

func (s Scanner) handleFile(
	ctx context.Context,
	path string,
	kind *metadata.Kind,
	logger *slog.Logger,
	results *[]metadata.Result,
) error {
	fullPath := filepath.Join(filepath.Dir(s.Root), path)

	result, reason, err := processFile(fullPath, path)
	if err != nil {
		return err
	}

	if reason != "" {
		logger.WarnContext(ctx, "invalid document", "path", path, "reason", reason)
		return nil
	}

	if kind == nil || result.Kind == *kind {
		*results = append(*results, *result)
	}

	return nil
}

func processFile(fullPath, relPath string) (*metadata.Result, string, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, "", err
	}

	defer func() { _ = f.Close() }()

	var meta metadata.Metadata

	body, err := frontmatter.Extract(f, &meta)
	if err != nil {
		reason := err.Error()
		if errors.Is(err, frontmatter.ErrNotFound) {
			reason = "no frontmatter found"
		}

		return nil, reason, nil
	}

	reason, err := metadata.Validate(meta)
	if err != nil {
		return nil, "", err
	}
	if reason != "" {
		return nil, reason, nil
	}

	title, ok := resolveTitle(meta.Title, body)
	if !ok {
		return nil, "missing title: no frontmatter title and no H1 heading found", nil
	}

	return &metadata.Result{
		Path:        relPath,
		Kind:        meta.Kind,
		Title:       title,
		Description: meta.Description,
	}, "", nil
}

func resolveTitle(frontmatterTitle string, body []byte) (string, bool) {
	if frontmatterTitle != "" {
		return frontmatterTitle, true
	}

	return heading.ExtractH1(body)
}
