package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// Create creates a new entry in the store. If the store is encrypted, it returns ErrStoreEncrypted.
// It takes a path relative to the entries folder, such as "food/pizza" and it will create intermediate directories.
func (s *Store) Create(path, content string) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	relPath := path
	path = filepath.Join(s.entriesPath, path)

	entryPath := filepath.Join(path, "entry.md")
	if exists(entryPath) {
		return ErrEntryAlreadyExists{path}
	}

	_, err = os.Stat(path)
	if err != nil {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(entryPath, []byte(content), 0644)
	if err != nil {
		return err
	}

	err = s.recordChange(relPath, "Add %s", relPath)
	if err != nil {
		return err
	}

	err = s.reload()
	if err != nil {
		return err
	}

	return nil
}
