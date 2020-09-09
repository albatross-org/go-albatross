package entries
import (
	"strings"
	"time"
)

// Filter is a function which returns true or false depending on whether an entry matches given criteria.
type Filter func(*Entry) bool

// FilterAnd takes multiple filters and creates a new filter. This new filter will only be true if all
// the filters it contains return true, i.e. an AND operation.
func FilterAnd(filters ...Filter) Filter {
	return Filter(func(entry *Entry) bool {
		for _, filter := range filters {
			if !filter(entry) {
				return false
			}
		}

		return true
	})
}

// FilterOr takes multiple filters and creates a new filter. This new filter will be true if any of the
// filters it contains return true, i.e. an OR operation.
func FilterOr(filters ...Filter) Filter {
	return Filter(func(entry *Entry) bool {
		for _, filter := range filters {
			if filter(entry) {
				return true
			}
		}

		return false
	})
}

// FilterNot takes a filter and negates it. For example,
//   FilterNot(FilterPathsMatch)
// will remove all matching paths. If multiple filters are given, they are converted into one using FilterAnd.
func FilterNot(filters ...Filter) Filter {
	return Filter(func(entry *Entry) bool {
		filter := FilterAnd(filters...)
		return !filter(entry)
	})
}

// FilterPathsMatch will allow only entries from the given paths.
// If a path is given which contains entries itself, all the entries inside that path are also allowed.
func FilterPathsMatch(paths ...string) Filter {
	return Filter(func(entry *Entry) bool {
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

// FilterPathsExact will allow only an entry with the given path.
// If a path is given which contains entries itself, only the parent path will be allowed.
func FilterPathsExact(paths ...string) Filter {
	return Filter(func(entry *Entry) bool {
		allowed := false

		for _, path := range paths {
			if entry.Path == path {
				allowed = true
				break
			}
		}

		return allowed
	})
}

// FilterTitlesMatch which match the given titles.
// This function will allow entries where the title given is a substring.
func FilterTitlesMatch(titles ...string) Filter {
	return Filter(func(entry *Entry) bool {
		allowed := false

		for _, title := range titles {
			if strings.Contains(entry.Title, title) {
				allowed = true
				break
			}
		}

		return allowed
	})
}

// FilterTitlesExact which match the given titles.
// This function matches exact titles, not substrings.
func FilterTitlesExact(titles ...string) Filter {
	return Filter(func(entry *Entry) bool {
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

// FilterTags only allows entries with the given tags.
func FilterTags(tags ...string) Filter {
	return Filter(func(entry *Entry) bool {
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

// FilterContentsMatch will allow entries with matching contents (i.e. the content contains one of the substrings specified).
func FilterContentsMatch(substrings ...string) Filter {
	return Filter(func(entry *Entry) bool {
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

// FilterContentsExact will allow entries with matching contents (i.e. the content contains one of the substrings specified).
func FilterContentsExact(contents ...string) Filter {
	return Filter(func(entry *Entry) bool {
		allowed := false
		for _, content := range contents {
			if entry.Contents == content {
				allowed = true
				break
			}
		}

		return allowed
	})
}

// FilterFrom will remove all entries before the given date.
func FilterFrom(date time.Time) Filter {
	return Filter(func(entry *Entry) bool {
		return !entry.Date.Before(date)
	})
}

// FilterUntil will remove all entries after the given date.
func FilterUntil(date time.Time) Filter {
	return Filter(func(entry *Entry) bool {
		return !entry.Date.After(date)
	})
}

// FilterLength will remove all entries under the given length.
func FilterLength(length int) Filter {
	return Filter(func(entry *Entry) bool {
		return len(entry.Contents) < length
	})
}

// Query represents a high-level specification of what entries should match.
// For options like ContentsMatch, they are specified as a slice of slices. Each sub-slice contains the arguments
// to the call to their matching filter function. So multiple slices will become multiple, seperate filter calls.
// This means options within sub-slices act as OR -- entries matching either criteria will be allowed. The sub-slices
// themselves will act as AND, meaning entries have to match all criteria to be allowed.
//
// Consider the difference between:
//   TitlesMatch: [][]string{{"Pizza", "Bananas"}} // Match any titles with Pizza OR Bananas
//   TitlesMatch: [][]string{{"Pizza"}, {"Bananas"}} // Match any titles with both Pizza AND Bananas
type Query struct {
	From  time.Time
	Until time.Time

	MinLength int
	MaxLength int

	Tags        []string
	TagsExclude []string

	ContentsExact        [][]string
	ContentsMatch        [][]string
	ContentsExactExclude [][]string
	ContentsMatchExclude [][]string

	PathsExact        [][]string
	PathsMatch        [][]string
	PathsExactExclude [][]string
	PathsMatchExclude [][]string

	TitlesExact        [][]string
	TitlesMatch        [][]string
	TitlesExactExclude [][]string
	TitlesMatchExclude [][]string
}

// Filter creates a entries.Filter type for a query.
func (q *Query) Filter() Filter {
	filters := []Filter{}

	if q.From != (time.Time{}) {
		filters = append(filters, FilterFrom(q.From))
	}

	if q.Until != (time.Time{}) {
		filters = append(filters, FilterUntil(q.Until))
	}

	if q.MinLength != 0 {
		filters = append(filters, FilterLength(q.MinLength))
	}

	if q.MaxLength != 0 {
		filters = append(filters, FilterNot(FilterLength(q.MaxLength)))
	}

	for _, c := range q.ContentsMatch {
		filters = append(filters, FilterContentsMatch(c...))
	}
	for _, c := range q.ContentsExact {
		filters = append(filters, FilterContentsExact(c...))
	}
	for _, c := range q.ContentsMatchExclude {
		filters = append(filters, FilterNot(FilterContentsMatch(c...)))
	}
	for _, c := range q.ContentsExactExclude {
		filters = append(filters, FilterNot(FilterContentsExact(c...)))
	}

	for _, c := range q.PathsMatch {
		filters = append(filters, FilterPathsMatch(c...))
	}
	for _, c := range q.PathsExact {
		filters = append(filters, FilterPathsExact(c...))
	}
	for _, c := range q.PathsMatchExclude {
		filters = append(filters, FilterNot(FilterPathsMatch(c...)))
	}
	for _, c := range q.PathsExactExclude {
		filters = append(filters, FilterNot(FilterPathsExact(c...)))
	}

	for _, c := range q.TitlesMatch {
		filters = append(filters, FilterTitlesMatch(c...))
	}
	for _, c := range q.TitlesExact {
		filters = append(filters, FilterTitlesExact(c...))
	}
	for _, c := range q.TitlesMatchExclude {
		filters = append(filters, FilterNot(FilterTitlesMatch(c...)))
	}
	for _, c := range q.TitlesExactExclude {
		filters = append(filters, FilterNot(FilterTitlesExact(c...)))
	}

	return FilterAnd(filters...)
}
