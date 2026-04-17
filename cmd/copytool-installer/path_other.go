//go:build !windows

package main

import (
	"os"
	"path/filepath"
)

func defaultInstallRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "copytool")
	}

	return filepath.Join(home, ".local", "copytool")
}

func ensureUserPathContains(dir string) (bool, error) {
	return false, nil
}
