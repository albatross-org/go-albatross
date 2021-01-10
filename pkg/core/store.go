package core

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/albatross-org/go-albatross/entries"

	"github.com/spf13/viper"
)

// Store represents an Albatross store.
type Store struct {
	Path string

	entriesPath     string
	configPath      string
	attachmentsPath string

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
	s.attachmentsPath = filepath.Join(path, "attachments")

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
	collection, entryErrs, err := entries.FromDirectoryAsync(s.entriesPath)
	if err != nil {
		return err
	}

	for _, entryErr := range entryErrs {
		log.Warn(entryErr)
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
// It will return an error if load failed.
func (s *Store) reload() error {
	s.unload()
	return s.load()
}

// recordChange records a change to the store if there is a git repository
func (s *Store) recordChange(path, message string, a ...interface{}) error {
	// If we're not using Git or Git has been disabled, don't do anything
	if s.repo == nil || s.disableGit == true {
		return nil
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
