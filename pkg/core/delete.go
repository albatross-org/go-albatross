package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Delete deletes an entry and all its attachments from the store. If the store is encrypted, it returns ErrStoreEncrypted.
// It takes a path relative to the entries folder, such as "food/pizza".
// The path given has to be an entry itself, this function cannot be used to delete whole folders of entries.
// If the entry given has subdirectories of entries itself, those subdirectories will be left intact.
func (s *Store) Delete(path string) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	relPath := path
	path = filepath.Join(s.entriesPath, path)

	// Check if the entry we're trying to delete actually exists.
	entryPath := filepath.Join(path, "entry.md")
	if !exists(entryPath) {
		return ErrEntryDoesntExist{path}
	}

	var containsSubEntries bool

	// Here we go through all the files and directories in the path given.
	// containsSubEntries will be set to true if the entry itself contains other entries nested in subdirectories.
	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If we find a directory that isn't the parent entry (i.e. the folder containing the entry we're trying to delete)
		// then we need to check if it's a folder that contains entries. Otherwise we can delete it.
		if info.IsDir() && subpath != path {
			containsEntry, err := folderContainsEntry(subpath)
			if err != nil {
				return fmt.Errorf("couldn't determine whether folder in entry contains sub-entries: %w", err)
			}

			if containsEntry {
				containsSubEntries = true
				return filepath.SkipDir
			}
		}

		// If it's not a directory, we remove the file directly and record the change via Git.
		if !info.IsDir() {
			err = os.Remove(subpath)
			if err != nil {
				return err
			}

			if s.repo != nil && s.disableGit != true {
				relSubpath := strings.TrimPrefix(subpath, s.entriesPath+"/")
				_, err := s.worktree.Add(relSubpath) // We use 'add' here although this is 'adding' a removal.
				if err != nil {
					return fmt.Errorf("couldn't record removal %s: %w", relSubpath, err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// If the folder containing the entry we were trying to delete contains no other entries, it's safe to remove completely.
	if !containsSubEntries {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	// If we're using Git we add a commit recording what we've deleted.
	if s.repo != nil && s.disableGit != true {
		_, err = s.worktree.Commit(
			fmt.Sprintf("(go-albatross) Delete %s", relPath),
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
	}

	// Reload the store to rebuilt the underlying Collection structure.
	err = s.reload()
	if err != nil {
		return err
	}

	return nil
}
