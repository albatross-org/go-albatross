package albatross

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/albatross-org/go-albatross/entries"
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

// FastCreate is an experimental method for adding an entry without the need for a full reload.
// It does this by adding the entry to the entries.Collection directly, rather than unloading and reloading the whole store.
func (s *Store) FastCreate(path, content string) error {
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

	entryParser, err := entries.NewParser(s.Config.DateFormat, s.Config.TagPrefix)
	if err != nil {
		return err
	}

	entry, err := entryParser.Parse(relPath, content)
	if err != nil {
		return err
	}

	err = s.coll.Add(entry)
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
