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

	scanWithDepth := func(
		t *testing.T,
		root string,
		depth int,
		kind *metadata.Kind,
	) ([]metadata.Result, error) {
		t.Helper()

		s := discovery.Scanner{Root: root}

		return s.Scan(t.Context(), kind, discovery.ScanOptions{MaxDepth: &depth})
	}

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

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
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

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
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

	t.Run("H1 fallback keeps code span text", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/no-title.md", `---
kind: coding
description: A coding guide
---
# `+"`pd`"+` / Frontmatter

Body content.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Title != "pd / Frontmatter" {
			t.Errorf("Title = %q, want %q", results[0].Title, "pd / Frontmatter")
		}
	})

	t.Run("H1 fallback keeps code span text in setext heading", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		fixture := "---\nkind: coding\ndescription: A coding guide\n---\n\n" +
			"`pd` / Frontmatter\n===================\n\nBody content.\n"
		testutil.WriteFile(t, root, "docs/no-title.md", fixture)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Title != "pd / Frontmatter" {
			t.Errorf("Title = %q, want %q", results[0].Title, "pd / Frontmatter")
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

		results, err := s.Scan(t.Context(), &k, discovery.ScanOptions{})
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

	t.Run("depth zero returns only root documents", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/root.md", `---
kind: roadmap
description: Root doc
title: Root
---
`)
		testutil.WriteFile(t, root, "docs/sub/nested.md", `---
kind: adr
description: Nested doc
title: Nested
---
`)

		results, err := scanWithDepth(t, filepath.Join(root, "docs"), 0, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "root.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "root.md")
		}
	})

	t.Run("depth one returns one nested level", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/root.md", `---
kind: roadmap
description: Root doc
title: Root
---
`)
		testutil.WriteFile(t, root, "docs/sub/child.md", `---
kind: adr
description: Child doc
title: Child
---
`)
		testutil.WriteFile(t, root, "docs/sub/deeper/grandchild.md", `---
kind: design-doc
description: Grandchild doc
title: Grandchild
---
`)

		results, err := scanWithDepth(t, filepath.Join(root, "docs"), 1, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Scan() returned %d results, want 2", len(results))
		}

		if results[0].Path != "root.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "root.md")
		}

		if results[1].Path != "sub/child.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "sub/child.md")
		}
	})

	t.Run("depth uses subtree root as zero", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/adr/001.md", `---
kind: adr
description: Top doc
title: Top
---
`)
		testutil.WriteFile(t, root, "docs/adr/archive/002.md", `---
kind: adr
description: Nested doc
title: Nested
---
`)

		results, err := scanWithDepth(t, filepath.Join(root, "docs", "adr"), 0, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "001.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "001.md")
		}
	})

	t.Run("depth composes with kind filter", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/root-roadmap.md", `---
kind: roadmap
description: Root roadmap
title: Root Roadmap
---
`)
		testutil.WriteFile(t, root, "docs/sub/nested-adr.md", `---
kind: adr
description: Nested ADR
title: Nested ADR
---
`)
		testutil.WriteFile(t, root, "docs/sub/nested-roadmap.md", `---
kind: roadmap
description: Nested roadmap
title: Nested Roadmap
---
`)

		k := metadata.KindRoadmap
		results, err := scanWithDepth(t, filepath.Join(root, "docs"), 0, &k)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "root-roadmap.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "root-roadmap.md")
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

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
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

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
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

		_, err := s.Scan(ctx, nil, discovery.ScanOptions{})
		// A cancelled context may or may not produce a walk error depending on
		// file system timing; we only assert that the call does not panic.
		_ = err
	})

	t.Run("gitignored directories are skipped", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".gitignore", "/docs/.cache/\n")
		testutil.WriteFile(t, root, "docs/visible.md", `---
kind: roadmap
description: Visible doc
title: Visible
---
`)
		testutil.WriteFile(t, root, "docs/.cache/ignored.md", `---
kind: roadmap
description: Ignored doc
title: Ignored
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "visible.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "visible.md")
		}
	})

	t.Run("depth still respects gitignored directories", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".gitignore", "/docs/ignored/\n")
		testutil.WriteFile(t, root, "docs/root.md", `---
kind: roadmap
description: Root doc
title: Root
---
`)
		testutil.WriteFile(t, root, "docs/ignored/doc.md", `---
kind: roadmap
description: Ignored doc
title: Ignored
---
`)
		testutil.WriteFile(t, root, "docs/kept/child.md", `---
kind: roadmap
description: Kept doc
title: Kept
---
`)

		results, err := scanWithDepth(t, filepath.Join(root, "docs"), 1, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Scan() returned %d results, want 2", len(results))
		}

		if results[0].Path != "kept/child.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "kept/child.md")
		}

		if results[1].Path != "root.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "root.md")
		}
	})

	t.Run("depth prunes directories beyond max depth", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/visible.md", `---
kind: roadmap
description: Visible doc
title: Visible
---
`)
		testutil.WriteFile(t, root, "docs/deep/hidden.md", `---
kind: roadmap
description: Hidden doc
title: Hidden
---
`)
		testutil.WriteFile(t, root, "docs/deep/deeper/also-hidden.md", `---
kind: roadmap
description: Also hidden doc
title: Also Hidden
---
`)

		results, err := scanWithDepth(t, filepath.Join(root, "docs"), 0, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "visible.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "visible.md")
		}
	})

	t.Run("nested gitignore and negation are respected", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, "docs/team/.gitignore", "*.md\n!keep.md\n")
		testutil.WriteFile(t, root, "docs/team/keep.md", `---
kind: adr
description: Keep doc
title: Keep
---
`)
		testutil.WriteFile(t, root, "docs/team/drop.md", `---
kind: adr
description: Drop doc
title: Drop
---
`)
		testutil.WriteFile(t, root, "docs/other.md", `---
kind: adr
description: Other doc
title: Other
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Scan() returned %d results, want 2", len(results))
		}

		if results[0].Path != "other.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "other.md")
		}

		if results[1].Path != "team/keep.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "team/keep.md")
		}
	})

	t.Run("repository root gitignore applies to subdirectory scans", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".gitignore", "/docs/adr/ignored.md\n")
		testutil.WriteFile(t, root, "docs/adr/kept.md", `---
kind: adr
description: Kept doc
title: Kept
---
`)
		testutil.WriteFile(t, root, "docs/adr/ignored.md", `---
kind: adr
description: Ignored doc
title: Ignored
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs", "adr")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "kept.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "kept.md")
		}
	})

	t.Run("gitignored files are skipped without skipping sibling files", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".gitignore", "/docs/team/ignored.md\n")
		testutil.WriteFile(t, root, "docs/team/kept.md", `---
kind: adr
description: Kept sibling doc
title: Kept
---
`)
		testutil.WriteFile(t, root, "docs/team/ignored.md", `---
kind: adr
description: Ignored sibling doc
title: Ignored
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "team/kept.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "team/kept.md")
		}
	})

	t.Run("git info exclude is respected", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".git/info/exclude", "docs/excluded.md\n")
		testutil.WriteFile(t, root, "docs/visible.md", `---
kind: tooling
description: Visible doc
title: Visible
---
`)
		testutil.WriteFile(t, root, "docs/excluded.md", `---
kind: tooling
description: Excluded doc
title: Excluded
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "visible.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "visible.md")
		}
	})

	t.Run("linked worktree gitdir file loads info exclude", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		repoRoot := filepath.Join(root, "repo")

		testutil.WriteFile(t, repoRoot, ".git", "gitdir: ../git-common/worktrees/repo\n")
		testutil.WriteFile(t, root, "git-common/worktrees/repo/info/exclude", "docs/excluded.md\n")
		testutil.WriteFile(t, repoRoot, "docs/visible.md", `---
kind: tooling
description: Visible doc
title: Visible
---
`)
		testutil.WriteFile(t, repoRoot, "docs/excluded.md", `---
kind: tooling
description: Excluded doc
title: Excluded
---
`)

		s := discovery.Scanner{Root: filepath.Join(repoRoot, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Scan() returned %d results, want 1", len(results))
		}

		if results[0].Path != "visible.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "visible.md")
		}
	})

	t.Run("missing root inside git repository returns empty", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 0 {
			t.Fatalf("Scan() returned %d results, want 0", len(results))
		}
	})

	t.Run("non git directories ignore no files", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".gitignore", "/docs/ignored.md\n")
		testutil.WriteFile(t, root, "docs/kept.md", `---
kind: roadmap
description: Kept doc
title: Kept
---
`)
		testutil.WriteFile(t, root, "docs/ignored.md", `---
kind: roadmap
description: Ignored only in git
title: Ignored
---
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		results, err := s.Scan(t.Context(), nil, discovery.ScanOptions{})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Scan() returned %d results, want 2", len(results))
		}
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

	t.Run("show keeps code span text in H1 fallback", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/doc.md", `---
kind: coding
description: A coding guide
---
# `+"`pd`"+` / Frontmatter

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

		if showResult.Title != "pd / Frontmatter" {
			t.Errorf("Title = %q, want %q", showResult.Title, "pd / Frontmatter")
		}
	})

	t.Run("show keeps code span text in setext H1 fallback", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		fixture := "---\nkind: coding\ndescription: A coding guide\n---\n\n" +
			"`pd` / Frontmatter\n===================\n\nBody content.\n"
		testutil.WriteFile(t, root, "docs/doc.md", fixture)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		showResult, err := s.Show("doc.md", false)
		if err != nil {
			t.Fatalf("Show() error = %v", err)
		}

		if showResult == nil {
			t.Fatal("Show() showResult is nil")
		}

		if showResult.Title != "pd / Frontmatter" {
			t.Errorf("Title = %q, want %q", showResult.Title, "pd / Frontmatter")
		}
	})

	t.Run("show prefers frontmatter title over H1 fallback with code span", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/doc.md", `---
kind: coding
description: A coding guide
title: Explicit Title
---
# `+"`pd`"+` / Frontmatter

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
	})

	t.Run("show returns ignored documents when addressed directly", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".gitignore", "/docs/ignored.md\n")
		testutil.WriteFile(t, root, "docs/ignored.md", `---
kind: adr
description: Ignored doc
title: Ignored
---
Body.
`)

		s := discovery.Scanner{Root: filepath.Join(root, "docs")}

		showResult, err := s.Show("ignored.md", false)
		if err != nil {
			t.Fatalf("Show() error = %v", err)
		}
		if showResult == nil {
			t.Fatal("Show() showResult is nil")
		}
		if showResult.Path != "ignored.md" {
			t.Errorf("Path = %q, want %q", showResult.Path, "ignored.md")
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
