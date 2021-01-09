package entries

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DirGraph returns an Collection built from a directory.
// It will return an Collection, a list of errors that occured while parsing entries and finally an error that occured
// when processing the directory or adding an entry.
func DirGraph(path string) (graph *Collection, entryErrs []error, err error) {
	graph = NewCollection()

	// We calculate this offset once to avoid working it every time.
	start := strings.Index(path, "entries")

	// Keeping track of attachments is a little difficult, here we create a map to hold all attachments sorted by the folder they were found in.
	// At the end, we go through all entries found and lookup the folder corresponding to that entry's path and add attachments if any were found.
	var attachmentsMap = map[string][]Attachment{}

	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.Contains(subpath, filepath.Join("entries", ".git")) {
			return filepath.SkipDir
		}

		if !strings.Contains(info.Name(), "entry.md") {
			if start == -1 {
				return fmt.Errorf("error getting relative path for attachment")
			}

			relPath := subpath[start+8:]
			relFolder := filepath.Dir(relPath)
			attachmentsMap[relFolder] = append(attachmentsMap[relFolder], Attachment{
				Name:    info.Name(),
				RelPath: relPath,
				AbsPath: subpath,
			})

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

	for _, entry := range graph.pathMap {
		entry.Attachments = attachmentsMap[entry.Path]
	}

	return graph, entryErrs, nil
}
