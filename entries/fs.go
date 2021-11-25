package entries

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FromDirectory returns an Collection built from a directory.
// It will return an Collection, a list of errors that occured while parsing entries and finally an error that occured
// when processing the directory or adding an entry.
func FromDirectory(path, dateLayout, tagPrefix string) (graph *Collection, entryErrs []error, err error) {
	graph = NewCollection()

	// We calculate a couple things up here once to avoid calculating it for every file.
	start := strings.Index(path, "entries")
	gitDir := filepath.Join("entries", ".git")

	// Create a parser that can be reused for each entry.
	parser, err := NewParser(dateLayout, tagPrefix)
	if err != nil {
		return nil, nil, err
	}

	// Keeping track of attachments is a little difficult, here we create a map to hold all attachments sorted by the folder they were found in.
	// At the end, we go through all entries found and lookup the folder corresponding to that entry's path and add attachments if any were found.
	attachmentsMap := map[string][]Attachment{}

	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case info.IsDir():
			return nil
		case strings.Contains(subpath, gitDir):
			return filepath.SkipDir
		case !strings.Contains(info.Name(), "entry.md"):
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

		entry, entryErr := parser.FromFile(subpath)

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

// FromDirectoryAsync returns an Collection built from a directory.
// It will return an Collection, a list of errors that occured while parsing entries and finally an error that occured
// when processing the directory or adding an entry.
// The main difference from FromDirectory is that this is done asynchronously by spawning threads to parse many entries simulataneously.
// This can be anywhere from 1.5-3.0x faster but means it is a more intensive task.
func FromDirectoryAsync(path, dateLayout, tagPrefix string) (graph *Collection, entryErrs []error, err error) {
	graph = NewCollection()

	// We calculate a couple things up here once to avoid calculating it for every file.
	start := strings.Index(path, "entries")
	gitDir := filepath.Join("entries", ".git")

	// Create a parser that can be reused for each entry.
	parser, err := NewParser(dateLayout, tagPrefix)
	if err != nil {
		return nil, nil, err
	}

	// Keeping track of attachments is a little difficult, here we create a map to hold all attachments sorted by the folder they were found in.
	// At the end, we go through all entries found and lookup the folder corresponding to that entry's path and add attachments if any were found.
	attachmentsMap := map[string][]Attachment{}

	// numWorkers is decremented every time a worker finishes
	numWorkers := 3
	// subpaths is a channel of paths to entries
	subpaths := make(chan string)
	// results is a channel of the results of a worker processing an entry which contains a .entry and a .err
	results := make(chan entryMsg)

	// Spawn the amount of workers specified:
	for w := 0; w < numWorkers; w++ {
		go fileWorker(w, parser, subpaths, results)
	}

	// In one goroutine, recursively walk the path to entries that is given and add entries that need to be passed
	// to the subpaths channel.
	// Setting an error like this seems like a bad idea since it might lead to a race condition -- though I'm pretty
	// sure that since filepath.Walk is the only thing that can change it and since we check the error after the work is done
	// it shouldn't be an issue.
	var walkErr error
	go func() {
		walkErr = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
			switch {
			case err != nil:
				return err
			case info.IsDir():
				return nil
			case strings.Contains(subpath, gitDir):
				return filepath.SkipDir
			case !strings.Contains(info.Name(), "entry.md"):
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

			// We don't need to worry about errors here, since they are sent via the results channel.
			subpaths <- subpath
			return nil
		})

		// Once we're done, close the channel.
		close(subpaths)
	}()

	for msg := range results {
		// If an worker is finished, it will send an entryMsg with two nil values, i.e. entryMsg{nil, nil}
		// This isn't obvious that it means it's finished, but it saves adding another field to the entryMsg
		// which could waste memory.
		if msg.entry == nil && msg.err == nil {
			numWorkers--

			// Either only one worker is finished or all the work is done.
			if numWorkers == 0 {
				break
			} else {
				continue
			}
		}

		if msg.err != nil {
			entryErrs = append(entryErrs, msg.err)
			continue
		}

		err = graph.Add(msg.entry)
		if err != nil {
			return nil, []error{}, err
		}
	}

	if walkErr != nil {
		return nil, []error{}, walkErr
	}

	// This shouldn't cause any issues seeing as this code should only run after all the parsing has finished.
	// Here we set all the attachments that have been found.
	for _, entry := range graph.pathMap {
		entry.Attachments = attachmentsMap[entry.Path]
	}

	return graph, entryErrs, nil
}

// entryMsg is how fileWorkers communicate with the main FromDirectory function.
// If both fields are set to nil, i.e. entryMsg{nil, nil}, it means a worker has finished.
type entryMsg struct {
	entry *Entry
	err   error
}

// fileWorker completes parsing all the entries asynchronously.
func fileWorker(id int, parser Parser, subpaths <-chan string, results chan<- entryMsg) {
	for subpath := range subpaths {
		entry, entryErr := parser.FromFile(subpath)
		results <- entryMsg{entry, entryErr}
	}

	results <- entryMsg{nil, nil}
}
