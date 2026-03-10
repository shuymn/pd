package metadata

import "fmt"

// Kind represents the document kind enum.
type Kind string

const (
	// KindRoadmap is the kind for roadmap documents.
	KindRoadmap Kind = "roadmap"
	// KindDesignDoc is the kind for design documents.
	KindDesignDoc Kind = "design-doc"
	// KindADR is the kind for architecture decision records.
	KindADR Kind = "adr"
	// KindCoding is the kind for coding convention documents.
	KindCoding Kind = "coding"
	// KindTesting is the kind for testing convention documents.
	KindTesting Kind = "testing"
	// KindTooling is the kind for tooling convention documents.
	KindTooling Kind = "tooling"
	// KindReview is the kind for review convention documents.
	KindReview Kind = "review"
	// KindUnknown is the kind for unclassified documents.
	KindUnknown Kind = "unknown"
)

var validKinds = map[Kind]struct{}{
	KindRoadmap:   {},
	KindDesignDoc: {},
	KindADR:       {},
	KindCoding:    {},
	KindTesting:   {},
	KindTooling:   {},
	KindReview:    {},
	KindUnknown:   {},
}

// IsValid reports whether k is a recognized Kind value.
func (k Kind) IsValid() bool {
	_, ok := validKinds[k]
	return ok
}

// ParseKind parses a string into a Kind.
func ParseKind(s string) (Kind, error) {
	k := Kind(s)
	if !k.IsValid() {
		return "", fmt.Errorf("invalid kind: %q", s)
	}

	return k, nil
}

// Metadata is the frontmatter schema for discovery documents.
type Metadata struct {
	Kind        Kind   `yaml:"kind"        json:"kind"`
	Description string `yaml:"description" json:"description"`
	Title       string `yaml:"title"       json:"title,omitempty"`
}

// Result is the output record for a discovered document.
type Result struct {
	Path        string `json:"path"`
	Kind        Kind   `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Validate validates the semantic constraints of Metadata.
// Title presence is the caller's responsibility (H1 fallback interaction).
// It returns a non-empty reason string when validation fails, and a non-nil error for internal failures.
func Validate(m Metadata) (reason string, err error) {
	if m.Kind == "" {
		return "missing required field: kind", nil
	}

	if !m.Kind.IsValid() {
		return fmt.Sprintf("invalid kind: %q", string(m.Kind)), nil
	}

	if m.Description == "" {
		return "missing required field: description", nil
	}

	return "", nil
}
