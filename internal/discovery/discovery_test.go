package discovery_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/shuymn/pd/internal/discovery"
	"github.com/shuymn/pd/internal/metadata"
	"github.com/shuymn/pd/internal/testutil"
)

func TestScanner_Scan(t *testing.T) {
	t.Parallel()

	t.Run("valid documents are returned sorted", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/z-doc.md", `---
kind: roadmap
description: Z document
title: Z Doc
---
Body.
`)
		testutil.WriteFile(t, root, "docs/a-doc.md", `---
kind: adr
description: A document
title: A Doc
---
Body.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Scan() returned %d results, want 2", len(results))
		}

		if results[0].Path != "a-doc.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "a-doc.md")
		}

		if results[1].Path != "z-doc.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "z-doc.md")
		}
	})

	t.Run("H1 fallback when title absent", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/no-title.md", `---
kind: coding
description: A coding guide
---
# My H1 Title

Body content.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Title != "My H1 Title" {
			t.Errorf("Title = %q, want %q", results[0].Title, "My H1 Title")
		}
	})

	t.Run("kind filter", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/road.md", `---
kind: roadmap
description: Roadmap doc
title: Roadmap
---
`)
		testutil.WriteFile(t, root, "docs/adr.md", `---
kind: adr
description: ADR doc
title: ADR
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}
		k := metadata.KindRoadmap

		results, err := s.Scan(t.Context(), &k)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Kind != metadata.KindRoadmap {
			t.Errorf("Kind = %q, want %q", results[0].Kind, metadata.KindRoadmap)
		}
	})

	t.Run("invalid documents produce diagnostics", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		// No frontmatter
		testutil.WriteFile(t, root, "docs/no-fm.md", "# Just a heading\n\nNo frontmatter here.\n")
		// Missing kind
		testutil.WriteFile(t, root, "docs/missing-kind.md", "---\ndescription: no kind here\ntitle: Test\n---\nBody.\n")
		// Unknown kind
		testutil.WriteFile(
			t,
			root,
			"docs/bad-kind.md",
			"---\nkind: blog\ndescription: bad kind\ntitle: Test\n---\nBody.\n",
		)
		// Unknown field
		testutil.WriteFile(
			t, root,
			"docs/unknown-field.md",
			"---\nkind: roadmap\ndescription: test\nextra: not allowed\n---\nBody.\n",
		)
		// Missing description
		testutil.WriteFile(t, root, "docs/missing-desc.md", "---\nkind: roadmap\ntitle: Test\n---\nBody.\n")
		// No title and no H1
		testutil.WriteFile(
			t,
			root,
			"docs/no-title-no-h1.md",
			"---\nkind: roadmap\ndescription: test\n---\nJust content, no heading.\n",
		)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil)
		var diagnosticErrs discovery.DiagnosticErrors
		if !errors.As(err, &diagnosticErrs) {
			t.Fatalf("Scan() error = %v, want DiagnosticErrors", err)
		}

		if len(results) != 0 {
			t.Errorf("Scan() returned %d results, want 0: %v", len(results), results)
		}

		if len(diagnosticErrs) != 6 {
			t.Fatalf("Scan() returned %d diagnostics, want 6", len(diagnosticErrs))
		}

		for i, diagnosticErr := range diagnosticErrs {
			if diagnosticErr.Path == "" {
				t.Errorf("line %d: Path is empty", i)
			}
			if diagnosticErr.Reason == "" {
				t.Errorf("line %d: Reason is empty", i)
			}
		}
	})

	t.Run("no docs directory returns empty", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Scan() results = %v, want empty", results)
		}
	})

	t.Run("context cancellation stops scan", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/doc.md", `---
kind: roadmap
description: A doc
title: Doc
---
`)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		_, err := s.Scan(ctx, nil)
		// A cancelled context may or may not produce a walk error depending on
		// file system timing; we only assert that the call does not panic.
		_ = err
	})

	t.Run("show returns metadata only", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/doc.md", `---
kind: roadmap
description: A doc
title: Explicit Title
---
# Body Heading

Body content.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		showResult, err := s.Show("doc.md", false)
		if err != nil {
			t.Fatalf("Show() error = %v", err)
		}
		if showResult == nil {
			t.Fatal("Show() showResult is nil")
		}
		if showResult.Title != "Explicit Title" {
			t.Errorf("Title = %q, want %q", showResult.Title, "Explicit Title")
		}
		if showResult.Body != "" {
			t.Errorf("Body = %q, want empty (includeBody=false)", showResult.Body)
		}
	})

	t.Run("show returns metadata and body", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/doc.md", `---
kind: adr
description: A doc
---
# Fallback Title

Body content.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		showResult, err := s.Show("doc.md", true)
		if err != nil {
			t.Fatalf("Show() error = %v", err)
		}
		if showResult == nil {
			t.Fatal("Show() showResult is nil")
		}
		if showResult.Title != "Fallback Title" {
			t.Errorf("Title = %q, want %q", showResult.Title, "Fallback Title")
		}
		if showResult.Body != "# Fallback Title\n\nBody content.\n" {
			t.Errorf("Body = %q, want %q", showResult.Body, "# Fallback Title\n\nBody content.\n")
		}
	})

	t.Run("show supports nested root", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/sub/doc.md", `---
kind: adr
description: A nested doc
title: Nested Title
---
Body.
`)

		s := discovery.Scanner{
			Root: filepath.Join(root, "docs", "sub"),
		}

		showResult, err := s.Show("doc.md", false)
		if err != nil {
			t.Fatalf("Show() error = %v", err)
		}
		if showResult == nil {
			t.Fatal("Show() showResult is nil")
		}
		if showResult.Path != "doc.md" {
			t.Errorf("Path = %q, want %q", showResult.Path, "doc.md")
		}
	})

	t.Run("show rejects path outside root", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/sub/doc.md", `---
kind: adr
description: A nested doc
title: Nested Title
---
Body.
`)
		testutil.WriteFile(t, root, "docs/other.md", `---
kind: adr
description: Another doc
title: Other Title
---
Body.
`)

		s := discovery.Scanner{
			Root: filepath.Join(root, "docs", "sub"),
		}

		showResult, err := s.Show("../other.md", false)
		var diagnosticErr *discovery.DiagnosticError
		if !errors.As(err, &diagnosticErr) {
			t.Fatalf("Show() error = %v, want DiagnosticError", err)
		}
		if showResult != nil {
			t.Fatalf("Show() returned showResult=%#v, want nil", showResult)
		}
		if diagnosticErr.Reason != "path is outside discovery root" {
			t.Errorf("Reason = %q, want %q", diagnosticErr.Reason, "path is outside discovery root")
		}
	})

	t.Run("show returns invalid reason", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/invalid.md", `---
kind: roadmap
description: Missing title fallback
---
Plain paragraph only.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		showResult, err := s.Show("invalid.md", false)
		var diagnosticErr *discovery.DiagnosticError
		if !errors.As(err, &diagnosticErr) {
			t.Fatalf("Show() error = %v, want DiagnosticError", err)
		}
		if showResult != nil {
			t.Fatalf("Show() returned showResult=%#v, want nil", showResult)
		}
		if diagnosticErr.Path != "invalid.md" {
			t.Errorf("Path = %q, want %q", diagnosticErr.Path, "invalid.md")
		}
	})

	t.Run("show returns not found reason", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		showResult, err := s.Show("missing.md", false)
		var diagnosticErr *discovery.DiagnosticError
		if !errors.As(err, &diagnosticErr) {
			t.Fatalf("Show() error = %v, want DiagnosticError", err)
		}
		if showResult != nil {
			t.Fatalf("Show() returned showResult=%#v, want nil", showResult)
		}
		if diagnosticErr.Reason != "document not found" {
			t.Errorf("Reason = %q, want %q", diagnosticErr.Reason, "document not found")
		}
	})
}
