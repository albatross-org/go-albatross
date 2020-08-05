package entries

import "text/scanner"

// Link represents an outbound link from one Entry to another.
type Link struct {
	Path  string // The path of the entry being linked to.
	Title string // The title of the entry being linked to.

	StartPos scanner.Position
	EndPos   scanner.Position
}
