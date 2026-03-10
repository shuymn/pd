package frontmatter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/shuymn/pd/internal/frontmatter"
	"github.com/shuymn/pd/internal/metadata"
)

func TestExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    metadata.Metadata
		wantErr bool
		errIs   error
	}{
		{
			name: "valid with title",
			input: `---
kind: roadmap
description: A roadmap document
title: My Roadmap
---
# Body content
`,
			want: metadata.Metadata{
				Kind:        metadata.KindRoadmap,
				Description: "A roadmap document",
				Title:       "My Roadmap",
			},
		},
		{
			name: "valid without title",
			input: `---
kind: adr
description: An ADR document
---
# Body content
`,
			want: metadata.Metadata{
				Kind:        metadata.KindADR,
				Description: "An ADR document",
			},
		},
		{
			name: "no frontmatter",
			input: `# Just a heading

Some content without frontmatter.
`,
			wantErr: true,
			errIs:   frontmatter.ErrNotFound,
		},
		{
			name: "unknown field rejected",
			input: `---
kind: roadmap
description: A roadmap document
extra_field: not allowed
---
Body
`,
			wantErr: true,
		},
		{
			name: "malformed YAML",
			input: `---
kind: [invalid yaml
---
Body
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var meta metadata.Metadata

			body, err := frontmatter.Extract(strings.NewReader(tt.input), &meta)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Extract() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Errorf("Extract() error = %v, want errors.Is(%v)", err, tt.errIs)
			}

			if !tt.wantErr {
				if meta.Kind != tt.want.Kind {
					t.Errorf("meta.Kind = %q, want %q", meta.Kind, tt.want.Kind)
				}

				if meta.Description != tt.want.Description {
					t.Errorf("meta.Description = %q, want %q", meta.Description, tt.want.Description)
				}

				if meta.Title != tt.want.Title {
					t.Errorf("meta.Title = %q, want %q", meta.Title, tt.want.Title)
				}

				if body == nil {
					t.Error("Extract() body is nil, want non-nil")
				}
			}
		})
	}
}
