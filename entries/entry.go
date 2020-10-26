package entries

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Entry represents a parsed `entry.md` file.
type Entry struct {
	// Path to entry file.
	Path string `json:"path"`

	// Contents is the contents of the file without front matter.
	Contents string `json:"contents"`

	// OriginalContents is the contents of the file with the front matter.
	OriginalContents string `json:"originalContents"`

	// Tags are all the tags present in the document. For example, "@!journal".
	Tags []string `json:"tags"`

	// OutboundLinks are links going from this entry to another one.
	// These are known when the entry is parsed.
	OutboundLinks []Link `json:"outboundLinks"`

	// Date extracted from the entry.
	Date time.Time `json:"date"`

	// ModTime is the modification time for the entry.
	// Note: this is not always accurate, since encrypting and decryting all the files will "modify" them. Therefore it cannot be used for sorting
	// accurately.
	ModTime time.Time `json:"mod_time"`

	// Title of the entry.
	Title string `json:"title"`

	// Metadata is all the front-matter.
	Metadata map[string]interface{} `json:"metadata"`
}

// NewEntryFromFile returns a new Entry given a file system and a path to the `entry.md` file in that file system.
// It will return an error if the entry cannot be read.
func NewEntryFromFile(originalPath string) (*Entry, error) {
	path := strings.TrimSuffix(originalPath, "/entry.md")

	file, err := os.Open(originalPath)
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

	stat, err := file.Stat()
	if err != nil {
		return nil, ErrEntryReadFailed{Path: path, Err: fmt.Errorf("error getting file stat: %w", err)}
	}

	entry.ModTime = stat.ModTime()

	if entry.Date == (time.Time{}) {
		entry.Date = entry.ModTime
	}

	// Here we strip the path to the store itselft from the store.
	// This means something like:
	// "/home/user/.local/share/albatross/default/entries/journal/2020/04/10"
	// becomes
	// "journal/2020/04/10"
	// Which is the format used by the rest of the program.
	start := strings.Index(path, "entries")
	if start != -1 {
		path = path[start+8:]
	}
	entry.Path = path

	return entry, nil
}
