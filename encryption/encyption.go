package encryption

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/albatross-org/go-pgp/pgp"
)

// EncryptDir takes the path to a directory an encrypts it using the public key specified.
// It will write out an encrypted file to newDirPath.
//   gzip -> tar -> pgp
func EncryptDir(dirPath, newDirPath, pathToPublicKey string) error {
	var buf bytes.Buffer

	err := compress(dirPath, &buf)
	if err != nil {
		return fmt.Errorf("error compressing dir at path %s: %w", dirPath, err)
	}

	encypted, err := encrypt(pathToPublicKey, &buf)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(newDirPath, encypted, 0644)
	if err != nil {
		return fmt.Errorf("error writing encrypted file '%s': %w", newDirPath, err)
	}

	return nil
}

// DecryptDir takes the path to an encrypted directory and decrypts it using the private key specified.
// It will write the decrypted directory to newDirPath.
//   pgp -> gzip -> tar
func DecryptDir(dirPath, newDirPath, pathToPublicKey, pathToPrivateKey, password string) error {
	f, err := os.Open(dirPath)
	if err != nil {
		return fmt.Errorf("error reading encrypted directory %s: %w", dirPath, err)
	}

	decrypted, err := decrypt(pathToPublicKey, pathToPrivateKey, password, f)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	_, err = buf.Write(decrypted)
	if err != nil {
		return fmt.Errorf("error writing to buffer: %w", err)
	}

	err = uncompress(&buf, newDirPath)
	if err != nil {
		return fmt.Errorf("error uncompressing decrypted directory %s to %s: %w", dirPath, newDirPath, err)
	}

	return nil
}

func encrypt(publicKeyPath string, src io.Reader) ([]byte, error) {
	publicKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading public key file: %w", err)
	}

	pubEntity, err := pgp.GetEntity(publicKey, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating public key entity: %w", err)
	}

	bytes, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("error reading from src: %w", err)
	}

	encrypted, err := pgp.Encrypt(pubEntity, bytes)
	if err != nil {
		return nil, fmt.Errorf("error encrypting: %w", err)
	}

	return encrypted, nil
}

func decrypt(publicKeyPath, privateKeyPath, password string, src io.Reader) ([]byte, error) {
	publicKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading public key file: %w", err)
	}

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %w", err)
	}

	privEntity, err := pgp.GetEntity(publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating private key entity: %w", err)
	}

	err = privEntity.PrivateKey.Decrypt([]byte(password))
	if err != nil {
		return nil, ErrPrivateKeyDecryptionFailed{PathToPrivateKey: privateKeyPath, Err: err}
	}

	bytes, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("error reading from src: %w", err)
	}

	decrypted, err := pgp.Decrypt(privEntity, bytes)
	if err != nil {
		return nil, fmt.Errorf("error decrypting: %w", err)
	}

	return decrypted, nil
}
