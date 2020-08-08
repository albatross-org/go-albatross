package main

import (
	"os"

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

func main() {
	store, err := core.Load("/home/olly/.local/share/albatross/default")
	if err != nil {
		logrus.Fatal(err)
	}

	g, viz, err := store.Collection.Graph()
	if err != nil {
		logrus.Fatal(err)
	}

	err = g.Render(viz, "dot", os.Stdout)
	if err != nil {
		logrus.Fatal(err)
	}
}
