package albatross

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/albatross-org/go-albatross/entries"
)

// Store represents an Albatross store.
type Store struct {
	Path   string
	Config *Config

	entriesPath     string
	configPath      string
	attachmentsPath string
	gitPath         string

	coll       *entries.Collection
	disableGit bool
	hasGit     bool
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
	return s.hasGit
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

	// We test for a Git enabled store by seeing if the '.git' folder exists.
	if exists(s.gitPath) {
		s.hasGit = true
	}

	return nil
}

// unload unloads the Collection contained within the Store.
func (s *Store) unload() {
	s.coll = nil
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
	if !s.UsingGit() || s.disableGit {
		return nil
	}

	start := time.Now()

	addCmd := s.gitCmd("add", path)
	err := addCmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't add change %s to Git: %w", path, err)
	}

	worktreeTime := time.Now()

	commitCmd := s.gitCmd("commit", "--author", "go-albatross <>", "-m", "(go-albatross) "+fmt.Sprintf(message, a...))
	err = commitCmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't commit changes %s to Git: %w", s.Path, err)
	}

	commitTime := time.Now()

	log.Debugf(
		"Recording change via Git, time to add to worktree: %s, time to commit: %s, overall: %s",
		worktreeTime.Sub(start),
		commitTime.Sub(worktreeTime),
		commitTime.Sub(start),
	)

	return nil
}

// gitCmd creates a new exec.Cmd that contains the correct flags.
func (s *Store) gitCmd(args ...string) *exec.Cmd {
	return exec.Command("git", append([]string{"--git-dir", s.gitPath, "--work-tree", s.entriesPath}, args...)...)
}
