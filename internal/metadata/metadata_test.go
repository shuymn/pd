package metadata_test

import (
	"testing"

	"github.com/shuymn/pd/internal/metadata"
)

func TestParseKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    metadata.Kind
		wantErr bool
	}{
		{name: "roadmap", input: "roadmap", want: metadata.KindRoadmap},
		{name: "design-doc", input: "design-doc", want: metadata.KindDesignDoc},
		{name: "adr", input: "adr", want: metadata.KindADR},
		{name: "coding", input: "coding", want: metadata.KindCoding},
		{name: "testing", input: "testing", want: metadata.KindTesting},
		{name: "tooling", input: "tooling", want: metadata.KindTooling},
		{name: "review", input: "review", want: metadata.KindReview},
		{name: "unknown", input: "unknown", want: metadata.KindUnknown},
		{name: "empty", input: "", wantErr: true},
		{name: "invalid", input: "blog", wantErr: true},
		{name: "uppercase", input: "Roadmap", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := metadata.ParseKind(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKind(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseKind(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      metadata.Metadata
		wantReason string
	}{
		{
			name: "valid with title",
			input: metadata.Metadata{
				Kind:        metadata.KindRoadmap,
				Description: "A roadmap document",
				Title:       "My Roadmap",
			},
		},
		{
			name: "valid without title",
			input: metadata.Metadata{
				Kind:        metadata.KindADR,
				Description: "An ADR document",
			},
		},
		{
			name:       "missing kind",
			input:      metadata.Metadata{Description: "desc"},
			wantReason: "missing required field: kind",
		},
		{
			name:       "invalid kind",
			input:      metadata.Metadata{Kind: "blog", Description: "desc"},
			wantReason: `invalid kind: "blog"`,
		},
		{
			name:       "missing description",
			input:      metadata.Metadata{Kind: metadata.KindRoadmap},
			wantReason: "missing required field: description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reason, err := metadata.Validate(tt.input)
			if err != nil {
				t.Fatalf("Validate() unexpected error = %v", err)
			}

			if reason != tt.wantReason {
				t.Errorf("Validate() reason = %q, want %q", reason, tt.wantReason)
			}
		})
	}
}
