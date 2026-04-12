package core

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type ignoreState struct {
	patterns []gitignore.Pattern
	matcher  gitignore.Matcher
}

func newIgnoreState(patterns []gitignore.Pattern) ignoreState {
	if len(patterns) == 0 {
		return ignoreState{}
	}

	return ignoreState{
		patterns: patterns,
		matcher:  gitignore.NewMatcher(patterns),
	}
}

func (s ignoreState) ExtendForDirectory(absDir, relDir string) (ignoreState, error) {
	localPatterns, err := loadGitIgnorePatterns(absDir, relDir)
	if err != nil {
		return s, err
	}

	if len(localPatterns) == 0 {
		return s, nil
	}

	combined := make([]gitignore.Pattern, 0, len(s.patterns)+len(localPatterns))
	combined = append(combined, s.patterns...)
	combined = append(combined, localPatterns...)

	return newIgnoreState(combined), nil
}

func (s ignoreState) Match(relPath string, isDir bool) bool {
	if len(s.patterns) == 0 {
		return false
	}

	parts := splitRelPath(relPath)
	if len(parts) == 0 {
		return false
	}

	return s.matcher.Match(parts, isDir)
}

func loadGitIgnorePatterns(absDir, relDir string) ([]gitignore.Pattern, error) {
	gitignorePath := filepath.Join(absDir, ".gitignore")

	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	domain := splitRelPath(relDir)
	var patterns []gitignore.Pattern

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "#") {
			continue
		}

		patterns = append(patterns, gitignore.ParsePattern(raw, domain))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

func splitRelPath(rel string) []string {
	clean := filepath.Clean(rel)
	if clean == "." || clean == "" {
		return nil
	}

	return strings.FieldsFunc(clean, func(r rune) bool {
		return r == '/' || r == '\\'
	})
}