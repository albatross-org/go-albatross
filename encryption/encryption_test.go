package encryption

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
)

func tempTestDir(t *testing.T) (path string, cleanup func()) {
	t.Helper()

	tmpDir, err := ioutil.TempDir("", "albatross-encryption-test")
	if err != nil {
		t.Fatalf("could not create temporary directory: %s", err)
	}

	err = copy.Copy("./testdata", filepath.Join(tmpDir, "testdata"))
	if err != nil {
		t.Fatalf("couldn't copy testdata: %s", err)
	}

	return tmpDir, func() {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("could not remove temporary directory: %s", err)
		}
	}
}

func TestEncryptionDecryptionValid(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	err := EncryptDir(
		filepath.Join(dir, "testdata", "example"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public.key"),
	)
	if err != nil {
		t.Fatalf("wasn't expecting error when encrypting: %s", err)
	}

	err = os.RemoveAll(filepath.Join(dir, "testdata", "example"))
	if err != nil {
		t.Fatalf("wasn't expecting error when removing original data: %s", err)
	}

	err = DecryptDir(
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "example-new"),
		filepath.Join(dir, "testdata", "public.key"),
		filepath.Join(dir, "testdata", "private.key"),
		"pa$$word", // super secure, this is my actual password so don't steal it
	)
	if err != nil {
		t.Fatalf("wasn't expecting error when decrypting: %s", err)
	}

	bs, err := ioutil.ReadFile(filepath.Join(dir, "testdata", "example-new", "text.txt"))
	if err != nil {
		t.Fatalf("wasn't expecting error when reading test data file: %s", err)
	}

	if string(bs) != "Hello, I'm some text." {
		t.Fatalf("encrypting then decrypting does not yield the same test, expected=\"Hello, I'm some text\", got=%s", string(bs))
	}
}

func TestEncryptionInvalid(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	err := EncryptDir(
		filepath.Join(dir, "testdata", "folder-that-doesnt-exist"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public.key"),
	)
	if err == nil {
		t.Errorf("expected error when attempting to encrypt non-existant folder: %s", err)
	}

	err = EncryptDir(
		filepath.Join(dir, "testdata", "example"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public-that-doesnt-exist.key"),
	)
	if err == nil {
		t.Errorf("expected error when attempting to encrypt with non-existant public key: %s", err)
	}
}

func TestDecryptionWrongFile(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	err := EncryptDir(
		filepath.Join(dir, "testdata", "example"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public.key"),
	)
	if err != nil {
		t.Fatalf("wasn't expecting error when encrypting: %s", err)
	}

	err = DecryptDir(
		filepath.Join(dir, "testdata", "encrypted-file-that-doesnt-exist.pgp"),
		filepath.Join(dir, "testdata", "example-new"),
		filepath.Join(dir, "testdata", "public.key"),
		filepath.Join(dir, "testdata", "private.key"),
		"pa$$word",
	)
	if err == nil {
		t.Fatalf("expecting error when attempting to decrypt non-existant file: %s", err)
	}
}

func TestDecryptionWrongPublicKey(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	err := EncryptDir(
		filepath.Join(dir, "testdata", "example"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public.key"),
	)
	if err != nil {
		t.Fatalf("wasn't expecting error when encrypting: %s", err)
	}

	err = DecryptDir(
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "example-new"),
		filepath.Join(dir, "testdata", "public-key-that-doesnt-exist.key"),
		filepath.Join(dir, "testdata", "private.key"),
		"pa$$word",
	)
	if err == nil {
		t.Fatalf("expecting error when attempting to decrypt with non-existant public key: %s", err)
	}
}

func TestDecryptionWrongPrivateKey(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	err := EncryptDir(
		filepath.Join(dir, "testdata", "example"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public.key"),
	)
	if err != nil {
		t.Fatalf("wasn't expecting error when encrypting: %s", err)
	}
	err = DecryptDir(
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "example-new"),
		filepath.Join(dir, "testdata", "private-key-that-doesnt-exist.key"),
		filepath.Join(dir, "testdata", "private.key"),
		"pa$$word",
	)
	if err == nil {
		t.Fatalf("expecting error when attempting to decrypt with non-existant private key: %s", err)
	}
}

func TestDecryptionWrongPassword(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	err := EncryptDir(
		filepath.Join(dir, "testdata", "example"),
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "public.key"),
	)
	if err != nil {
		t.Fatalf("wasn't expecting error when encrypting: %s", err)
	}

	err = DecryptDir(
		filepath.Join(dir, "testdata", "example.pgp"),
		filepath.Join(dir, "testdata", "example-new"),
		filepath.Join(dir, "testdata", "public.key"),
		filepath.Join(dir, "testdata", "private-key-that-doesnt-exist.key"),
		"pa$$word-that-is-incorrect",
	)
	if err == nil {
		t.Fatalf("expecting error when attempting to decrypt with incorrect password: %s", err)
	}
}
