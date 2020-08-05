package entries

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/afero"
)

// Entry represents a parsed `entry.md` file.
type Entry struct {
	// Path to entry file.
	Path string

	// Contents is the contents of the file without front matter.
	Contents string

	// Tags are all the tags present in the document. For example, "@!journal".
	Tags []string

	// OutboundLinks are links going from this entry to another one.
	// These are known when the entry is parsed.
	OutboundLinks []Link

	// InboundLinks are links coming from a different entry to this one.
	// These are known when the entry is added to an EntryGraph.
	InboundLinks []*Entry

	// Date extracted from the entry.
	Date time.Time

	// Title of the entry.
	Title string

	// Metadata is all the front-matter.
	Metadata map[string]interface{}
}

// NewEntry returns a new Entry given a file system and a path to the `entry.md` file in that file system.
// It will return an error if the entry cannot be read.
func NewEntry(fs afero.Fs, path string) (*Entry, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, ErrEntryReadFailed{Path: path, Err: err}
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, ErrEntryReadFailed{Path: path, Err: err}
	}

	content := string(bytes)

	dateLayout := "2006-01-02 15:04" // TODO: get date format from config or something. Hard-coded for now.
	builtinTagPrefix := "@!"         // TODO: get tag prefixes from config or something.
	customTagPrefix := "@?"

	parser, err := NewParser(dateLayout, builtinTagPrefix, customTagPrefix)
	if err != nil {
		return nil, err
	}

	entry, err := parser.Parse(path, content)
	if err != nil {
		return nil, err
	}

	if entry.Date == (time.Time{}) {
		stat, err := file.Stat()
		if err != nil {
			return nil, ErrEntryReadFailed{Path: path, Err: fmt.Errorf("error getting file stat: %w", err)}
		}

		entry.Date = stat.ModTime()
	}

	return entry, nil
}
