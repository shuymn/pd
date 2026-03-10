package heading_test

import (
	"testing"

	"github.com/shuymn/pd/internal/heading"
)

func TestExtractH1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		body      []byte
		wantTitle string
		wantFound bool
	}{
		{
			name:      "ATX h1",
			body:      []byte("# My Title\n\nSome content."),
			wantTitle: "My Title",
			wantFound: true,
		},
		{
			name:      "Setext h1",
			body:      []byte("My Title\n========\n\nSome content."),
			wantTitle: "My Title",
			wantFound: true,
		},
		{
			name:      "first of multiple headings",
			body:      []byte("# First\n\n## Second\n\n# Third"),
			wantTitle: "First",
			wantFound: true,
		},
		{
			name:      "h2 only",
			body:      []byte("## Not H1\n\nContent."),
			wantTitle: "",
			wantFound: false,
		},
		{
			name:      "no heading",
			body:      []byte("Just some plain text."),
			wantTitle: "",
			wantFound: false,
		},
		{
			name:      "empty body",
			body:      []byte{},
			wantTitle: "",
			wantFound: false,
		},
		{
			name:      "h2 before h1",
			body:      []byte("## Section\n\n# Title\n\nContent."),
			wantTitle: "Title",
			wantFound: true,
		},
		{
			name:      "ATX h1 with code span",
			body:      []byte("# `pd` / Frontmatter\n\nSome content."),
			wantTitle: "pd / Frontmatter",
			wantFound: true,
		},
		{
			name:      "Setext h1 with code span",
			body:      []byte("`pd` / Frontmatter\n===================\n\nSome content."),
			wantTitle: "pd / Frontmatter",
			wantFound: true,
		},
		{
			name:      "mixed inline content in h1",
			body:      []byte("# Prefix `pd` suffix\n\nSome content."),
			wantTitle: "Prefix pd suffix",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, found := heading.ExtractH1(tt.body)
			if found != tt.wantFound {
				t.Fatalf("ExtractH1() found = %v, want %v", found, tt.wantFound)
			}

			if got != tt.wantTitle {
				t.Errorf("ExtractH1() = %q, want %q", got, tt.wantTitle)
			}
		})
	}
}
