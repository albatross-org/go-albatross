package entries

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func dummyEntry(path, title string) *Entry {
	return &Entry{
		Path:  path,
		Title: title,
	}
}

func TestEntryGraphBasic(t *testing.T) {
	graph := NewEntryGraph()

	entry1 := dummyEntry("food/pizza", "Pizza")
	entry2 := dummyEntry("moods/hunger", "Hunger")
	entry3 := dummyEntry("journal/2020-08-05", "Eating Out")

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

}
