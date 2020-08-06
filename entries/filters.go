package entries

import (
	"strings"
	"time"
)

// Filter is a function which filters an EntryGraph.
type Filter func(*EntryGraph) error

// FilterEntryAllower takes a function which returns true or false depending on whether the entry is allowd.
func FilterEntryAllower(allower func(*Entry) bool) Filter {
	return func(graph *EntryGraph) error {
		remove := []*Entry{}

		for _, entry := range graph.pathMap {
			if !allower(entry) {
				remove = append(remove, entry)
			}
		}

		for _, entry := range remove {
			err := graph.Delete(entry)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// FilterPathsInclude will allow only entries from the given paths.
func FilterPathsInclude(paths ...string) Filter {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := false
		for _, path := range paths {
			if strings.HasPrefix(entry.Path, path) {
				allowed = true
				break
			}
		}

		return allowed
	})
}

// FilterPathsExlude will remove all entries from the given paths.
func FilterPathsExlude(paths ...string) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := true
		for _, path := range paths {
			if strings.HasPrefix(entry.Path, path) {
				allowed = false
				break
			}
		}

		return allowed
	})
}

// FilterTitlesInclude only allows entries with the given titles.
// This function matches full titles, not a substring.
func FilterTitlesInclude(titles ...string) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := false
		for _, title := range titles {
			if entry.Title == title {
				allowed = true
				break
			}
		}

		return allowed
	})
}

// FilterTitlesExclude only allows entries that don't have the specified titles.
// This function matches full titles, not a substring.
func FilterTitlesExclude(titles ...string) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := true
		for _, title := range titles {
			if entry.Title == title {
				allowed = false
				break
			}
		}

		return allowed
	})
}

// FilterTagsInclude only allows entries with the given tags.
func FilterTagsInclude(tags ...string) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := false
		for _, tag := range tags {
			for _, entryTag := range entry.Tags {
				if entryTag == tag {
					allowed = true
					break
				}
			}
		}

		return allowed
	})
}

// FilterTagsExclude only allows entries that don't have the specified tags.
func FilterTagsExclude(tags ...string) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := true
		for _, tag := range tags {
			for _, entryTag := range entry.Tags {
				if entryTag == tag {
					allowed = false
					break
				}
			}
		}

		return allowed
	})
}

// FilterMatchInclude will allow entries with the given substrings.
func FilterMatchInclude(substrings ...string) Filter {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := false
		for _, substring := range substrings {
			if strings.Contains(entry.Contents, substring) {
				allowed = true
				break
			}
		}

		return allowed
	})
}

// FilterMatchExclude will allow entries that don't have the specified substrings.
func FilterMatchExclude(substrings ...string) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		allowed := true
		for _, substring := range substrings {
			if strings.Contains(entry.Contents, substring) {
				allowed = false
				break
			}
		}

		return allowed
	})
}

// FilterFrom will remove all entries before the given date.
func FilterFrom(date time.Time) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		return !entry.Date.Before(date)
	})
}

// FilterUntil will remove all entries after the given date.
func FilterUntil(date time.Time) func(*EntryGraph) error {
	return FilterEntryAllower(func(entry *Entry) bool {
		return !entry.Date.After(date)
	})
}
