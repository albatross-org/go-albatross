package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// Store represents an Albatross store.
type Store struct {
	Path string

	entriesPath string
	configPath  string

	coll       *entries.Collection
	repo       *git.Repository
	worktree   *git.Worktree
	disableGit bool

	config *viper.Viper
}

// Load returns a new Albatross store representation.
func Load(path string) (*Store, error) {
	var s = &Store{Path: path, disableGit: false}

	s.entriesPath = filepath.Join(path, "entries")
	s.configPath = filepath.Join(path, "config.yaml")

	config, err := parseConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get config file %s: %w", s.configPath, err)
	}

	s.config = config

	encrypted, err := s.Encrypted()
	if err != nil {
		return nil, err
	}

	if !encrypted {
		err = s.load()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Collection returns the *entries.Collection for the store. It will give an error if the store is currently encrypted.
func (s *Store) Collection() (*entries.Collection, error) {
	encrypted, err := s.Encrypted()
	if err != nil {
		return nil, err
	} else if encrypted {
		return nil, ErrStoreEncrypted{s.Path}
	}

	if s.coll == nil {
		err = s.load()
		if err != nil {
			return nil, err
		}
	}

	return s.coll, nil
}

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

	s.reload()
	return nil
}

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

	s.reload()
	return nil
}

// Attach attaches a file to an entry by copying it into the entry's folder from the location specified. If the store is encrypted, it
// will return ErrStoreEncrypted.
func (s *Store) Attach(path, attachmentPath string) error {
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

	stat, err := os.Stat(attachmentPath)
	if err != nil {
		return fmt.Errorf("attachment %s doesn't exist", attachmentPath)
	}

	attachmentDestinationPath := filepath.Join(path, stat.Name())
	if exists(attachmentDestinationPath) {
		return fmt.Errorf("cannot attach file %s to %s, file already exists", attachmentPath, attachmentDestinationPath)
	}

	err = copyFile(attachmentPath, attachmentDestinationPath)
	if err != nil {
		return fmt.Errorf("cannot copy attachment from %s to %s: %w", attachmentPath, attachmentDestinationPath, err)
	}

	err = s.recordChange(relPath, "Attach %s to %s", attachmentPath, relPath)
	if err != nil {
		return err
	}

	s.reload()
	return nil
}

// Delete deletes an entry and all its attachments from the store. If the store is encrypted, it returns ErrStoreEncrypted.
// It takes a path relative to the entries folder, such as "food/pizza".
// The path given has to be an entry itself, this function cannot be used to delete whole folders of entries.
// If the entry given has subdirectories of entries itself, those subdirectories will be left intact.
// BUG(ollybritton): This code won't delete attachments if they're in a folder next to the entry. The code needs to recursively search
// all subdirectories to determine if they're folders or not.
func (s *Store) Delete(path string) error {
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

	var containsSubEntries bool

	// Here we go through all the files and directories in the path given.
	// containsSubEntries will be set to true if the entry itself contains other entries nested in subdirectories.
	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if info.IsDir() && subpath != path {
			containsSubEntries = true
			return filepath.SkipDir
		}

		if !info.IsDir() {
			return os.Remove(subpath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !containsSubEntries {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	err = s.recordChange(relPath, "Delete %s", relPath)
	if err != nil {
		return err
	}

	s.reload()

	return nil
}

// UsingGit returns true or false depending on whether the store is using Git.
// This will still return true after a call to .DisableGit. The reasoning is that the store is still
// using Git, it's just Git functionality isn't being used by the client.
func (s *Store) UsingGit() bool {
	return s.worktree != nil
}

// DisableGit disables the use of git.
// Calling .UsingGit will still return true. The reasoning is that the store is still
// using Git, it's just Git functionality isn't being used by the client.
func (s *Store) DisableGit() {
	s.disableGit = true
}

// load loads the Collection and in-memory git repository contained within the Store.
func (s *Store) load() error {
	collection, entryErrs, err := entries.DirGraph(s.entriesPath)
	if err != nil {
		return err
	}

	for _, entryErr := range entryErrs {
		logrus.Warn(entryErr)
	}

	s.coll = collection

	err = s.loadGit()
	if err != nil {
		return err
	}

	return nil
}

// loadGit loads git
func (s *Store) loadGit() error {
	repo, err := git.PlainOpen(s.entriesPath)
	if err != nil {
		// Here we ignore an error if we open the git repository.
		// This means that if we're not using git then it won't cause any errors.
		return nil
	}
	s.repo = repo

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	s.worktree = worktree

	return nil
}

// unload unloads the Collection contained within the Store.
func (s *Store) unload() {
	s.coll = nil
	s.repo = nil
	s.worktree = nil
}

// reload is an unload followed by a load. It means changes made are reflected in the store's internal collection.
func (s *Store) reload() {
	s.unload()
	s.load()
}

// recordChange records a change to the store if there is a git repository
func (s *Store) recordChange(path, message string, a ...interface{}) error {
	if s.repo == nil {
		return nil // If we're not using Git, don't do anything.
	}

	if s.disableGit == true {
		return nil // If git has been disabled, also don't do anything
	}

	_, err := s.worktree.Add(path)
	if err != nil {
		return err
	}

	_, err = s.worktree.Commit(
		fmt.Sprintf("(go-albatross) %s", fmt.Sprintf(message, a...)),
		&git.CommitOptions{
			Author: &object.Signature{
				Name: "go-albatross",
				When: time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}
