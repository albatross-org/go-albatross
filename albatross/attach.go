package albatross

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/plus3it/gorecurcopy"
)

// AttachCopy attaches a file to an entry by copying it into the entry's folder from the location specified. If the store is encrypted, it
// will return ErrStoreEncrypted.
func (s *Store) AttachCopy(path, attachmentPath string) error {
	// Check what is being attached actually exists.
	stat, err := os.Stat(attachmentPath)
	if err != nil {
		return fmt.Errorf("attachment %s doesn't exist: %w", attachmentPath, err)
	}

	attachmentName := stat.Name()

	return s.AttachCopyWithName(path, attachmentPath, attachmentName)
}

// AttachCopyWithName attaches a file to an entry by copying it into the entry's folder from the location specified. If the store is encrypted, it
// will return ErrStoreEncrypted.
// AttachCopyWithName differs from AttachCopy because you can specify an attachmentName, which is what the file will be called when attached.
func (s *Store) AttachCopyWithName(path, attachmentPath, attachmentName string) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	relPath := path
	path = filepath.Join(s.entriesPath, path)

	// Check the entry we're trying to attach something to actually exists.
	entryPath := filepath.Join(path, "entry.md")
	if !exists(entryPath) {
		return ErrEntryDoesntExist{path}
	}

	// Check what is being attached actually exists.
	stat, err := os.Stat(attachmentPath)
	if err != nil {
		return fmt.Errorf("attachment %s doesn't exist: %w", attachmentPath, err)
	}

	// Find out what the attachment should be called and get what the path to attachment should be.
	newAttachmentPath := filepath.Join(path, attachmentName)

	// Check if there already exists an attachment with that name in the entry.
	if exists(newAttachmentPath) {
		return fmt.Errorf("cannot attach file %s to %s, file already exists", attachmentPath, newAttachmentPath)
	}

	if !stat.IsDir() {
		err = gorecurcopy.Copy(attachmentPath, newAttachmentPath)
		if err != nil {
			return fmt.Errorf("couldn't attach file: %w", err)
		}
	} else {
		err = gorecurcopy.CopyDirectory(attachmentPath, newAttachmentPath)
		if err != nil {
			return fmt.Errorf("couldn't attach folder: %w", err)
		}
	}

	err = s.recordChange(relPath, "Attach via Copy %s to %s", attachmentPath, relPath)
	if err != nil {
		return err
	}

	err = s.reload()
	if err != nil {
		return err
	}

	return nil
}

// AttachSymlink attaches a file by creating a symlink to the store's attachment's folder. If the store is encrypted,
// it will return ErrStoreEncrypted. path is the path to the entry, relative to the store, and attachmentPath is the
// path to what is being attached.
func (s *Store) AttachSymlink(path, attachmentPath string) error {
	// Check what is being attached actually exists.
	stat, err := os.Stat(attachmentPath)
	if err != nil {
		return fmt.Errorf("attachment %s doesn't exist: %w", attachmentPath, err)
	}

	attachmentName := stat.Name()

	return s.AttachSymlinkWithName(path, attachmentPath, attachmentName)
}

// AttachSymlinkWithName attaches a file by creating a symlink to the store's attachment's folder. If the store is encrypted,
// it will return ErrStoreEncrypted. path is the path to the entry, relative to the store, and attachmentPath is the
// path to what is being attached.
// This differs from AttachSymlink as you can give a new name for the attachment.
func (s *Store) AttachSymlinkWithName(path, attachmentPath, attachmentName string) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	relPath := path
	path = filepath.Join(s.entriesPath, path)

	// Check the entry we're trying to attach something to actually exists.
	entryPath := filepath.Join(path, "entry.md")
	if !exists(entryPath) {
		return ErrEntryDoesntExist{path}
	}

	// Check what is being attached actually exists.
	stat, err := os.Stat(attachmentPath)
	if err != nil {
		return fmt.Errorf("attachment %s doesn't exist: %w", attachmentPath, err)
	}

	// Find out what the attachment should be called and get what the path to attachment should be.
	newAttachmentPath := filepath.Join(path, attachmentName)

	// Check if there already exists an attachment with that name in the entry.
	if exists(newAttachmentPath) {
		return fmt.Errorf("cannot attach file %s to %s, file already exists", attachmentPath, newAttachmentPath)
	}

	// Check if the store has an `attachments/`. If not, create one.
	if !exists(s.attachmentsPath) {
		err = os.Mkdir(s.attachmentsPath, 0755)
		if err != nil {
			return fmt.Errorf("couldn't create attachments/ folder in root of store: %w", err)
		}
	}

	// Calculate a hash for the file/folder.
	// The filename will become something like 'attachments/attachment-4ae25f492...'
	hash, err := hashPath(attachmentPath)
	if err != nil {
		return fmt.Errorf("error creating hash for attachment: %w", err)
	}
	filename := filepath.Join(s.attachmentsPath, "attachment-"+hash)

	// Copy the file to the correct location. If we get an error that the file already exists, we don't actually need
	// to worry about it because if two files have the same hash then they are likely the same file.
	if !stat.IsDir() {
		err = gorecurcopy.Copy(attachmentPath, filename)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf("couldn't attach file: %w", err)
		}
	} else {
		err = os.Mkdir(filename, 0755)
		if err != nil {
			return fmt.Errorf("couldn't attach folder: couldn't create folder: %w", err)
		}

		err = gorecurcopy.CopyDirectory(attachmentPath, filename)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf("couldn't attach folder: %w", err)
		}
	}

	// Here we need to create a relative path from the entry's path (relPath) that points to the attachment
	// This is because the symlink needs a relative path rather than an absolute one otherwise to different computers
	// with varying paths to the stores would not both have correct paths.
	//
	// For example, if we're attaching 'attachment-4ae25f492...' to the 'roadtrip/' entry:
	// ~/.local/share/albatross/default/entries/roadtrip (exact path of the entry): ".." goes to
	// ~/.local/share/albatross/default/entries 								    "../.." goes to
	// ~/.local/share/albatross/default/        								    "../../attachments" goes to
	// ~/.local/share/albatross/default/attachments
	//
	// Which is what we want. So in that example we got "../../attachments" as the path we wanted.
	// One way of thinking about this problem is "how many calls to filepath.Dir() until we get the path of the store".
	count := 0
	currentPath := path

	for currentPath != s.Path && currentPath != "." {
		currentPath = filepath.Dir(currentPath)
		count++
	}

	// Something went wrong -- finding the parent directory of the entry never matched the path to the store.
	if currentPath == "." || count == 0 {
		return fmt.Errorf("error finding a relative path from %s to %s", path, s.attachmentsPath)
	}

	// Finally, get the relative path to the attachment and create the symlink.
	relativePathToAttachment := strings.Repeat("../", count) + "attachments/" + "attachment-" + hash
	err = os.Symlink(relativePathToAttachment, newAttachmentPath)
	if err != nil {
		return fmt.Errorf("couldn't create relative symlink from %s to %s: %w", newAttachmentPath, relativePathToAttachment, err)
	}

	// Record the change via Git if it's being used.
	err = s.recordChange(relPath, "Attach via Symlink %s to %s", attachmentPath, relPath)
	if err != nil {
		return err
	}

	// Reload the store to refresh the underlying store.Collection.
	err = s.reload()
	if err != nil {
		return err
	}

	return nil
}

// Attach attaches a file by creating a symlink to the store's attachment's folder. If the store is encrypted,
// it will return ErrStoreEncrypted. path is the path to the entry, relative to the store, and attachmentPath is the
// path to what is being attached.
// This command is an alias for AttachSymlink, AttachCopy is also available.
func (s *Store) Attach(path, attachmentPath string) error {
	return s.AttachSymlink(path, attachmentPath)
}

// hash calculates the hash of a folder or a file. If it's a file, the contents are used as input. If it's a folder, the
// folder is walked and the hash is calculated for each file combined with it's path. These hashes are then combined and
// then this string is then hashed.
func hashPath(path string) (string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("couldn't stat path %s for hashing: %w", path, err)
	}

	if stat.IsDir() {
		var out bytes.Buffer

		err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := os.Open(subpath)
			if err != nil {
				return fmt.Errorf("couldn't open subpath %s: %w", subpath, err)
			}
			defer f.Close()

			// Add the file's contents as arguments for what is to be hashed.
			h := sha256.New()
			_, err = io.Copy(h, f)
			if err != nil {
				return err
			}

			// Add the file's path relative to the folder for what is to be hashed.
			_, err = io.WriteString(h, strings.TrimPrefix(subpath, path))
			if err != nil {
				return err
			}

			out.WriteString(string(h.Sum(nil)))

			return nil
		})
		if err != nil {
			return "", err
		}

		h := sha256.New()
		io.WriteString(h, out.String())

		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("couldn't hash path %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
