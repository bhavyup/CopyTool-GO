//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func defaultInstallRoot() string {
	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join("C:\\", "Users", "Public", "AppData", "Local", "Programs", "copytool")
		}
		localAppData = filepath.Join(home, "AppData", "Local")
	}

	return filepath.Join(localAppData, "Programs", "copytool")
}

func ensureUserPathContains(dir string) (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, fmt.Errorf("open user environment key: %w", err)
	}
	defer key.Close()

	currentPath := ""
	storedPath, _, err := key.GetStringValue("Path")
	if err != nil {
		if err != registry.ErrNotExist {
			return false, fmt.Errorf("read user Path: %w", err)
		}
	} else {
		currentPath = storedPath
	}

	if pathHasEntry(currentPath, dir) {
		return false, nil
	}

	newPath := currentPath
	if strings.TrimSpace(newPath) == "" {
		newPath = dir
	} else {
		newPath = newPath + ";" + dir
	}

	if err := key.SetStringValue("Path", newPath); err != nil {
		return false, fmt.Errorf("write user Path: %w", err)
	}

	return true, nil
}

func pathHasEntry(pathValue, dir string) bool {
	target := normalizePathEntry(dir)
	if target == "" {
		return false
	}

	for _, entry := range strings.Split(pathValue, ";") {
		if normalizePathEntry(entry) == target {
			return true
		}
	}

	return false
}

func normalizePathEntry(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	return strings.ToLower(strings.TrimRight(trimmed, `\\/`))
}
