package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// exists returns true if the given file exists in a file system.
func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// copyFile copies a file from one location to the given destination.
func copyFile(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error opening attachment source: %s", err)
	}

	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creating attachment destination: %s", err)
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying attachment source to attachment destination: %s", err)
	}

	return nil
}

// folderContainsEntry returns true if a folder contains an entry.md file, no matter how nested it is.
func folderContainsEntry(folder string) (bool, error) {
	containsEntry := false

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "entry.md" {
			containsEntry = true
		}

		if info.IsDir() && path != folder {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return containsEntry, nil
}
