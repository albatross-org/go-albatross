package core

import (
	"fmt"
	"io"
	"os"
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
