package albatross

import (
	"path/filepath"
	"testing"

	. "github.com/stretchr/testify/assert"
)

func getTestStore(dir string) (*Store, error) {
	config := NewConfig()
	config.Path = filepath.Join(dir, "testdata", "stores", "testing.albatross")
	config.Encryption.PrivateKey = filepath.Join(dir, "testdata", "keys", "private.key")
	config.Encryption.PublicKey = filepath.Join(dir, "testdata", "keys", "public.key")

	return FromConfig(config)
}

func TestAttachHashing(t *testing.T) {
	dir, cleanup := tempTestDir(t)

	hash, err := hashPath(filepath.Join(dir, "testdata", "stores", "testing.albatross"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("hash for test folder: %s", hash)

	hash, err = hashPath(filepath.Join(dir, "testdata", "stores", "testing.albatross", "entries", "food", "ice-cream", "entry.md"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("hash for test entry: %s", hash)

	defer cleanup()
}

func TestAttachSymlink(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	t.Logf("Temporary test dir: %s", dir)

	store, err := getTestStore(dir)
	Nil(t, err, "not expecting error when loading test store")

	t.Log("Creating truffles entry...")
	err = store.Create("food/truffles", `---
title: "Truffles"
date: "2020-08-08 20:00"
---

This is an entry all about truffles. I love truffles so much, but they are a bit pretentious.`)
	Nil(t, err, "not expecting error when creating truffles entry")
	if err != nil {
		t.Fatalf("not expecting error when getting collection from test store: %s", err)
	}

	t.Log("Attaching truffle photo...")
	err = store.AttachSymlink("food/truffles", filepath.Join(dir, "testdata", "truffle.jpg"))
	Nil(t, err, "not expecting error when attaching truffles photo entry")
	if err != nil {
		t.Fatalf("not expecting error when getting collection from test store: %s", err)
	}
}

func TestAttachSymlinkFolder(t *testing.T) {
	dir, cleanup := tempTestDir(t)
	defer cleanup()

	t.Logf("Temporary test dir: %s", dir)

	store, err := getTestStore(dir)
	Nil(t, err, "not expecting error when loading test store")

	t.Log("Creating truffles entry...")
	err = store.Create("food/truffles", `---
title: "Truffles"
date: "2020-08-08 20:00"
---

This is an entry all about truffles. I love truffles so much, but they are a bit pretentious.`)
	Nil(t, err, "not expecting error when creating truffles entry")
	if err != nil {
		t.Fatalf("not expecting error when getting collection from test store: %s", err)
	}

	t.Log("Attaching photos folder...")
	err = store.AttachSymlink("food/truffles", filepath.Join(dir, "testdata", "photos"))
	Nil(t, err, "not expecting error when attaching photos")
	if err != nil {
		t.Fatalf("not expecting error when getting collection from test store: %s", err)
	}
}
