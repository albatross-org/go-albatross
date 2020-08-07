package entries

// LinkType represents a type of link. This could be:
// - A link by title (LinkTitleNoName), e.g. "[[Pizza]]"
// - A link by title with a name (LinkTitleWithName), e.g. "[[Pizza](Alternate name)]"
// - A link by path (LinkPathNoName), e.g. "{{food/pizza}}"
// - A link by path with a name (LinkPathWithName), e.g. "{{food/pizza}(Altername name)"
type LinkType int

const (
	// LinkTitleNoName is a link by title, e.g. "[[Pizza]]"
	LinkTitleNoName LinkType = iota

	// LinkTitleWithName is a link by title with a name, e.g. "[[Pizza](Alternate name)]"
	LinkTitleWithName

	// LinkPathNoName is a link by path, e.g. "{{food/pizza}}"
	LinkPathNoName

	// LinkPathWithName is a link by path with a name, e.g. "{{food/pizza}(Altername name)"
	LinkPathWithName
)

// Link represents an outbound link from one Entry to another.
type Link struct {
	// Parent is the entry that is being linked from.
	Parent *Entry

	// Path is the path to the entry being linked to. This is blank if it's a title link.
	Path string

	// Title is the title of the entry being linked to. This is blank if it's a path link.
	Title string

	// Name is the name of the link. If no other name was specified, this is blank.
	Name string

	// Type is the type of link.
	Type LinkType

	// Loc is the location of the link in the entry text, represented by a two-element slice of the start and and positions.
	// The link text itself is at strippedContents[Loc[0]:Loc[1]]
	Loc []int
}
