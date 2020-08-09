package core

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/albatross-org/go-albatross/encryption"
	"github.com/spf13/afero"

	. "github.com/stretchr/testify/assert"
)

// getMemFsFromDir
func getMemFsFromDir(dir string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	osFs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	err := afero.Walk(osFs, "", func(subpath string, info os.FileInfo, err error) error {
		// fmt.Println(subpath)
		if info.IsDir() {
			return fs.MkdirAll(subpath, 0755)
		}

		bs, err := afero.ReadFile(osFs, subpath)
		if err != nil {
			return err
		}

		afero.WriteFile(fs, subpath, bs, 0644)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fs, nil
}

func getTestFs() (afero.Fs, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	testdata := os.Getenv("GO_ALBATROSS_TESTDATA")
	if testdata == "" {
		testdata = filepath.Join(pwd, "testdata")
	}

	_, err = os.Stat(testdata)
	if err != nil {
		fmt.Printf("Can't access testdata directory: %s\n", testdata)
		fmt.Println("Either run from the root of the go-albatross repository or set the location with the GO_ALBATROSS_TESTDATA environment variable.")
		os.Exit(1)
	}

	return getMemFsFromDir(testdata)
}

func TestStoreLoad(t *testing.T) {
	store, err := Load("stores/testing.albatross")
	Nil(t, err, "not expecting error when loading test store")

	collection, err := store.Collection()
	if err != nil {
		t.Fatalf("not expecting error when getting collection from test store: %s", err)
	}

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
	err = store.Attach("food/truffles", "truffle.jpg")
	Nil(t, err, "not expecting error when attaching truffles photo entry")
	if err != nil {
		t.Fatalf("not expecting error when getting collection from test store: %s", err)
	}

	t.Log("Creating truffles sub-entry...")
	err = store.Create("food/truffles/history", `---
title: "Truffles History"
date: "2020-08-08 21:39"
---

The word truffle comes from the Latin word “tuber”, which means outgrowth. It dates back to as early as the ancient Egyptians, who held truffles in high esteem and ate them coated in goose fat.`)
	if err != nil {
		t.Fatalf("not expecting error when creating truffles sub entry: %s", err)
	}

	t.Log("Encrypting store...")
	err = store.Encrypt()
	if err != nil {
		t.Fatalf("not expecting error when decrypting store: %s", err)
	}

	t.Log("Decrypting store...")
	err = store.Decrypt(encryption.GetPassword)
	if err != nil {
		t.Fatalf("not expecting error when decrypting store: %s", err)
	}

	t.Log("Updating truffles entry...")
	err = store.Update("food/truffles", `---
title: "Truffles"
date: "2020-08-08 20:00"
---

This is an entry all about truffles. I love truffles so much, but they are a bit pretentious.

Actually, I've changed my mind about truffles now.`)
	if err != nil {
		t.Fatalf("not expecting error when updating truffles entry: %s", err)
	}

	t.Log("Deleting truffles entry...")
	err = store.Delete("food/truffles")
	if err != nil {
		t.Fatalf("not expecting error when deleting truffles entry: %s", err)
	}

	t.Log("Deleting truffles sub-entry...")
	err = store.Delete("food/truffles/mmmm")
	if err != nil {
		t.Fatalf("not expecting error when deleting truffles sub entry: %s", err)
	}

	fmt.Println(collection.Len())
}

func TestMain(m *testing.M) {
	fs, err := getTestFs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Fs = fs // override the default OS filesystem backend to use the test one.
	encryption.Fs = fs

	os.Exit(m.Run())
}
