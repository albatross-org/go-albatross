package entries

import (
	"sort"
	"unicode"
)

// List is a list of entries.
type List struct {
	list []*Entry
}

func copyEntrySlice(slice []*Entry) []*Entry {
	return append([]*Entry{}, slice...)
}

// FromOffset returns n entries from a given offset.
// If there aren't enough entries from offset, it will return as many as it can.
// If offset is out of bounds, it will return an ErrListOutOfbounds
func (es List) FromOffset(offset, n int) (List, error) {
	if offset >= len(es.list) {
		return List{}, ErrListOutOfBounds{offset, len(es.list)}
	}

	if offset+n > len(es.list) {
		return List{es.list[offset:]}, nil
	}

	return List{es.list[offset : offset+n]}, nil
}

// First returns the first N entries in the list.
// If there's not N entries, it will return as many as possible.
func (es List) First(n int) List {
	if n > len(es.list) {
		return List{copyEntrySlice(es.list)}
	}

	newList, _ := es.FromOffset(0, n)
	return newList
}

// Last returns the last N entries in the list.
// If there's not N entries, it will return as many as possible.
func (es List) Last(n int) List {
	if n > len(es.list) {
		return List{copyEntrySlice(es.list)}
	}

	newList, _ := es.FromOffset(len(es.list)-n, n)
	return newList
}

// Slice returns the entries as a slice of *Entry.
func (es List) Slice() []*Entry {
	return copyEntrySlice(es.list)
}

// Reverse reverses an entry list.
func (es List) Reverse() List {
	newList := []*Entry{}

	for i := range es.list {
		newList = append(newList, es.list[len(es.list)-i-1])
	}

	return List{newList}
}

// Sort sorts an List.
func (es List) Sort(sortType SortType) List {
	var sortable sort.Interface
	var entries = copyEntrySlice(es.list)

	switch sortType {
	case SortAlpha:
		sortable = SortableByAlpha(entries)
	case SortDate:
		sortable = SortableByDate(entries)
	}

	sort.Sort(sortable)

	return List{list: entries}
}

// SortType is the method used to sort an List.
type SortType int

const (
	// SortAlpha uses alphabetical sorting for entries.
	SortAlpha SortType = iota

	// SortDate uses date sorting for entries.
	SortDate
)

// SortableByAlpha implements sort.Interface for []*Entry based on the alphabetical ordering of titles.
// Courtesy of this StackOverflow answer: https://stackoverflow.com/questions/35076109/in-golang-how-can-i-sort-a-list-of-strings-alphabetically-without-completely-ig
type SortableByAlpha []*Entry

func (es SortableByAlpha) Len() int      { return len(es) }
func (es SortableByAlpha) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
func (es SortableByAlpha) Less(i, j int) bool {
	iRunes := []rune(es[i].Title)
	jRunes := []rune(es[j].Title)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false
}

// SortableByDate implements the sort.Interface for []*Entry based on entry dates.
type SortableByDate []*Entry

func (es SortableByDate) Len() int           { return len(es) }
func (es SortableByDate) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }
func (es SortableByDate) Less(i, j int) bool { return es[i].Date.Before(es[j].Date) }
