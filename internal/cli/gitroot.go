package cli

import (
	"context"
	"os/exec"
	"strings"
)

// findGitRoot returns the absolute path of the git repository root by running
// "git rev-parse --show-toplevel". If git is not installed or the current
// directory is not inside a repository, it returns "." as a fallback.
func findGitRoot(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		// git root が見つからない場合は CWD にフォールバック
		return ".", nil //nolint:nilerr // intentional fallback: git not available or not in a repo
	}

	return strings.TrimSpace(string(out)), nil
}
