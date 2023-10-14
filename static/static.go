package static

import (
	"os"
	"path/filepath"
)

func Assets() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	path = filepath.Join(path, "..", "public")
	if err = createDirIfNotExists(path); err != nil {
		return "", err
	}

	path = filepath.Join(path, "assets")
	if err = createDirIfNotExists(path); err != nil {
		return "", err
	}

	return path, nil
}

func createDirIfNotExists(name string) error {
	if _, err := os.Stat(name); err != nil {
		if err = os.Mkdir(name, 0755); err != nil {
			return err
		}
	}

	return nil
}
