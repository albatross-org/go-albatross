package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/albatross-org/go-albatross/encryption"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// Store represents an Albatross store.
type Store struct {
	Path string

	entriesPath string
	configPath  string

	coll *entries.Collection

	config *viper.Viper
}

// Load returns a new Albatross store representation.
func Load(path string) (*Store, error) {
	var s = &Store{}

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
		err = s.loadCollection()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Encrypted returns true or false depending on whether the store is encrypted or decrypted.
func (s *Store) Encrypted() (bool, error) {
	stat, err := os.Stat(s.entriesPath)
	if err == nil {
		return !stat.IsDir(), nil
	}

	encryptedPath := s.entriesPath + ".gpg"
	stat, err = os.Stat(encryptedPath)
	if err != nil {
		return false, fmt.Errorf("cannot read path specified: %w", err)
	}

	return !stat.IsDir(), nil
}

// Collection returns the *entries.Collection for the store. It will give an error if the store is currently encrypted.
func (s *Store) Collection() (*entries.Collection, error) {
	encrypted, err := s.Encrypted()
	if err != nil {
		return nil, err
	}

	if encrypted {
		return nil, ErrStoreEncrypted{s.Path}
	}

	if s.coll == nil {
		err = s.loadCollection()
		if err != nil {
			return nil, err
		}
	}

	return s.coll, nil
}

// Encrypt encrypts the store. If the store is already encrypted, it returns ErrStoreEncrypted.
func (s *Store) Encrypt() error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	}

	if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	return encryption.EncryptDir(
		s.entriesPath,
		s.entriesPath+".gpg",
		s.config.GetString("encryption.public-key"),
	)
}

// Decrypt decrypts the store. If the store is already decrypted, it will return ErrStoreDecrypted.
// It takes a password func, which is anything that returns a string and an error. This allows to specify the password
// without having to hard code it in.
func (s *Store) Decrypt(passwordFunc func() (string, error)) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	}

	if !encrypted {
		return ErrStoreDecrypted{Path: s.Path}
	}

	pass, err := passwordFunc()
	if err != nil {
		return err
	}

	return encryption.DecryptDir(
		s.entriesPath+".gpg",
		s.entriesPath,
		viper.GetString("encryption.public-key"),
		viper.GetString("encryption.private-key"),
		pass,
	)
}

// loadCollection loads the Collection contained within the Store.
func (s *Store) loadCollection() error {
	collection, entryErrs, err := entries.DirGraph(entries.NewBaseFs(s.entriesPath), "")
	if err != nil {
		return err
	}

	for _, entryErr := range entryErrs {
		logrus.Warn(entryErr)
	}

	s.coll = collection
	return nil
}

// unloadCollection unloads the Collection contained within the Store.
func (s *Store) unloadCollection() {
	s.coll = nil
}
