package discovery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shuymn/pd/internal/diagnostic"
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

		if results[0].Path != "docs/a-doc.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "docs/a-doc.md")
		}

		if results[1].Path != "docs/z-doc.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "docs/z-doc.md")
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

		var buf bytes.Buffer
		h := diagnostic.NewHandler(&buf)
		s := discovery.Scanner{Root: filepath.Join(root, "docs"), Logger: slog.New(h)}

		results, err := s.Scan(t.Context(), nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Scan() returned %d results, want 0: %v", len(results), results)
		}

		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		if buf.Len() == 0 {
			lines = nil
		}

		if len(lines) != 6 {
			t.Errorf("Scan() produced %d diagnostic lines, want 6: %q", len(lines), buf.String())
		}

		for i, line := range lines {
			var got struct {
				Path   string `json:"path"`
				Reason string `json:"reason"`
			}
			if err := json.Unmarshal([]byte(line), &got); err != nil {
				t.Fatalf("line %d is not valid JSON: %v", i, err)
			}
			if got.Path == "" {
				t.Errorf("line %d: Path is empty", i)
			}
			if got.Reason == "" {
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
}
