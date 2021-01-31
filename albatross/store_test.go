package albatross

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"

	. "github.com/stretchr/testify/assert"
)

func tempTestDir(t *testing.T) (path string, cleanup func()) {
	t.Helper()

	tmpDir, err := ioutil.TempDir("", "albatross-core-test")
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

func staticPassword(pass string) func() (string, error) {
	return func() (string, error) {
		return pass, nil
	}
}

func TestStoreFull(t *testing.T) {
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
	err = store.AttachCopy("food/truffles", filepath.Join(dir, "testdata", "truffle.jpg"))
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
	err = store.Decrypt(staticPassword("pa$$word"))
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
	err = store.Delete("food/truffles/history")
	if err != nil {
		t.Fatalf("not expecting error when deleting truffles sub entry: %s", err)
	}
}
