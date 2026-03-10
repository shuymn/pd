package discovery

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const gitIgnoreFileName = ".gitignore"

type pathIgnorer interface {
	EnterDir(path string) error
	Match(path string, isDir bool) bool
}

type noopIgnorer struct{}

func (noopIgnorer) EnterDir(string) error   { return nil }
func (noopIgnorer) Match(string, bool) bool { return false }

// repositoryIgnorer applies gitignore rules from a git repository.
// It loads descendant .gitignore files lazily as the walk enters each directory,
// eliminating a separate pre-walk pass.
//
// NOTE: pointer receivers are required because EnterDir mutates the accumulated
// pattern list and rebuilds the matcher.
type repositoryIgnorer struct {
	repoRoot         string
	scanRoot         string
	scanRootSegments []string
	patterns         []gitignore.Pattern
	matcher          gitignore.Matcher
}

func newPathIgnorer(scanRoot string) (pathIgnorer, error) {
	absScanRoot, err := filepath.Abs(scanRoot)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(absScanRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return noopIgnorer{}, nil
		}

		return nil, err
	}

	repoRoot, gitDir, ok, err := findGitWorkTreeRoot(absScanRoot)
	if err != nil {
		return nil, err
	}
	if !ok {
		return noopIgnorer{}, nil
	}

	patterns, err := loadInitialPatterns(repoRoot, gitDir, absScanRoot)
	if err != nil {
		return nil, err
	}

	relScanRoot, err := filepath.Rel(repoRoot, absScanRoot)
	if err != nil {
		return nil, err
	}

	return &repositoryIgnorer{
		repoRoot:         repoRoot,
		scanRoot:         absScanRoot,
		scanRootSegments: splitRelativePath(relScanRoot),
		patterns:         patterns,
		matcher:          gitignore.NewMatcher(patterns),
	}, nil
}

// EnterDir loads the .gitignore from the given directory (relative to scanRoot)
// and updates the matcher. The scanRoot itself is skipped because its .gitignore
// was already loaded by loadInitialPatterns as part of the ancestor chain.
func (ri *repositoryIgnorer) EnterDir(relPath string) error {
	if relPath == "." {
		return nil
	}

	absPath := filepath.Join(ri.scanRoot, filepath.FromSlash(relPath))
	relToRepo, err := filepath.Rel(ri.repoRoot, absPath)
	if err != nil {
		return err
	}

	domain := splitRelativePath(relToRepo)
	newPatterns, err := parseIgnoreFile(filepath.Join(absPath, gitIgnoreFileName), domain)
	if err != nil {
		return err
	}
	if len(newPatterns) == 0 {
		return nil
	}

	ri.patterns = append(ri.patterns, newPatterns...)
	ri.matcher = gitignore.NewMatcher(ri.patterns)
	return nil
}

func (ri *repositoryIgnorer) Match(path string, isDir bool) bool {
	segments := append([]string{}, ri.scanRootSegments...)
	segments = append(segments, splitRelativePath(path)...)

	return ri.matcher.Match(segments, isDir)
}

func findGitWorkTreeRoot(start string) (string, string, bool, error) {
	dir := filepath.Clean(start)

	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			if info.IsDir() {
				return dir, gitPath, true, nil
			}

			gitDir, resolveErr := resolveGitDir(gitPath)
			if resolveErr != nil {
				return "", "", false, resolveErr
			}

			return dir, gitDir, true, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", "", false, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", false, nil
		}

		dir = parent
	}
}

func resolveGitDir(gitPath string) (string, error) {
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return "", err
	}

	line := strings.TrimSpace(string(content))
	gitDir, ok := strings.CutPrefix(line, "gitdir: ")
	if !ok {
		return "", fmt.Errorf("parse %s: invalid gitdir file", gitPath)
	}

	if filepath.IsAbs(gitDir) {
		return filepath.Clean(gitDir), nil
	}

	return filepath.Clean(filepath.Join(filepath.Dir(gitPath), gitDir)), nil
}

// loadInitialPatterns loads the info/exclude file and all ancestor .gitignore files
// from repoRoot up to and including scanRoot. Descendant patterns are loaded lazily
// by EnterDir during the main walk.
func loadInitialPatterns(repoRoot, gitDir, scanRoot string) ([]gitignore.Pattern, error) {
	patterns := make([]gitignore.Pattern, 0)

	infoExcludePatterns, err := parseIgnoreFile(filepath.Join(gitDir, "info", "exclude"), nil)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, infoExcludePatterns...)

	ancestorPatterns, err := loadAncestorPatterns(repoRoot, scanRoot)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, ancestorPatterns...)

	return patterns, nil
}

func loadAncestorPatterns(repoRoot, scanRoot string) ([]gitignore.Pattern, error) {
	ancestorDirs, err := ancestorDomains(repoRoot, scanRoot)
	if err != nil {
		return nil, err
	}

	patterns := make([]gitignore.Pattern, 0)
	for _, domain := range ancestorDirs {
		gitignorePatterns, readErr := parseIgnoreFile(
			filepath.Join(repoRoot, filepath.Join(domain...), gitIgnoreFileName),
			domain,
		)
		if readErr != nil {
			return nil, readErr
		}

		patterns = append(patterns, gitignorePatterns...)
	}

	return patterns, nil
}

func ancestorDomains(repoRoot, scanRoot string) ([][]string, error) {
	relPath, err := filepath.Rel(repoRoot, scanRoot)
	if err != nil {
		return nil, err
	}

	segments := splitRelativePath(relPath)
	domains := make([][]string, 0, len(segments)+1)
	for i := 0; i <= len(segments); i++ {
		domain := append([]string{}, segments[:i]...)
		domains = append(domains, domain)
	}

	return domains, nil
}

func parseIgnoreFile(path string, domain []string) ([]gitignore.Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	defer func() { _ = f.Close() }()

	patterns := make([]gitignore.Pattern, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		patterns = append(patterns, gitignore.ParsePattern(line, domain))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

func splitRelativePath(path string) []string {
	cleaned := filepath.Clean(path)
	if cleaned == "." || cleaned == "" {
		return nil
	}

	return strings.Split(cleaned, string(filepath.Separator))
}
