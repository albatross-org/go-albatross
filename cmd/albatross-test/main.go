package main

import (
	"fmt"

	"github.com/albatross-org/go-albatross/encryption"
	"github.com/albatross-org/go-albatross/pkg/core"
	"github.com/sirupsen/logrus"
)

// func main() {
// 	fs := entries.NewBaseFs("/home/olly/.local/share/albatross/default/entries")

// 	graph, entryErrs, err := entries.DirGraph(fs, "")
// 	if err != nil {
// 		logrus.Fatal(err)
// 	}

// 	if len(entryErrs) != 0 {
// 		for _, err := range entryErrs {
// 			logrus.Error(err)
// 		}

// 		logrus.Fatal("wasn't expecting entry errs")
// 	}

// 	g, viz, err := graph.Graph()
// 	if err != nil {
// 		logrus.Fatal(err)
// 	}

// 	err = g.Render(viz, "dot", os.Stdout)
// 	if err != nil {
// 		logrus.Fatal(err)
// 	}
// }

// func main() {
// 	const publicKeyPath = "/home/olly/.config/albatross/keys/public.key"
// 	const privateKeyPath = "/home/olly/.config/albatross/keys/private.key"

// 	const originalPath = "/home/olly/code/go/src/github.com/albatross-org/go-albatross/test/stores/testing.albatross"
// 	const decryptedPath = "/home/olly/code/go/src/github.com/albatross-org/go-albatross/test/stores/testing.albatross.new"
// 	const encryptedPath = "/home/olly/code/go/src/github.com/albatross-org/go-albatross/test/stores/testing.albatross.gpg"

// 	// err := encryption.EncryptDir(originalPath, encryptedPath, publicKeyPath)
// 	err := encryption.DecryptDir(encryptedPath, decryptedPath, publicKeyPath, privateKeyPath)
// 	if err != nil {
// 		logrus.Fatal(err)
// 	}
// }

// Store locations.
var (
	TestStore    = "/home/olly/code/go/src/github.com/albatross-org/go-albatross/pkg/core/testdata/stores/testing.albatross"
	DefaultStore = "/home/olly/.local/share/albatross/default"
)

func main() {
	fmt.Println("Loading", TestStore)
	store, err := core.Load(TestStore)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Creating truffles entry...")
	err = store.Create("food/truffles", `---
title: "Truffles"
date: "2020-08-08 20:00"
---

This is an entry all about truffles. I love truffles so much, but they are a bit pretentious.`)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Attaching truffle photo...")
	err = store.Attach("food/truffles", "/home/olly/code/go/src/github.com/albatross-org/go-albatross/pkg/core/testdata/truffle.jpg")
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Creating truffles sub-entry...")
	err = store.Create("food/truffles/mmmm", `---
title: "mmmm"
date: "2020-08-08 21:39"
---

mmmm? mm. mmmmm.`)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Encrypting store...")
	err = store.Encrypt()
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Decrypting store...")
	err = store.Decrypt(encryption.GetPassword)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Updating truffles entry...")
	err = store.Update("food/truffles", `---
title: "Truffles"
date: "2020-08-08 20:00"
---

This is an entry all about truffles. I love truffles so much, but they are a bit pretentious.

Actually, I've changed my mind about truffles now.`)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Deleting truffles entry...")
	err = store.Delete("food/truffles")
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Deleting truffles sub-entry...")
	err = store.Delete("food/truffles/mmmm")
	if err != nil {
		logrus.Fatal(err)
	}

	// collection, err := store.Collection()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// collection, err = collection.Filter(entries.FilterPathsExlude("journal"))
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// fmt.Printf("\nFound %d entries.\n", collection.Len())

	// g, viz, err := collection.Graph()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// err = g.Render(viz, "dot", os.Stdout)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }
}
