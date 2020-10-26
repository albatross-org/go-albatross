package entries

import (
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
)

func dummyEntry(path, title, content string) *Entry {
	return &Entry{
		Path:     path,
		Title:    title,
		Contents: content,
	}
}

func TestCollectionBasic(t *testing.T) {
	collection := NewCollection()

	entry1 := dummyEntry("food/pizza", "Pizza", "")
	entry2 := dummyEntry("moods/hunger", "Hunger", "")
	entry3 := dummyEntry("journal/2020-08-05", "Eating Out", "")

	err := collection.Add(entry1)
	Nil(t, err, "adding entry1, err should be nil")
	True(t, collection.In(entry1), "entry1 should be in collection")

	err = collection.Add(entry2)
	Nil(t, err, "adding entry2, err should be nil")
	True(t, collection.In(entry2), "entry2 should be in collection")

	Equal(t, 2, collection.Len(), "there should be two entries in the collection")

	False(t, collection.In(entry3), "entry3 should not be in collection")

	err = collection.Add(entry1)
	Equal(t, ErrEntryAlreadyExists{Path: entry1.Path, Title: entry1.Title}, err, "adding entry1 again should result in ErrEntryAlreadyExists")

	err = collection.Delete(entry3)
	Equal(t, ErrEntryDoesntExist{Path: entry3.Path, Title: entry3.Title}, err, "removing non-existant entry3 should result in ErrDoesntAlreadyExist")

	err = collection.Delete(entry2)
	Nil(t, err, "deleting entry2, err should be nil")
	False(t, collection.In(entry2), "entry2 should not be in collection")
	Equal(t, 1, collection.Len(), "there should be one entry in the collection")

	err = collection.Delete(entry1)
	Nil(t, err, "deleting entry1, err should be nil")
	False(t, collection.In(entry1), "entry1 should not be in collection")
	Equal(t, 0, collection.Len(), "there should be no entries in the collection")
}

func TestCollectionLinks(t *testing.T) {
	collection := NewCollection()

	pizzaEntry := &Entry{
		Path:     "food/pizza",
		Title:    "Pizza",
		Contents: "I feel {{moods/hunger}(Hungry)} when I don't eat pizza.",
		OutboundLinks: []Link{
			{
				Path:  "moods/hunger",
				Title: "",
				Name:  "Hungry",
				Type:  LinkPathWithName,
			},
		},
	}
	pizzaEntry.OutboundLinks[0].Parent = pizzaEntry

	hungerEntry := &Entry{
		Path:     "moods/hunger",
		Title:    "Hunger",
		Contents: "This is an entry all about the mood hunger.",
	}

	err := collection.Add(pizzaEntry)
	Nil(t, err, "adding pizzaEntry, err should be nil")

	err = collection.Add(hungerEntry)
	Nil(t, err, "adding hungerEntry, err should be nil")

	Equal(t, 1, len(collection.FindLinksTo(hungerEntry)), "hungerEntry should have one inbound link")
	Equal(t, pizzaEntry.Path, collection.FindLinksTo(hungerEntry)[0].Parent.Path, "hungerEntry should have an inbound link from pizzaEntry")

	err = collection.Delete(pizzaEntry)
	Nil(t, err, "removing pizzaEntry, err should be nil")

	Equal(t, 0, len(collection.FindLinksTo(hungerEntry)), "hungerEntry should still have no inbound links after pizza entry was removed")
}

func TestCollectionFilterPaths(t *testing.T) {
	collection := NewCollection()

	entryFood1 := dummyEntry("food/pizza", "Pizza", "Pizza is great.")
	entryFood2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is amazing.")
	entryFood3 := dummyEntry("food/beans", "BEANS!", "BEANS!!!")

	entryAnimals1 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entryAnimals2 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")

	entryPlants1 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	err := collection.AddMany(entryFood1, entryFood2, entryFood3, entryAnimals1, entryAnimals2, entryPlants1)
	Nil(t, err, "adding all entries, err should be nil")
	Equal(t, 6, collection.Len(), "there should be 6 entries in the collection")

	collectionFood, err := collection.Filter(FilterPathsMatch("food/"))
	Nil(t, err, "filtering for food only, err should be nil")
	Equal(t, 3, collectionFood.Len(), "there should be 3 entries in the food-only collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")

	collectionNoAnimals, err := collection.Filter(FilterNot(FilterPathsMatch("animals/")))
	Nil(t, err, "filtering for anyhting but animals, err should be nil")
	Equal(t, 4, collectionNoAnimals.Len(), "there should be 4 entries in the no animals collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")

	collectionNoPlants, err := collection.Filter(FilterPathsMatch("food/", "animals/"))
	Nil(t, err, "filtering for anything but plants, err should be nil")
	Equal(t, 5, collectionNoPlants.Len(), "there should be 5 entries in the no animals collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")
}

func TestCollectionFilterTitles(t *testing.T) {
	collection := NewCollection()

	entryFood1 := dummyEntry("food/pizza", "Pizza", "Pizza is food, and great.")
	entryFood2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is food, and amazing.")
	entryFood3 := dummyEntry("food/beans", "BEANS!", "BEANS are food!!!")

	entryAnimals1 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entryAnimals2 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")

	entryPlants1 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	err := collection.AddMany(entryFood1, entryFood2, entryFood3, entryAnimals1, entryAnimals2, entryPlants1)
	Nil(t, err, "adding all entries, err should be nil")
	Equal(t, 6, collection.Len(), "there should be 6 entries in the collection")

	collectionPizzaOnly, err := collection.Filter(FilterTitlesMatch("Pizza"))
	Nil(t, err, "filtering for entries called 'Pizza', err should be nil")
	Equal(t, 1, collectionPizzaOnly.Len(), "there should be 1 entries in the pizza-only collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")

	collectionNoPizza, err := collection.Filter(FilterNot(FilterTitlesMatch("Pizza")))
	Nil(t, err, "filtering for everything not called 'Pizza', err should be nil")
	Equal(t, 5, collectionNoPizza.Len(), "there should be 5 entries in the no-pizza collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")
}

func TestCollectionFilterTags(t *testing.T) {
	collection := NewCollection()

	entryFood1 := &Entry{Path: "food/pizza", Title: "Pizza", Tags: []string{"food", "warm"}}
	entryFood2 := &Entry{Path: "food/ice-cream", Title: "Ice Cream", Tags: []string{"food", "cold"}}
	entryFood3 := &Entry{Path: "food/beans", Title: "BEANS!", Tags: []string{"food", "warm"}}

	err := collection.AddMany(entryFood1, entryFood2, entryFood3)
	Nil(t, err, "adding all entries, err should be nil")
	Equal(t, 3, collection.Len(), "there should be 3 entries in the collection")

	collectionWarmOnly, err := collection.Filter(FilterTags("warm"))
	Nil(t, err, "filtering for entries warm food only, err should be nil")
	Equal(t, 2, collectionWarmOnly.Len(), "there should be 2 entries in the warm-only collection")
	Equal(t, 3, collection.Len(), "there should be stll be 3 entries in the orignal collection after filter")

	collectionColdOnly, err := collection.Filter(FilterTags("cold"))
	Nil(t, err, "filtering for entries that are cold food only, err should be nil")
	Equal(t, 1, collectionColdOnly.Len(), "there should be 1 entries in the cold-only collection")
	Equal(t, 3, collection.Len(), "there should be stll be 3 entries in the orignal collection after filter")

	collectionColdOnlyAlt, err := collection.Filter(FilterNot(FilterTags("warm")))
	Nil(t, err, "filtering for entries that are cold food only by excluding warm ones, err should be nil")
	Equal(t, 1, collectionColdOnlyAlt.Len(), "there should be 1 entries in the cold-only collection")
	Equal(t, 3, collection.Len(), "there should be stll be 3 entries in the orignal collection after filter")
}

func TestCollectionFilterMatch(t *testing.T) {
	collection := NewCollection()

	entryFood1 := dummyEntry("food/pizza", "Pizza", "Pizza is food, and great.")
	entryFood2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is food, and amazing.")
	entryFood3 := dummyEntry("food/beans", "BEANS!", "BEANS are food!!!")

	entryAnimals1 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entryAnimals2 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")

	entryPlants1 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	err := collection.AddMany(entryFood1, entryFood2, entryFood3, entryAnimals1, entryAnimals2, entryPlants1)
	Nil(t, err, "adding all entries, err should be nil")
	Equal(t, 6, collection.Len(), "there should be 6 entries in the collection")

	collectionContainsIs, err := collection.Filter(FilterContentsMatch("is"))
	Nil(t, err, "filtering for entries containing 'is', err should be nil")
	Equal(t, 2, collectionContainsIs.Len(), "there should be 2 entries in the containing-is collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")

	collectionComplexFilter, err := collection.Filter(FilterContentsMatch("is", "Whales"), FilterNot(FilterContentsMatch("Pizza")))
	Nil(t, err, "filtering for everything containing 'is' or 'Whales' but not 'Pizza', err should be nil")
	Equal(t, 2, collectionComplexFilter.Len(), "there should be 2 entries in the no-pizza collection")
	Equal(t, 6, collection.Len(), "there should be stll be 6 entries in the orignal collection after filter")
}

func TestCollectionFilterDates(t *testing.T) {
	collection := NewCollection()

	entry1 := &Entry{Path: "grief/denial", Title: "Denial", Date: time.Date(2016, time.January, 1, 0, 0, 0, 0, &time.Location{})}
	entry2 := &Entry{Path: "grief/anger", Title: "Anger", Date: time.Date(2017, time.January, 1, 0, 0, 0, 0, &time.Location{})}
	entry3 := &Entry{Path: "grief/bargaining", Title: "Bargaining", Date: time.Date(2018, time.January, 1, 0, 0, 0, 0, &time.Location{})}
	entry4 := &Entry{Path: "grief/depression", Title: "depression", Date: time.Date(2019, time.January, 1, 0, 0, 0, 0, &time.Location{})}
	entry5 := &Entry{Path: "grief/acceptance", Title: "Acceptance", Date: time.Date(2020, time.January, 1, 0, 0, 0, 0, &time.Location{})}

	err := collection.AddMany(entry1, entry2, entry3, entry4, entry5)
	Nil(t, err, "adding all entries, err should be nil")
	Equal(t, 5, collection.Len(), "there should be 5 entries in the collection")

	collectionNoDenial, err := collection.Filter(FilterFrom(time.Date(2017, time.January, 1, 0, 0, 0, 0, &time.Location{})))
	Nil(t, err, "filtering for entries after 2016, err should be nil")
	Equal(t, 4, collectionNoDenial.Len(), "there should be 4 entries in the after-2017 collection")
	Equal(t, 5, collection.Len(), "there should be stll be 5 entries in the orignal collection after filter")

	collectionNoAcceptance, err := collection.Filter(FilterUntil(time.Date(2019, time.January, 1, 0, 0, 0, 0, &time.Location{})))
	Nil(t, err, "filtering for entries after before 2020, err should be nil")
	Equal(t, 4, collectionNoAcceptance.Len(), "there should be 4 entries in the before-2020 collection")
	Equal(t, 5, collection.Len(), "there should be stll be 5 entries in the orignal collection after filter")

	collectionAngerToDepression, err := collection.Filter(
		FilterFrom(time.Date(2017, time.January, 1, 0, 0, 0, 0, &time.Location{})),
		FilterUntil(time.Date(2019, time.January, 1, 0, 0, 0, 0, &time.Location{})),
	)

	Nil(t, err, "filtering for entries from anger to depression, err should be nil")
	Equal(t, 3, collectionAngerToDepression.Len(), "there should be 3 entries in the anger to depression collection")
	Equal(t, 5, collection.Len(), "there should be stll be 5 entries in the orignal collection after filter")
}
