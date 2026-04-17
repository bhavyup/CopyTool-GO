package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var defaultRepo = "bhavyup/CopyTool-GO"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "copytool installer failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	versionFlag := flag.String("version", "latest", "release version (for example v0.1.2) or 'latest'")
	repoFlag := flag.String("repo", defaultRepo, "GitHub repository in owner/name format")
	installRootFlag := flag.String("install-root", defaultInstallRoot(), "installation root directory")
	skipPathFlag := flag.Bool("skip-path", false, "skip adding install bin directory to user PATH")
	timeoutFlag := flag.Duration("timeout", 90*time.Second, "download timeout")

	flag.Parse()

	if runtime.GOOS != "windows" {
		return errors.New("copytool-installer.exe is intended for Windows")
	}

	repo := strings.TrimSpace(*repoFlag)
	if repo == "" || !strings.Contains(repo, "/") {
		return fmt.Errorf("invalid repo %q, expected owner/name", repo)
	}

	tag, err := resolveTag(repo, strings.TrimSpace(*versionFlag), *timeoutFlag)
	if err != nil {
		return err
	}

	arch, err := releaseArch()
	if err != nil {
		return err
	}

	versionNoPrefix := strings.TrimPrefix(tag, "v")
	asset := fmt.Sprintf("copytool_%s_windows_%s.exe", versionNoPrefix, arch)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, asset)

	fmt.Printf("Downloading %s\n", url)
	tempPath, err := downloadAsset(url, *timeoutFlag)
	if err != nil {
		return err
	}
	defer os.Remove(tempPath)

	installRoot := strings.TrimSpace(*installRootFlag)
	if installRoot == "" {
		return errors.New("install-root cannot be empty")
	}

	binDir := filepath.Join(installRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create install directory: %w", err)
	}

	targetPath := filepath.Join(binDir, "copytool.exe")
	if err := installBinary(tempPath, targetPath); err != nil {
		return err
	}

	pathUpdated := false
	if !*skipPathFlag {
		added, err := ensureUserPathContains(binDir)
		if err != nil {
			return err
		}
		pathUpdated = added
	}

	fmt.Println("")
	fmt.Println("Installation complete")
	fmt.Printf("  binary: %s\n", targetPath)

	if *skipPathFlag {
		fmt.Println("  PATH: skipped by flag")
	} else if pathUpdated {
		fmt.Printf("  PATH: added %s\n", binDir)
	} else {
		fmt.Printf("  PATH: already contained %s\n", binDir)
	}

	fmt.Println("")
	fmt.Println("Open a NEW terminal window to pick up PATH changes.")

	if out, err := exec.Command(targetPath, "-version").CombinedOutput(); err == nil {
		versionText := strings.TrimSpace(string(out))
		if versionText != "" {
			fmt.Println("")
			fmt.Println("Installed version:")
			fmt.Println(versionText)
		}
	}

	return nil
}

func releaseArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64", "arm64":
		return runtime.GOARCH, nil
	default:
		return "", fmt.Errorf("unsupported Windows architecture: %s", runtime.GOARCH)
	}
}

func resolveTag(repo, requested string, timeout time.Duration) (string, error) {
	if requested == "" || strings.EqualFold(requested, "latest") {
		return fetchLatestTag(repo, timeout)
	}

	tag := strings.TrimSpace(requested)
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	return tag, nil
}

type latestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func fetchLatestTag(repo string, timeout time.Duration) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("build latest release request: %w", err)
	}

	req.Header.Set("User-Agent", "copytool-installer")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request latest release metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("latest release lookup failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload latestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("parse latest release metadata: %w", err)
	}

	tag := strings.TrimSpace(payload.TagName)
	if tag == "" {
		return "", errors.New("latest release metadata did not include tag_name")
	}

	return tag, nil
}

func downloadAsset(url string, timeout time.Duration) (string, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build asset download request: %w", err)
	}

	req.Header.Set("User-Agent", "copytool-installer")
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download release asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("asset download failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	tmpFile, err := os.CreateTemp("", "copytool-asset-*.exe")
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}

	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("write downloaded asset: %w", err)
	}

	return tmpFile.Name(), nil
}

func installBinary(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open downloaded asset: %w", err)
	}
	defer srcFile.Close()

	tmpPath := dstPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create install temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("copy binary to install temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("flush install temp file: %w", err)
	}

	if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("remove previous install: %w", err)
	}

	if err := os.Rename(tmpPath, dstPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("finalize binary install: %w", err)
	}

	return nil
}
