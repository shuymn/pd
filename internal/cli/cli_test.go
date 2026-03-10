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

func TestCLI_List_InvalidDoc_StderrDiagnostic(t *testing.T) {
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

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit when diagnostics are emitted, got success")
	}

	// stdout should have the valid doc
	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 1 {
		t.Errorf("got %d valid results, want 1: %v", len(results), results)
	}

	// stderr should have the diagnostic for the invalid doc
	stderrBytes := stderr.Bytes()
	if len(stderrBytes) == 0 {
		t.Fatal("stderr is empty, want diagnostic JSON")
	}

	var diag struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}

	if err := json.Unmarshal(bytes.TrimRight(stderrBytes, "\n"), &diag); err != nil {
		t.Fatalf("unmarshal stderr: %v\nstderr: %s", err, stderrBytes)
	}

	if diag.Path == "" {
		t.Error("diag.Path is empty")
	}

	if diag.Reason == "" {
		t.Error("diag.Reason is empty")
	}
}

func TestCLI_List_DiagnosticsAreWrittenAfterStdout(t *testing.T) {
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

	if err := cmd.Run(); err == nil {
		t.Fatal("expected non-zero exit when diagnostics are emitted, got success")
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

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit when diagnostics are emitted, got success")
	}

	// No valid results
	var results []metadata.Result

	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout.String())
	}

	if len(results) != 0 {
		t.Errorf("got %d results, want 0: %v", len(results), results)
	}

	// 6 diagnostics in stderr (one per line)
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

		if err := cmd.Run(); err == nil {
			t.Fatal("expected non-zero exit when diagnostics are emitted, got success")
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

		if err := cmd.Run(); err == nil {
			t.Fatal("expected non-zero exit when diagnostics are emitted, got success")
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
