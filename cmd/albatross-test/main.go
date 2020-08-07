package main

import (
	"os"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/sirupsen/logrus"
)

func main() {
	fs := entries.NewBaseFs("./test/stores/testing.albatross/entries")

	graph, entryErrs, err := entries.DirGraph(fs, "")
	if err != nil {
		logrus.Fatal(err)
	}

	if len(entryErrs) != 0 {
		for _, err := range entryErrs {
			logrus.Error(err)
		}

		logrus.Fatal("wasn't expecting entry errs")
	}

	g, viz, err := graph.Graph()
	if err != nil {
		logrus.Fatal(err)
	}

	err = g.Render(viz, "dot", os.Stdout)
	if err != nil {
		logrus.Fatal(err)
	}
}
