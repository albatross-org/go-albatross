package albatross

import (
	"fmt"

	"github.com/albatross-org/go-albatross/entries"
)

// Get gets a specific entry with the given path.
// For manipulating more than a couple entries, it's recommended to use s.Collection() and work with an entries.Collection instead.
func (s *Store) Get(path string) (*entries.Entry, error) {
	encrypted, err := s.Encrypted()
	if err != nil {
		return nil, err
	} else if encrypted {
		return nil, ErrStoreEncrypted{Path: s.Path}
	}

	filtered, err := s.coll.Filter(entries.FilterPathsExact(path))
	if err != nil {
		return nil, err
	}

	if filtered.Len() == 0 {
		return nil, ErrEntryDoesntExist{Path: path}
	}

	if filtered.Len() > 1 {
		return nil, fmt.Errorf("get matched more than one entry, somehow")
	}

	return filtered.List().Slice()[0], nil
}
