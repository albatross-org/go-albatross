package entries

import (
	"os"
	"path/filepath"
	"strings"
)

// DirGraph returns an Collection built from a directory .
// It will return an Collection, a list of errors that occured while parsing entries and finally an error that occured
// when processing the directory or adding an entry.
func DirGraph(path string) (graph *Collection, entryErrs []error, err error) {
	graph = NewCollection()

	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.Contains(info.Name(), "entry.md") {
			return nil
		}

		entry, entryErr := NewEntryFromFile(subpath)
		if entryErr != nil {
			entryErrs = append(entryErrs, entryErr)
			return nil
		}

		err = graph.Add(entry)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, entryErrs, err
	}

	return graph, entryErrs, nil
}
