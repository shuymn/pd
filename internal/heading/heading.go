package heading

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var defaultParser = goldmark.DefaultParser()

// ExtractH1 extracts the first level-1 heading text from a Markdown document body.
// It returns the heading text and true if found, or empty string and false otherwise.
// Both ATX (# Title) and Setext (Title\n===) headings are supported via goldmark.
func ExtractH1(body []byte) (string, bool) {
	reader := text.NewReader(body)
	doc := defaultParser.Parse(reader)

	var result string

	if err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		h, ok := n.(*ast.Heading)
		if !ok || h.Level != 1 {
			return ast.WalkContinue, nil
		}

		result = extractHeadingText(h, body)

		return ast.WalkStop, nil
	}); err != nil {
		return "", false
	}

	if result == "" {
		return "", false
	}

	return result, true
}

func extractHeadingText(h *ast.Heading, source []byte) string {
	var sb strings.Builder

	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if textNode, ok := c.(*ast.Text); ok {
			sb.Write(textNode.Segment.Value(source))
		}
	}

	return sb.String()
}
