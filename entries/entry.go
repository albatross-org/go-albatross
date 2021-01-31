package entries

import (
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

	// Attachments are a list of the attachments to the entry.
	Attachments []Attachment

	// Synthetic specifys whether this is a pretend entry, such as used in the --parse-content option of the CLI or the result of a template-generated
	// entry.
	Synthetic bool
}
