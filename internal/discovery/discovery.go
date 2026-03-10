package discovery

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/shuymn/pd/internal/frontmatter"
	"github.com/shuymn/pd/internal/heading"
	"github.com/shuymn/pd/internal/metadata"
)

// Scanner walks a directory tree and extracts discovery metadata from Markdown files.
//
// NOTE: value receivers are used because all methods are read-only and the struct
// holds only two strings (cheap to copy). Switch to pointer receivers if
// a mutating method is added, the struct grows beyond ~4 fields, or it appears in
// hot copy paths.
type Scanner struct {
	Root string
}

type document struct {
	result metadata.Result
	body   string
}

type documentDiagnosticError struct {
	reason string
}

func (de *documentDiagnosticError) Error() string {
	return de.reason
}

// DiagnosticError reports a single invalid or missing document.
type DiagnosticError struct {
	Path   string
	Reason string
}

// Error returns the diagnostic summary.
func (de *DiagnosticError) Error() string {
	return fmt.Sprintf("%s: %s", de.Path, de.Reason)
}

// DiagnosticErrors collects multiple diagnostic errors.
type DiagnosticErrors []*DiagnosticError

// Error returns the diagnostic summary.
func (de DiagnosticErrors) Error() string {
	return fmt.Sprintf("%d diagnostics", len(de))
}

// Unwrap exposes the contained diagnostics for errors.As/errors.Is traversal.
func (de DiagnosticErrors) Unwrap() []error {
	errs := make([]error, 0, len(de))
	for _, diagnosticErr := range de {
		if diagnosticErr != nil {
			errs = append(errs, diagnosticErr)
		}
	}

	return errs
}

// Scan walks the docs directory, extracts frontmatter, validates, and applies optional kind filter.
// Valid documents are returned as Results (sorted by path).
func (s Scanner) Scan(ctx context.Context, kind *metadata.Kind) ([]metadata.Result, error) {
	results := make([]metadata.Result, 0)
	diagnostics := make(DiagnosticErrors, 0)

	walk := s.newWalkFunc(ctx, kind, &results, &diagnostics)

	err := fs.WalkDir(os.DirFS(s.Root), ".", walk)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return results, nil
		}

		return nil, err
	}

	slices.SortFunc(results, func(a, b metadata.Result) int {
		return strings.Compare(a.Path, b.Path)
	})

	if len(diagnostics) > 0 {
		return results, diagnostics
	}

	return results, nil
}

// Show reads a single document, validates its discovery metadata, and optionally includes the body.
func (s Scanner) Show(
	path string,
	includeBody bool,
) (*metadata.ShowResult, error) {
	rootPath, err := s.showFullPath(path)
	if err != nil {
		var diagnosticErr *documentDiagnosticError
		if errors.As(err, &diagnosticErr) {
			return nil, &DiagnosticError{Path: path, Reason: diagnosticErr.reason}
		}

		return nil, err
	}

	doc, err := readDocument(rootPath, path, includeBody)
	if err != nil {
		var diagnosticErr *documentDiagnosticError
		if errors.As(err, &diagnosticErr) {
			return nil, &DiagnosticError{Path: path, Reason: diagnosticErr.reason}
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		return nil, &DiagnosticError{Path: path, Reason: "document not found"}
	}

	return &metadata.ShowResult{
		Result: doc.result,
		Body:   doc.body,
	}, nil
}

func (s Scanner) showFullPath(path string) (string, error) {
	fullPath := filepath.Join(s.Root, path)
	relToRoot, err := filepath.Rel(s.Root, fullPath)
	if err != nil {
		return "", err
	}

	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return "", &documentDiagnosticError{reason: "path is outside discovery root"}
	}

	return fullPath, nil
}

func (s Scanner) newWalkFunc(
	ctx context.Context,
	kind *metadata.Kind,
	results *[]metadata.Result,
	diagnostics *DiagnosticErrors,
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

		return s.handleFile(path, kind, results, diagnostics)
	}
}

func (s Scanner) handleFile(
	path string,
	kind *metadata.Kind,
	results *[]metadata.Result,
	diagnostics *DiagnosticErrors,
) error {
	fullPath := filepath.Join(s.Root, path)

	doc, err := readDocument(fullPath, path, false)
	if err != nil {
		var diagnosticErr *documentDiagnosticError
		if errors.As(err, &diagnosticErr) {
			*diagnostics = append(*diagnostics, &DiagnosticError{Path: path, Reason: diagnosticErr.reason})
			return nil
		}

		return err
	}

	if kind == nil || doc.result.Kind == *kind {
		*results = append(*results, doc.result)
	}

	return nil
}

func readDocument(fullPath, relPath string, needBody bool) (*document, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	var meta metadata.Metadata

	body, err := frontmatter.Extract(f, &meta)
	if err != nil {
		reason := err.Error()
		if errors.Is(err, frontmatter.ErrNotFound) {
			reason = "no frontmatter found"
		}

		return nil, &documentDiagnosticError{reason: reason}
	}

	reason, err := metadata.Validate(meta)
	if err != nil {
		return nil, err
	}
	if reason != "" {
		return nil, &documentDiagnosticError{reason: reason}
	}

	title, ok := resolveTitle(meta.Title, body)
	if !ok {
		return nil, &documentDiagnosticError{reason: "missing title: no frontmatter title and no H1 heading found"}
	}

	var bodyStr string
	if needBody {
		bodyStr = string(body)
	}

	return &document{
		result: metadata.Result{
			Path:        relPath,
			Kind:        meta.Kind,
			Title:       title,
			Description: meta.Description,
		},
		body: bodyStr,
	}, nil
}

func resolveTitle(frontmatterTitle string, body []byte) (string, bool) {
	if frontmatterTitle != "" {
		return frontmatterTitle, true
	}

	return heading.ExtractH1(body)
}
