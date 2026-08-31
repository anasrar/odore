package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func UnpackPath(path string) (string, string, error) {
	filename := fmt.Sprintf("UNPACK_%s", filepath.Base(path))
	directory := ""

	if filepath.IsAbs(path) {
		directory = filepath.Dir(filepath.Clean(path))
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}

		directory = filepath.Dir(filepath.Clean(filepath.Join(cwd, path)))
	}

	pathDirectoryFiles := filepath.Join(directory, filename, "FILES")
	pathMetadata := filepath.Join(directory, filename, "METADATA.json")

	return pathDirectoryFiles, pathMetadata, nil
}

func PackPath(path string) string {
	directory := filepath.Dir(filepath.Clean(path))
	return directory
}
