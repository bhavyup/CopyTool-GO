package core

import (
	"path/filepath"
	"sort"
	"strings"
)

type FilterConfig struct {
	IncludeExts map[string]bool
	ExcludeExts map[string]bool
	ExcludeDirs map[string]bool
}

var hardExcludedDirNames = []string{
	".git",
	"tools",
}

var defaultUserExcludedDirNames = []string{
	"node_modules",
	"bin",
	"obj",
	".vs",
	"dist",
	"build",
}

func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		IncludeExts: map[string]bool{},
		ExcludeExts: map[string]bool{},
		ExcludeDirs: DefaultUserExcludeDirs(),
	}
}

func DefaultUserExcludeDirs() map[string]bool {
	return makeNameSet(defaultUserExcludedDirNames)
}

func HardExcludeDirs() map[string]bool {
	return makeNameSet(hardExcludedDirNames)
}

func DefaultExcludeDirs() map[string]bool {
	merged := HardExcludeDirs()
	for name := range DefaultUserExcludeDirs() {
		merged[name] = true
	}
	return merged
}

func IsHardExcludedDir(name string) bool {
	_, ok := HardExcludeDirs()[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func ParseExtensionCSV(raw string) map[string]bool {
	set := map[string]bool{}

	for _, part := range strings.Split(raw, ",") {
		normalized := NormalizeExtension(part)
		if normalized == "" {
			continue
		}
		set[normalized] = true
	}

	return set
}

func ParseNameCSV(raw string) map[string]bool {
	set := map[string]bool{}

	for _, part := range strings.Split(raw, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		set[name] = true
	}

	return set
}

func FormatSetCSV(set map[string]bool) string {
	if len(set) == 0 {
		return ""
	}

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}

	sort.Strings(values)
	return strings.Join(values, ",")
}

func NormalizeExtension(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	ext := trimmed
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	ext = strings.ToLower(ext)
	if ext == "." {
		return ""
	}

	return ext
}

func FilePassesExtensionFilters(path string, filters FilterConfig) bool {
	ext := strings.ToLower(filepath.Ext(path))

	if len(filters.IncludeExts) > 0 {
		if !filters.IncludeExts[ext] {
			return false
		}
	}

	if filters.ExcludeExts[ext] {
		return false
	}

	return true
}

func makeNameSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[strings.ToLower(strings.TrimSpace(value))] = true
	}
	return set
}
