package entries

import (
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
)

func TestListFromOffset(t *testing.T) {
	entry1 := dummyEntry("food/pizza", "Pizza", "Pizza is great.")
	entry2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is amazing.")
	entry3 := dummyEntry("food/beans", "BEANS!", "BEANS!!!")
	entry4 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entry5 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")
	entry6 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	list := List{[]*Entry{entry1, entry2, entry3, entry4, entry5, entry6}}

	ts := []struct {
		offset, n int
		expected  []*Entry

		err error
	}{
		{
			0, 3,
			[]*Entry{entry1, entry3, entry3},
			nil,
		},
		{
			1, 2,
			[]*Entry{entry2, entry3},
			nil,
		},
		{
			3, 1000,
			[]*Entry{entry4, entry5, entry6},
			nil,
		},
		{
			6, 1000,
			[]*Entry{},
			ErrListOutOfBounds{6, 6},
		},
		{
			12, 1000,
			[]*Entry{},
			ErrListOutOfBounds{12, 6},
		},
	}

	for _, tc := range ts {
		newList, err := list.FromOffset(tc.offset, tc.n)
		if err != tc.err {
			t.Fatalf("errors don't match for from offset, got=%s", err)
		}

		if len(tc.expected) != len(newList.list) {
			t.Fatalf("new offset list doesn't match, expected len=%d, god len=%d", len(tc.expected), len(newList.list))
		}

		Equal(t, len(tc.expected), len(newList.list))
	}
}

func TestListFirstLast(t *testing.T) {
	entry1 := dummyEntry("food/pizza", "Pizza", "Pizza is great.")
	entry2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is amazing.")
	entry3 := dummyEntry("food/beans", "BEANS!", "BEANS!!!")
	entry4 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entry5 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")
	entry6 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	list := List{[]*Entry{entry1, entry2, entry3, entry4, entry5, entry6}}

	newList := list.First(5).
		Last(2)

	if len(newList.list) != 2 {
		t.Fatalf("new length of entry list expected=2, got=%d", len(newList.list))
	}

	Equal(t, entry4, newList.Slice()[0], "first entry in entry list should be entry4")
	Equal(t, entry5, newList.Slice()[1], "seconed entry in entry list should be entry5")
}

func TestListSortAlpha(t *testing.T) {
	entry1 := dummyEntry("food/pizza", "Pizza", "Pizza is great.")               // 3
	entry2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is amazing.") // 2
	entry3 := dummyEntry("food/beans", "BEANS!", "BEANS!!!")                     // 1
	entry4 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")      // 5
	entry5 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")       // 6
	entry6 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.") // 4

	list := List{[]*Entry{entry1, entry2, entry3, entry4, entry5, entry6}}
	sortedList := list.Sort(SortAlpha)

	Equal(t, entry3, sortedList.Slice()[0], "alphabetical sort should have entry3 1st")
	Equal(t, entry2, sortedList.Slice()[1], "alphabetical sort should have entry2 2nd")
	Equal(t, entry1, sortedList.Slice()[2], "alphabetical sort should have entry1 3rd")
	Equal(t, entry6, sortedList.Slice()[3], "alphabetical sort should have entry6 4th")
	Equal(t, entry4, sortedList.Slice()[4], "alphabetical sort should have entry4 5th")
	Equal(t, entry5, sortedList.Slice()[5], "alphabetical sort should have entry5 6th")
}

func TestListSortDate(t *testing.T) {
	entry1 := &Entry{Path: "food/pizza", Date: time.Date(2017, time.January, 0, 0, 0, 0, 0, &time.Location{})}        // 3
	entry2 := &Entry{Path: "food/ice-cream", Date: time.Date(2016, time.January, 0, 0, 0, 0, 0, &time.Location{})}    // 2
	entry3 := &Entry{Path: "food/beans", Date: time.Date(2015, time.January, 0, 0, 0, 0, 0, &time.Location{})}        // 1
	entry4 := &Entry{Path: "animals/tiger", Date: time.Date(2020, time.January, 0, 0, 0, 0, 0, &time.Location{})}     // 6
	entry5 := &Entry{Path: "animals/whale", Date: time.Date(2019, time.January, 0, 0, 0, 0, 0, &time.Location{})}     // 5
	entry6 := &Entry{Path: "plants/sunflowers", Date: time.Date(2018, time.January, 0, 0, 0, 0, 0, &time.Location{})} // 4

	list := List{[]*Entry{entry1, entry2, entry3, entry4, entry5, entry6}}
	sortedList := list.Sort(SortDate)

	Equal(t, entry3, sortedList.Slice()[0], "alphabetical sort should have entry3 1st")
	Equal(t, entry2, sortedList.Slice()[1], "alphabetical sort should have entry2 2nd")
	Equal(t, entry1, sortedList.Slice()[2], "alphabetical sort should have entry1 3rd")
	Equal(t, entry6, sortedList.Slice()[3], "alphabetical sort should have entry6 4th")
	Equal(t, entry5, sortedList.Slice()[4], "alphabetical sort should have entry4 5th")
	Equal(t, entry4, sortedList.Slice()[5], "alphabetical sort should have entry5 6th")
}

func TestListReverse(t *testing.T) {
	entry1 := dummyEntry("food/pizza", "Pizza", "Pizza is great.")
	entry2 := dummyEntry("food/ice-cream", "Ice Cream", "Ice cream is amazing.")
	entry3 := dummyEntry("food/beans", "BEANS!", "BEANS!!!")
	entry4 := dummyEntry("animals/tiger", "Tigers", "Love me some tigers.")
	entry5 := dummyEntry("animals/whale", "Whales", "Whales. Oh, Whales!")
	entry6 := dummyEntry("plants/sunflowers", "Sunflowers", "Pretty and sunny.")

	list := List{[]*Entry{entry1, entry2, entry3, entry4, entry5, entry6}}
	reversedList := list.Reverse()

	Equal(t, entry6, reversedList.Slice()[0], "alphabetical sort should have entry6 1st")
	Equal(t, entry5, reversedList.Slice()[1], "alphabetical sort should have entry5 2nd")
	Equal(t, entry4, reversedList.Slice()[2], "alphabetical sort should have entry4 3rd")
	Equal(t, entry3, reversedList.Slice()[3], "alphabetical sort should have entry3 4th")
	Equal(t, entry2, reversedList.Slice()[4], "alphabetical sort should have entry2 5th")
	Equal(t, entry1, reversedList.Slice()[5], "alphabetical sort should have entry1 6th")
}
