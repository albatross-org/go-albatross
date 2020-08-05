package entries

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func dummyEntry(path, title, content string) *Entry {
	return &Entry{
		Path:     path,
		Title:    title,
		Contents: content,
	}
}

func TestEntryGraphBasic(t *testing.T) {
	graph := NewEntryGraph()

	entry1 := dummyEntry("food/pizza", "Pizza", "")
	entry2 := dummyEntry("moods/hunger", "Hunger", "")
	entry3 := dummyEntry("journal/2020-08-05", "Eating Out", "")

	err := graph.Add(entry1)
	Nil(t, err, "adding entry1, err should be nil")
	True(t, graph.In(entry1), "entry1 should be in graph")

	err = graph.Add(entry2)
	Nil(t, err, "adding entry2, err should be nil")
	True(t, graph.In(entry2), "entry2 should be in graph")

	Equal(t, 2, graph.Len(), "there should be two entries in the graph")

	False(t, graph.In(entry3), "entry3 should not be in graph")

	err = graph.Add(entry1)
	Equal(t, ErrEntryAlreadyExists{Path: entry1.Path, Title: entry1.Title}, err, "adding entry1 again should result in ErrEntryAlreadyExists")

	err = graph.Delete(entry3)
	Equal(t, ErrEntryDoesntExist{Path: entry3.Path, Title: entry3.Title}, err, "removing non-existant entry3 should result in ErrDoesntAlreadyExist")

	err = graph.Delete(entry2)
	Nil(t, err, "deleting entry2, err should be nil")
	False(t, graph.In(entry2), "entry2 should not be in graph")
	Equal(t, 1, graph.Len(), "there should be one entry in the graph")

	err = graph.Delete(entry1)
	Nil(t, err, "deleting entry1, err should be nil")
	False(t, graph.In(entry1), "entry1 should not be in graph")
	Equal(t, 0, graph.Len(), "there should be no entries in the graph")
}

func TestEntryGraphLinks(t *testing.T) {
	graph := NewEntryGraph()

	pizzaEntry := &Entry{
		Path:     "food/pizza",
		Title:    "Pizza",
		Contents: "I feel {{moods/hunger}(Hungry)} when I don't eat pizza.",
		OutboundLinks: []Link{
			Link{
				Path:  "moods/hunger",
				Title: "",
				Name:  "Hungry",
				Type:  LinkPathWithName,
			},
		},
	}

	hungerEntry := &Entry{
		Path:     "moods/hunger",
		Title:    "Hunger",
		Contents: "This is an entry all about the mood hunger.",
	}

	err := graph.Add(pizzaEntry)
	Nil(t, err, "adding pizzaEntry, err should be nil")

	err = graph.Add(hungerEntry)
	Nil(t, err, "adding hungerEntry, err should be nil")

	Equal(t, 1, len(hungerEntry.InboundLinks), "hungerEntry should have one inbound link")
	Equal(t, pizzaEntry.Path, hungerEntry.InboundLinks[0].Path, "hungerEntry should have an inbound link from pizzaEntry")

	err = graph.Delete(pizzaEntry)
	Nil(t, err, "removing pizzaEntry, err should be nil")

	Equal(t, 1, len(hungerEntry.InboundLinks), "hungerEntry should still have an inbound links for pizzaEntry after it was removed")
	Equal(t, pizzaEntry.Path, hungerEntry.InboundLinks[0].Path, "hungerEntry should still have an inbound links for pizzaEntry after it was removed")

	graph.RemoveInboundLinks(pizzaEntry)

	Equal(t, 0, len(hungerEntry.InboundLinks), "hungerEntry should have no inbound links after pizzaEntry's inbound links were removed")
}

func TestEntryGraphFilterPath(t *testing.T) {
	graph := NewEntryGraph()

	entryFood1 := dummyEntry("food/pizza", "Pizza", "Pizza is great.")
	entryFood2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is amazing.")
	entryFood3 := dummyEntry("food/beans", "BEANS!", "BEANS!!!")

	entryAnimals1 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entryAnimals2 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")

	entryPlants1 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	err := graph.AddMany(entryFood1, entryFood2, entryFood3, entryAnimals1, entryAnimals2, entryPlants1)
	Nil(t, err, "adding all entries, err should be nil")
	Equal(t, 6, graph.Len(), "there should be 5 entries in the graph")

	graphFood, err := graph.Filter(FilterPathsInclude("food/"))
	Nil(t, err, "filtering for food only, err should be nil")
	Equal(t, 3, graphFood.Len(), "there should be 3 entries in the food-only graph")
	Equal(t, 6, graph.Len(), "there should be stll be 6 entries in the orignal graph after filter")

	graphNoAnimals, err := graph.Filter(FilterPathsExlude("animals/"))
	Nil(t, err, "filtering for anyhting but animals, err should be nil")
	Equal(t, 4, graphNoAnimals.Len(), "there should be 4 entries in the no animals graph")
	Equal(t, 6, graph.Len(), "there should be stll be 6 entries in the orignal graph after filter")

	graphNoPlants, err := graph.Filter(FilterPathsInclude("food/", "animals/"))
	Nil(t, err, "filtering for anything but plants, err should be nil")
	Equal(t, 5, graphNoPlants.Len(), "there should be 5 entries in the no animals graph")
	Equal(t, 6, graph.Len(), "there should be stll be 6 entries in the orignal graph after filter")
}

// TODO: write tests for the rest of the filters
