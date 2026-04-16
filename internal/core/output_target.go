package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveOutputPath(rootPath, outputSpec string) (string, error) {
	spec := strings.TrimSpace(outputSpec)

	if spec == "" {
		return filepath.Join(rootPath, "tools", "output.txt"), nil
	}

	if spec == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current working directory for output: %w", err)
		}
		return filepath.Join(cwd, "tools", "output.txt"), nil
	}

	absTarget, err := filepath.Abs(spec)
	if err != nil {
		return "", fmt.Errorf("resolve output target: %w", err)
	}

	info, err := os.Stat(absTarget)
	if err == nil {
		if info.IsDir() {
			return filepath.Join(absTarget, "tools", "output.txt"), nil
		}
		return absTarget, nil
	}

	if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect output target: %w", err)
	}

	// Path does not exist yet, so infer intent.
	// Rules:
	// - trailing slash => directory
	// - explicit file extension => file
	// - otherwise treat as directory
	if hasTrailingSeparator(spec) {
		return filepath.Join(absTarget, "tools", "output.txt"), nil
	}

	if ext := filepath.Ext(filepath.Base(absTarget)); ext != "" {
		return absTarget, nil
	}

	return filepath.Join(absTarget, "tools", "output.txt"), nil
}

func hasTrailingSeparator(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, `\`)
}
