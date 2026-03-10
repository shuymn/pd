package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shuymn/pd/internal/metadata"
	"github.com/shuymn/pd/internal/testutil"
)

// binaryPath is set by TestMain.
var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "pd-test-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tmp, "pd")

	modRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "abs path: %v\n", err)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Dir = modRoot
	build.Env = append(os.Environ(),
		"GOMODCACHE="+filepath.Join(modRoot, ".cache/go-mod"),
		"GOCACHE="+filepath.Join(modRoot, ".cache/go-build"),
	)

	if out, buildErr := build.CombinedOutput(); buildErr != nil {
		fmt.Fprintf(os.Stderr, "build binary: %s\n", out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(tmp)

	os.Exit(code)
}

func TestCLI_List_JSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/roadmap.md", `---
kind: roadmap
description: Project roadmap
title: Project Roadmap
---
# Project Roadmap

Content here.
`)
	testutil.WriteFile(t, root, "docs/adr/001-test.md", `---
kind: adr
description: Test ADR
title: Test ADR 001
---
Body.
`)

	cmd := exec.Command(binaryPath, "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2: %v", len(results), results)
	}

	// Results should be sorted by path
	if results[0].Path != "docs/adr/001-test.md" {
		t.Errorf("results[0].Path = %q, want %q", results[0].Path, "docs/adr/001-test.md")
	}

	if results[1].Path != "docs/roadmap.md" {
		t.Errorf("results[1].Path = %q, want %q", results[1].Path, "docs/roadmap.md")
	}
}

func TestCLI_List_H1Fallback(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/no-title.md", `---
kind: coding
description: A coding guide
---
# Guide Title from H1

Body content.
`)

	cmd := exec.Command(binaryPath, "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Title != "Guide Title from H1" {
		t.Errorf("Title = %q, want %q", results[0].Title, "Guide Title from H1")
	}
}

func TestCLI_List_H1Fallback_KeepsCodeSpanText(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/no-title.md", `---
kind: coding
description: A coding guide
---
# `+"`pd`"+` / Frontmatter

Body content.
`)

	results, stderr := runList(t, root, "list", "--json")
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Title != "pd / Frontmatter" {
		t.Errorf("Title = %q, want %q", results[0].Title, "pd / Frontmatter")
	}
}

func TestCLI_List_H1Fallback_KeepsCodeSpanText_Setext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	fixture := "---\nkind: coding\ndescription: A coding guide\n---\n\n" +
		"`pd` / Frontmatter\n===================\n\nBody content.\n"
	testutil.WriteFile(t, root, "docs/no-title.md", fixture)

	results, stderr := runList(t, root, "list", "--json")

	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Title != "pd / Frontmatter" {
		t.Errorf("Title = %q, want %q", results[0].Title, "pd / Frontmatter")
	}
}

func TestCLI_List_KindFilter(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/road.md", `---
kind: roadmap
description: A roadmap
title: Roadmap
---
`)
	testutil.WriteFile(t, root, "docs/adr.md", `---
kind: adr
description: An ADR
title: ADR
---
`)

	cmd := exec.Command(binaryPath, "list", "--kind", "roadmap", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --kind roadmap --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Kind != metadata.KindRoadmap {
		t.Errorf("Kind = %q, want %q", results[0].Kind, metadata.KindRoadmap)
	}
}

func TestCLI_List_InvalidDoc_IsSilentWithoutVerbose(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/valid.md", `---
kind: roadmap
description: Valid doc
title: Valid
---
`)
	testutil.WriteFile(t, root, "docs/invalid.md", `# No frontmatter

Just a heading, no YAML block.
`)

	cmd := exec.Command(binaryPath, "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 1 {
		t.Errorf("got %d valid results, want 1: %v", len(results), results)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestCLI_List_VerboseWritesDiagnosticsAfterStdout(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/valid.md", `---
kind: roadmap
description: Valid doc
title: Valid
---
`)
	testutil.WriteFile(t, root, "docs/invalid.md", `# No frontmatter

Just a heading, no YAML block.
`)

	cmd := exec.Command(binaryPath, "--verbose", "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --verbose --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 1 {
		t.Fatalf("got %d valid results, want 1: %v", len(results), results)
	}

	var diag struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(bytes.TrimRight(stderr.Bytes(), "\n"), &diag); err != nil {
		t.Fatalf("unmarshal stderr: %v\nstderr: %s", err, stderr.String())
	}
}

func TestCLI_List_AllInvalidScenarios(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// No frontmatter
	testutil.WriteFile(t, root, "docs/no-fm.md", "# Just a heading\n\nNo frontmatter.\n")
	// Missing kind
	testutil.WriteFile(t, root, "docs/missing-kind.md", "---\ndescription: no kind\ntitle: Test\n---\nBody.\n")
	// Unknown kind
	testutil.WriteFile(t, root, "docs/bad-kind.md", "---\nkind: blog\ndescription: bad\ntitle: Test\n---\nBody.\n")
	// Unknown field
	testutil.WriteFile(
		t, root,
		"docs/unknown-field.md",
		"---\nkind: roadmap\ndescription: test\nextra: not allowed\n---\nBody.\n",
	)
	// Missing description
	testutil.WriteFile(t, root, "docs/missing-desc.md", "---\nkind: roadmap\ntitle: Test\n---\nBody.\n")
	// No title and no H1
	testutil.WriteFile(t, root, "docs/no-title-no-h1.md", "---\nkind: roadmap\ndescription: test\n---\nJust content.\n")

	cmd := exec.Command(binaryPath, "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --json failed: %v\nstderr: %s", err, stderr.String())
	}

	// No valid results
	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 0 {
		t.Errorf("got %d results, want 0: %v", len(results), results)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestCLI_List_AllInvalidScenarios_VerbosePrintsDiagnostics(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/no-fm.md", "# Just a heading\n\nNo frontmatter.\n")
	testutil.WriteFile(t, root, "docs/missing-kind.md", "---\ndescription: no kind\ntitle: Test\n---\nBody.\n")
	testutil.WriteFile(t, root, "docs/bad-kind.md", "---\nkind: blog\ndescription: bad\ntitle: Test\n---\nBody.\n")
	testutil.WriteFile(
		t, root,
		"docs/unknown-field.md",
		"---\nkind: roadmap\ndescription: test\nextra: not allowed\n---\nBody.\n",
	)
	testutil.WriteFile(t, root, "docs/missing-desc.md", "---\nkind: roadmap\ntitle: Test\n---\nBody.\n")
	testutil.WriteFile(t, root, "docs/no-title-no-h1.md", "---\nkind: roadmap\ndescription: test\n---\nJust content.\n")

	cmd := exec.Command(binaryPath, "--verbose", "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --verbose --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 0 {
		t.Errorf("got %d results, want 0: %v", len(results), results)
	}

	stderrStr := stderr.String()
	lines := bytes.Split(bytes.TrimRight([]byte(stderrStr), "\n"), []byte("\n"))

	if len(lines) != 6 {
		t.Errorf("got %d stderr lines, want 6:\n%s", len(lines), stderrStr)
	}

	for i, line := range lines {
		var diag struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		}

		if err := json.Unmarshal(line, &diag); err != nil {
			t.Errorf("line %d is not valid JSON: %v\nline: %s", i, err, line)
			continue
		}

		if diag.Path == "" {
			t.Errorf("line %d: Path is empty", i)
		}

		if diag.Reason == "" {
			t.Errorf("line %d: Reason is empty", i)
		}
	}
}

func TestCLI_List_EmptyDocs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// No docs directory at all
	cmd := exec.Command(binaryPath, "list", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --json failed: %v\nstderr: %s", err, stderr.String())
	}

	// Should return empty JSON array
	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestCLI_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(binaryPath)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("pd with no args failed: %v\nstderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if out == "" {
		t.Fatal("stdout is empty, want usage output")
	}

	if !bytes.Contains(stdout.Bytes(), []byte("Usage")) {
		t.Errorf("stdout does not contain \"Usage\":\n%s", out)
	}
}

func TestCLI_List_InvalidKind(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	cmd := exec.Command(binaryPath, "list", "--kind", "nonexistent", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for invalid kind, got success")
	}
}

func TestCLI_List_RootValidation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	cases := []struct {
		name string
		path string
	}{
		{"absolute path outside cwd", "/absolute/path"},
		{"parent traversal", "../outside"},
		{"nested traversal", "foo/../../outside"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command(binaryPath, "list", "--root", tc.path, "--json") //nolint:gosec // test-controlled input
			cmd.Dir = root

			var stdout, stderr bytes.Buffer

			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected non-zero exit for --root %q, got success\nstdout: %s", tc.path, stdout.String())
			}
		})
	}
}

func TestCLI_List_AbsoluteRootWithinCWD(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/roadmap.md", `---
kind: roadmap
description: Project roadmap
title: Project Roadmap
---
`)

	cmd := exec.Command( //nolint:gosec // test-controlled input
		binaryPath,
		"list",
		"--root",
		filepath.Join(root, "docs"),
		"--json",
	)
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd list --root <abs> --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var results []metadata.Result
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Path != "roadmap.md" {
		t.Errorf("Path = %q, want %q", results[0].Path, "roadmap.md")
	}
}

func TestCLI_List_ExplicitRootChangesPathSurface(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/adr/001.md", `---
kind: adr
description: Architecture decision
title: ADR 001
---
Body.
`)
	testutil.WriteFile(t, root, "docs/adr/invalid.md", `# No frontmatter`)

	t.Run("explicit root returns root-relative success paths", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "list", "--root", "docs/adr", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd list --root docs/adr --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var results []metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}

		if results[0].Path != "001.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "001.md")
		}

		if stderr.Len() != 0 {
			t.Fatalf("stderr = %q, want empty", stderr.String())
		}
	})

	t.Run("explicit root keeps diagnostics root-relative in verbose mode", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "--verbose", "list", "--root", "docs/adr", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd list --root docs/adr --verbose --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var results []metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}

		var diag struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(bytes.TrimRight(stderr.Bytes(), "\n"), &diag); err != nil {
			t.Fatalf("unmarshal stderr: %v\nstderr: %s", err, stderr.String())
		}

		if diag.Path != "invalid.md" {
			t.Errorf("diag.Path = %q, want %q", diag.Path, "invalid.md")
		}
	})

	t.Run("omitted root keeps cwd-relative success paths", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "list", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd list --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var results []metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}

		if results[0].Path != "docs/adr/001.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "docs/adr/001.md")
		}

		if stderr.Len() != 0 {
			t.Fatalf("stderr = %q, want empty", stderr.String())
		}
	})
}

func TestCLI_List_Depth(t *testing.T) {
	t.Parallel()

	t.Run("omitted depth defaults to three", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, "docs/l0.md", `---
kind: roadmap
description: Level 0
title: Level 0
---
`)
		testutil.WriteFile(t, root, "docs/a/l1.md", `---
kind: roadmap
description: Level 1
title: Level 1
---
`)
		testutil.WriteFile(t, root, "docs/a/b/l2.md", `---
kind: roadmap
description: Level 2
title: Level 2
---
`)
		testutil.WriteFile(t, root, "docs/a/b/c/l3.md", `---
kind: roadmap
description: Level 3
title: Level 3
---
`)
		testutil.WriteFile(t, root, "docs/a/b/c/d/l4.md", `---
kind: roadmap
description: Level 4
title: Level 4
---
`)

		results, stderr := runList(t, root, "list", "--root", "docs", "--json")
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 4 {
			t.Fatalf("got %d results, want 4: %v", len(results), results)
		}

		if results[0].Path != "a/b/c/l3.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "a/b/c/l3.md")
		}

		if results[3].Path != "l0.md" {
			t.Errorf("results[3].Path = %q, want %q", results[3].Path, "l0.md")
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

		results, stderr := runList(t, root, "list", "--root", "docs", "--depth", "0", "--json")
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1: %v", len(results), results)
		}

		if results[0].Path != "root.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "root.md")
		}
	})

	t.Run("depth one returns nested level", func(t *testing.T) {
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

		results, stderr := runList(t, root, "list", "--root", "docs", "--depth", "1", "--json")
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 2 {
			t.Fatalf("got %d results, want 2: %v", len(results), results)
		}

		if results[0].Path != "root.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "root.md")
		}

		if results[1].Path != "sub/child.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "sub/child.md")
		}
	})

	t.Run("explicit subtree root makes depth relative to subtree", func(t *testing.T) {
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

		results, stderr := runList(t, root, "list", "--root", "docs/adr", "--depth", "0", "--json")
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1: %v", len(results), results)
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
		testutil.WriteFile(t, root, "docs/sub/nested-roadmap.md", `---
kind: roadmap
description: Nested roadmap
title: Nested Roadmap
---
`)
		testutil.WriteFile(t, root, "docs/sub/nested-adr.md", `---
kind: adr
description: Nested ADR
title: Nested ADR
---
`)

		results, stderr := runList(t, root, "list", "--root", "docs", "--depth", "0", "--kind", "roadmap", "--json")
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1: %v", len(results), results)
		}

		if results[0].Path != "root-roadmap.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "root-roadmap.md")
		}
	})

	t.Run("negative depth exits non-zero", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		cmd := exec.Command(binaryPath, "list", "--depth", "-1", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected non-zero exit for negative depth, got success\nstdout: %s", stdout.String())
		}
	})

	t.Run("non integer depth exits non-zero", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		cmd := exec.Command(binaryPath, "list", "--depth", "bad", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected non-zero exit for invalid depth, got success\nstdout: %s", stdout.String())
		}
	})
}

func TestCLI_List_GitIgnoreContracts(t *testing.T) {
	t.Parallel()

	t.Run("root gitignore prunes ignored markdown but keeps siblings", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()

		testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
		testutil.WriteFile(t, root, ".gitignore", "/docs/.cache/\n/docs/team/ignored.md\n")
		testutil.WriteFile(t, root, "docs/visible.md", `---
kind: roadmap
description: Visible doc
title: Visible
---
`)
		testutil.WriteFile(t, root, "docs/team/kept.md", `---
kind: roadmap
description: Kept sibling doc
title: Kept
---
`)
		testutil.WriteFile(t, root, "docs/team/ignored.md", `---
kind: roadmap
description: Ignored sibling doc
title: Ignored
---
`)
		testutil.WriteFile(t, root, "docs/.cache/ignored.md", `---
kind: roadmap
description: Ignored cache doc
title: Ignored Cache
---
`)

		results, stderr := runList(t, root, "list", "--json")

		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 2 {
			t.Fatalf("got %d results, want 2: %v", len(results), results)
		}

		if results[0].Path != "docs/team/kept.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "docs/team/kept.md")
		}

		if results[1].Path != "docs/visible.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "docs/visible.md")
		}
	})

	t.Run("git info exclude hides files from list", func(t *testing.T) {
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

		results, stderr := runList(t, root, "list", "--json")

		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1: %v", len(results), results)
		}

		if results[0].Path != "docs/visible.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "docs/visible.md")
		}
	})

	t.Run("repo root gitignore still applies to subtree scans", func(t *testing.T) {
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

		results, stderr := runList(t, root, "list", "--root", "docs/adr", "--json")

		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 1 {
			t.Fatalf("got %d results, want 1: %v", len(results), results)
		}

		if results[0].Path != "kept.md" {
			t.Errorf("Path = %q, want %q", results[0].Path, "kept.md")
		}
	})

	t.Run("gitignore still applies when depth is set", func(t *testing.T) {
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

		results, stderr := runList(t, root, "list", "--root", "docs", "--depth", "1", "--json")

		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		if len(results) != 2 {
			t.Fatalf("got %d results, want 2: %v", len(results), results)
		}

		if results[0].Path != "kept/child.md" {
			t.Errorf("results[0].Path = %q, want %q", results[0].Path, "kept/child.md")
		}

		if results[1].Path != "root.md" {
			t.Errorf("results[1].Path = %q, want %q", results[1].Path, "root.md")
		}
	})
}

func TestCLI_Show_JSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/design.md", `---
kind: design-doc
description: Design summary
title: Discovery Design
---
# Ignored Body H1

Body content.
`)

	cmd := exec.Command(binaryPath, "show", "docs/design.md", "--json")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd show --json failed: %v\nstderr: %s", err, stderr.String())
	}

	var got metadata.Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if got.Path != "docs/design.md" {
		t.Errorf("Path = %q, want %q", got.Path, "docs/design.md")
	}

	if got.Title != "Discovery Design" {
		t.Errorf("Title = %q, want %q", got.Title, "Discovery Design")
	}

	if bytes.Contains(stdout.Bytes(), []byte(`"body"`)) {
		t.Errorf("stdout unexpectedly contains body field: %s", stdout.String())
	}
}

func TestCLI_Show_Body(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	content := `---
kind: adr
description: Architecture decision
---
# Decision Title

Body content.
`
	testutil.WriteFile(t, root, "docs/adr/001.md", content)

	cmd := exec.Command(binaryPath, "show", "docs/adr/001.md", "--body")
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd show --body failed: %v\nstderr: %s", err, stderr.String())
	}

	var got metadata.ShowResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if got.Title != "Decision Title" {
		t.Errorf("Title = %q, want %q", got.Title, "Decision Title")
	}

	if got.Body != "# Decision Title\n\nBody content.\n" {
		t.Errorf("Body = %q, want %q", got.Body, "# Decision Title\n\nBody content.\n")
	}
}

func TestCLI_Show_InvalidDocument(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/invalid.md", `---
kind: roadmap
description: Invalid doc
---
No heading here.
`)

	diag := runShowExpectDiagnostic(t, root, "docs/invalid.md", "--json")
	if diag.Path != "docs/invalid.md" {
		t.Errorf("Path = %q, want %q", diag.Path, "docs/invalid.md")
	}

	if diag.Reason == "" {
		t.Error("Reason is empty")
	}
}

func TestCLI_Show_AllInvalidScenarios(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		path    string
		content string
	}{
		{
			name: "malformed frontmatter",
			path: "docs/malformed.md",
			content: `---
kind: [invalid yaml
---
Body.
`,
		},
		{
			name: "missing kind",
			path: "docs/missing-kind.md",
			content: `---
description: no kind
title: Missing Kind
---
Body.
`,
		},
		{
			name: "missing description",
			path: "docs/missing-description.md",
			content: `---
kind: roadmap
title: Missing Description
---
Body.
`,
		},
		{
			name: "unknown field",
			path: "docs/unknown-field.md",
			content: `---
kind: roadmap
description: test
extra: not allowed
---
Body.
`,
		},
		{
			name: "invalid kind",
			path: "docs/invalid-kind.md",
			content: `---
kind: blog
description: bad
title: Invalid Kind
---
Body.
`,
		},
		{
			name: "missing title fallback",
			path: "docs/no-title-no-h1.md",
			content: `---
kind: roadmap
description: missing title fallback
---
Plain paragraph only.
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			caseRoot := t.TempDir()
			testutil.WriteFile(t, caseRoot, tc.path, tc.content)

			diag := runShowExpectDiagnostic(t, caseRoot, tc.path, "--json")
			if diag.Path != tc.path {
				t.Errorf("Path = %q, want %q", diag.Path, tc.path)
			}
			if diag.Reason == "" {
				t.Fatal("Reason is empty")
			}
		})
	}
}

func TestCLI_Show_NotFound(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	diag := runShowExpectDiagnostic(t, root, "docs/missing.md", "--json")
	if diag.Path != "docs/missing.md" {
		t.Errorf("Path = %q, want %q", diag.Path, "docs/missing.md")
	}

	if diag.Reason != "document not found" {
		t.Errorf("Reason = %q, want %q", diag.Reason, "document not found")
	}
}

func TestCLI_Show_IgnoredPathExplicitlySucceeds(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, ".git/HEAD", "ref: refs/heads/main\n")
	testutil.WriteFile(t, root, ".gitignore", "/docs/ignored.md\n/docs/sub/hidden.md\n")
	testutil.WriteFile(t, root, "docs/ignored.md", `---
kind: roadmap
description: Ignored doc
title: Ignored
---
`)
	testutil.WriteFile(t, root, "docs/sub/hidden.md", `---
kind: adr
description: Hidden doc
title: Hidden
---
`)

	t.Run("default root still shows ignored file by explicit path", func(t *testing.T) {
		t.Parallel()

		var got metadata.Result
		runShowExpectSuccess(t, root, &got, "show", "docs/ignored.md", "--json")

		if got.Path != "docs/ignored.md" {
			t.Errorf("Path = %q, want %q", got.Path, "docs/ignored.md")
		}
	})

	t.Run("explicit root still shows ignored file by root relative path", func(t *testing.T) {
		t.Parallel()

		var got metadata.Result
		runShowExpectSuccess(t, root, &got, "show", "--root", "docs/sub", "hidden.md", "--json")

		if got.Path != "hidden.md" {
			t.Errorf("Path = %q, want %q", got.Path, "hidden.md")
		}
	})
}

func TestCLI_Show_DepthDoesNotAffectExplicitPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/adr/001.md", `---
kind: adr
description: Architecture decision
title: ADR 001
---
Body.
`)

	var got metadata.Result
	runShowExpectSuccess(t, root, &got, "show", "--depth", "0", "docs/adr/001.md", "--json")

	if got.Path != "docs/adr/001.md" {
		t.Errorf("Path = %q, want %q", got.Path, "docs/adr/001.md")
	}
}

func TestCLI_Show_PathValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{"absolute path outside cwd", "/absolute/path.md"},
		{"parent traversal", "../outside.md"},
		{"nested traversal", "docs/../../outside.md"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()

			cmd := exec.Command(binaryPath, "show", tc.path, "--json") //nolint:gosec // test-controlled input
			cmd.Dir = root

			var stdout, stderr bytes.Buffer

			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected non-zero exit for path %q, got success\nstdout: %s", tc.path, stdout.String())
			}
		})
	}
}

func TestCLI_Show_RootScope(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rootFlag := filepath.Join(root, "docs", "sub")

	testutil.WriteFile(t, root, "docs/sub/doc.md", `---
kind: adr
description: Nested doc
title: Nested Doc
---
Body.
`)
	testutil.WriteFile(t, root, "docs/other.md", `---
kind: adr
description: Other doc
title: Other Doc
---
Body.
`)

	t.Run("nested root success", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "show", "--root", rootFlag, "--json", "doc.md")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd show --root docs/sub doc.md --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var got metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if got.Path != "doc.md" {
			t.Errorf("Path = %q, want %q", got.Path, "doc.md")
		}
	})

	t.Run("outside root rejected", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "show", "--root", rootFlag, "--json", "../other.md")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected non-zero exit for parent traversal, got success")
		}

		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}

		if !bytes.Contains(stderr.Bytes(), []byte("must not traverse above the current working directory")) {
			t.Fatalf("stderr = %q, want traversal error", stderr.String())
		}
	})

	t.Run("repo-root-relative input rejected when root is explicit", func(t *testing.T) {
		t.Parallel()

		diag := runShowExpectDiagnostic(
			t,
			root,
			filepath.ToSlash(filepath.Join("docs", "sub", "doc.md")),
			"--root",
			rootFlag,
			"--json",
		)
		if diag.Reason != "document not found" {
			t.Errorf("Reason = %q, want %q", diag.Reason, "document not found")
		}
	})

	t.Run("absolute path within explicit root still succeeds", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(
			binaryPath,
			"show",
			"--root",
			rootFlag,
			filepath.Join(root, "docs", "sub", "doc.md"),
			"--json",
		)
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd show --root <abs> --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var got metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if got.Path != "doc.md" {
			t.Errorf("Path = %q, want %q", got.Path, "doc.md")
		}
	})
}

func TestCLI_Show_DefaultRootIsCWD(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "adr/001.md", `---
kind: adr
description: Architecture decision
title: ADR 001
---
Body.
`)

	t.Run("omitted root reads from cwd", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "show", "adr/001.md", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd show adr/001.md --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var got metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if got.Path != "adr/001.md" {
			t.Errorf("Path = %q, want %q", got.Path, "adr/001.md")
		}
	})

	t.Run("explicit dot root keeps dot-relative input", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command(binaryPath, "show", "--root", ".", "adr/001.md", "--json")
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd show --root . adr/001.md --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var got metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if got.Path != "adr/001.md" {
			t.Errorf("Path = %q, want %q", got.Path, "adr/001.md")
		}
	})

	t.Run("cwd-contained absolute path succeeds", func(t *testing.T) {
		t.Parallel()

		cmd := exec.Command( //nolint:gosec // test-controlled absolute path under temp dir
			binaryPath,
			"show",
			filepath.Join(root, "adr", "001.md"),
			"--json",
		)
		cmd.Dir = root

		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("pd show <abs in cwd> --json failed: %v\nstderr: %s", err, stderr.String())
		}

		var got metadata.Result
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
		}

		if got.Path != "adr/001.md" {
			t.Errorf("Path = %q, want %q", got.Path, "adr/001.md")
		}
	})
}

func TestCLI_Show_H1Fallback_KeepsCodeSpanText(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.WriteFile(t, root, "docs/doc.md", `---
kind: coding
description: A coding guide
---
# `+"`pd`"+` / Frontmatter

Body content.
`)

	var got metadata.Result
	runShowExpectSuccess(t, root, &got, "show", "docs/doc.md", "--json")

	if got.Title != "pd / Frontmatter" {
		t.Errorf("Title = %q, want %q", got.Title, "pd / Frontmatter")
	}
}

func TestCLI_Show_H1Fallback_KeepsCodeSpanText_Setext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	fixture := "---\nkind: coding\ndescription: A coding guide\n---\n\n" +
		"`pd` / Frontmatter\n===================\n\nBody content.\n"
	testutil.WriteFile(t, root, "docs/doc.md", fixture)

	var got metadata.Result
	runShowExpectSuccess(t, root, &got, "show", "docs/doc.md", "--json")

	if got.Title != "pd / Frontmatter" {
		t.Errorf("Title = %q, want %q", got.Title, "pd / Frontmatter")
	}
}

func runShowExpectDiagnostic(t *testing.T, root, path string, extraArgs ...string) struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
} {
	t.Helper()

	args := []string{"show"}
	args = append(args, extraArgs...)
	args = append(args, path)

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit for pd %v, got success", args)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	var diag struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(bytes.TrimRight(stderr.Bytes(), "\n"), &diag); err != nil {
		t.Fatalf("unmarshal stderr: %v\nstderr: %s", err, stderr.String())
	}

	return diag
}

func runList(t *testing.T, root string, args ...string) ([]metadata.Result, string) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd %v failed: %v\nstderr: %s", args, err, stderr.String())
	}

	var results []metadata.Result
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	return results, stderr.String()
}

func runShowExpectSuccess(t *testing.T, root string, target any, args ...string) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = root

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("pd %v failed: %v\nstderr: %s", args, err, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if err := json.Unmarshal(stdout.Bytes(), target); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}
}
