package main

import (
	"os"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/sirupsen/logrus"
)

func main() {
	fs := entries.NewBaseFs("./test/stores/testing.albatross/entries")

	entryPizza, err := entries.NewEntry(fs, "food/pizza")
	if err != nil {
		logrus.Fatal(err)
	}

	entryIceCream, err := entries.NewEntry(fs, "food/ice-cream")
	if err != nil {
		logrus.Fatal(err)
	}

	entryHunger, err := entries.NewEntry(fs, "moods/hunger")
	if err != nil {
		logrus.Fatal(err)
	}

	graph := entries.NewEntryGraph()

	err = graph.AddMany(entryPizza, entryIceCream, entryHunger)
	if err != nil {
		logrus.Fatal(err)
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
