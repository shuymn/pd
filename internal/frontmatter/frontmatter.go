package frontmatter

import (
	"errors"
	"fmt"
	"io"

	"github.com/adrg/frontmatter"
	"github.com/goccy/go-yaml"

	"github.com/shuymn/pd/internal/metadata"
)

// ErrNotFound is returned when no frontmatter is found in the document.
var ErrNotFound = frontmatter.ErrNotFound

// strictUnmarshal rejects unknown fields and duplicate keys.
func strictUnmarshal(data []byte, v any) error {
	return yaml.UnmarshalWithOptions(data, v, yaml.Strict())
}

// strictYAMLFormat is a custom Format that uses strictUnmarshal
// to reject unknown fields and duplicate keys.
var strictYAMLFormat = frontmatter.NewFormat("---", "---", strictUnmarshal)

// Extract reads frontmatter from r into meta using strict YAML parsing,
// and returns the document body after the frontmatter.
// Returns ErrNotFound if no frontmatter is present.
func Extract(r io.Reader, meta *metadata.Metadata) ([]byte, error) {
	body, err := frontmatter.MustParse(r, meta, strictYAMLFormat)
	if err != nil {
		if errors.Is(err, frontmatter.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	return body, nil
}
