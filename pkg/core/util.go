package core

import (
	"io"

	"github.com/spf13/afero"
)

// isEmpty returns true if the directory given is empty.
// Thanks to https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func isEmpty(fs afero.Fs, name string) (bool, error) {
	f, err := fs.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// exists returns true if the given file exists in a file system.
func exists(fs afero.Fs, name string) bool {
	_, err := fs.Stat(name)
	return err == nil
}

// copyFile copies a file from one location to the given destination.
func copyFile(source, dest string) error {
	sourceFile, err := Fs.Open(source)
	if err != nil {
		return err
	}

	destFile, err := Fs.Create(dest)
	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(sourceFile, destFile)
	if err != nil {
		return err
	}

	return nil
}
