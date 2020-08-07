package entries

import (
	"os"
	"strings"

	"github.com/spf13/afero"
)

var (
	// Os points to the real OS filesystem.
	Os = &afero.OsFs{}
)

// NewBaseFs returns a new file system for an albatross folder.
// It is read-only (to prevent accidental side-effects) and deals with relative paths behind the scenes.
// For example, "food/pizza" will be converted into "/path/to/albatross/entries/food/pizza"
func NewBaseFs(basePath string) afero.Fs {
	return afero.NewBasePathFs(afero.NewReadOnlyFs(Os), basePath)
}

// DirGraph returns an Collection built from a directory in a file system.
// It will return an Collection, a list of errors that occured while parsing entries and finally an error that occured
// when processing the directory or adding an entry.
func DirGraph(fs afero.Fs, path string) (graph *Collection, entryErrs []error, err error) {
	graph = NewCollection()

	err = afero.Walk(fs, path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.Contains(info.Name(), "entry.md") {
			return nil
		}

		entry, entryErr := NewEntryFromFile(fs, subpath)
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

	return graph, nil, nil
}
