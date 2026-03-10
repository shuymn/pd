package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

// normalizePath returns a base-dir-relative path and rejects values that escape the base directory.
func normalizePath(flagName, baseDir, path string) (string, error) {
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("resolve base dir %q: %w", baseDir, err)
	}

	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(absBaseDir, path)
		if err != nil {
			return "", fmt.Errorf("%s must stay within base dir %q: %w", flagName, absBaseDir, err)
		}

		path = rel
	}

	cleaned := filepath.Clean(path)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s must not traverse above the current working directory, got %q", flagName, path)
	}

	return cleaned, nil
}
