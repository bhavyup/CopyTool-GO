//go:build windows

package platform

import "github.com/sqweek/dialog"

func SupportsNativePicker() bool { return true }

func PickFolder(startDir string) (path string, cancelled bool, err error) {
	d := dialog.Directory().Title("Select folder")
	if startDir != "" {
		d = d.SetStartDir(startDir)
	}

	path, err = d.Browse()
	if err == dialog.Cancelled {
		return "", true, nil
	}
	if err != nil {
		return "", false, err
	}
	return path, false, nil
}

func PickSaveFile(startDir string, defaultName string) (path string, cancelled bool, err error) {
	d := dialog.File().Title("Select output file")
	if startDir != "" {
		d = d.SetStartDir(startDir)
	}
	if defaultName != "" {
		d = d.SetStartFile(defaultName)
	}

	path, err = d.Save()
	if err == dialog.Cancelled {
		return "", true, nil
	}
	if err != nil {
		return "", false, err
	}
	return path, false, nil
}