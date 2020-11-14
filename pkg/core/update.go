package core

import (
	"io/ioutil"
	"path/filepath"
)

// Update updates the given entry. If the store is encrypted, it returns ErrStoreEncrypted.
func (s *Store) Update(path, content string) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	relPath := path
	path = filepath.Join(s.entriesPath, path)

	entryPath := filepath.Join(path, "entry.md")
	if !exists(entryPath) {
		return ErrEntryDoesntExist{path}
	}

	err = ioutil.WriteFile(entryPath, []byte(content), 0644)
	if err != nil {
		return err
	}

	err = s.recordChange(relPath, "Update %s", relPath)
	if err != nil {
		return err
	}

	err = s.reload()
	if err != nil {
		return err
	}

	return nil
}
