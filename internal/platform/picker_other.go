//go:build !windows

package platform

func SupportsNativePicker() bool { return false }

func PickFolder(startDir string) (path string, cancelled bool, err error) {
	return "", true, nil
}

func PickSaveFile(startDir string, defaultName string) (path string, cancelled bool, err error) {
	return "", true, nil
}