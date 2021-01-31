package albatross

import (
	"fmt"
	"os"

	"github.com/albatross-org/go-albatross/encryption"
)

// Encrypted returns true or false depending on whether the store is encrypted or decrypted.
func (s *Store) Encrypted() (bool, error) {
	_, err := os.Stat(s.entriesPath)
	if err == nil {
		return false, nil
	}

	encryptedPath := s.entriesPath + ".gpg"
	_, err = os.Stat(encryptedPath)
	if err != nil {
		return false, fmt.Errorf("cannot read path specified: %s", err)
	}

	return true, nil
}

// Encrypt encrypts the store. If the store is already encrypted, it returns ErrStoreEncrypted.
func (s *Store) Encrypt() error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if encrypted {
		return ErrStoreEncrypted{Path: s.Path}
	}

	err = encryption.EncryptDir(
		s.entriesPath,
		s.entriesPath+".gpg",
		s.Config.Encryption.PublicKey,
	)
	if err != nil {
		return err
	}

	return os.RemoveAll(s.entriesPath)
}

// Decrypt decrypts the store. If the store is already decrypted, it will return ErrStoreDecrypted.
// It takes a password func, which is anything that returns a string and an error. This allows to specify the password
// without having to hard code it in.
func (s *Store) Decrypt(passwordFunc func() (string, error)) error {
	encrypted, err := s.Encrypted()
	if err != nil {
		return err
	} else if !encrypted {
		return ErrStoreDecrypted{Path: s.Path}
	}

	pass, err := passwordFunc()
	if err != nil {
		return err
	}

	err = encryption.DecryptDir(
		s.entriesPath+".gpg",
		s.entriesPath,
		s.Config.Encryption.PublicKey,
		s.Config.Encryption.PrivateKey,
		pass,
	)
	if err != nil {
		return err
	}

	return os.RemoveAll(s.entriesPath + ".gpg")
}
