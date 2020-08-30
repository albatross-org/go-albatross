package entries

import (
	"fmt"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// Collection represents a searchable collection of entries.
// It can be used to resolve links.
type Collection struct {
	titleMap map[string][]*Entry // entries can share titles
	pathMap  map[string]*Entry   // paths are unique
}

// NewCollection returns a new, initialised Collection.
func NewCollection() *Collection {
	return &Collection{
		titleMap: make(map[string][]*Entry),
		pathMap:  make(map[string]*Entry),
	}
}

// Len returns the amount of entries in the collection.
func (collection *Collection) Len() int {
	return len(collection.pathMap)
}

// In returns true if the entry exists in the collection. It returns false otherwise.
func (collection *Collection) In(entry *Entry) bool {
	return collection.pathMap[entry.Path] != nil
}

// FindLinksTo returns a list of links present in the collection which link to the entry specified..
func (collection *Collection) FindLinksTo(entry *Entry) []Link {
	links := []Link{}

	for _, existingEntry := range collection.pathMap {
		for _, link := range existingEntry.OutboundLinks {
			if link.Path == entry.Path {
				links = append(links, link)
			} else if link.Title == entry.Title {
				links = append(links, link)
			}
		}
	}

	return links
}

// ResolveLink takes a link and returns the entry that this link points to.
// TODO: come up with a better way of handling links which match multiple entries (because they share titles). At the moment it returns the first match.
// If it can't find the matching entry, it will return nil.
func (collection *Collection) ResolveLink(link Link) *Entry {
	switch link.Type {
	case LinkPathNoName, LinkPathWithName:
		return collection.pathMap[link.Path]
	case LinkTitleNoName, LinkTitleWithName:
		matching := collection.titleMap[link.Title]
		if len(matching) == 0 {
			return nil
		}

		return matching[0]
	}

	panic(fmt.Errorf("unknown link type '%d'", link.Type))
}

// Add adds an entry to the entry collection.
// If it already exists, it will return an ErrEntryAlreadyExists.
func (collection *Collection) Add(entry *Entry) error {
	if collection.pathMap[entry.Path] != nil {
		return ErrEntryAlreadyExists{Path: entry.Path, Title: entry.Title}
	}

	if len(collection.titleMap[entry.Title]) != 0 {
		for _, existingEntry := range collection.titleMap[entry.Title] {
			if existingEntry.Path == entry.Path {
				return ErrEntryAlreadyExists{Path: entry.Path, Title: entry.Title}
			}
		}
	}

	collection.pathMap[entry.Path] = entry

	if collection.titleMap[entry.Title] == nil {
		collection.titleMap[entry.Title] = []*Entry{}
	}
	collection.titleMap[entry.Title] = append(collection.titleMap[entry.Title], entry)

	return nil
}

// AddMany adds multiple Entries to the collection at once. It is a shorthand for manually calling .Add on every argument.
func (collection *Collection) AddMany(entries ...*Entry) error {
	for _, entry := range entries {
		err := collection.Add(entry)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes an entry from the entry collection.
// If doesn't exist, it will return an ErrEntryDoesntExist.
func (collection *Collection) Delete(entry *Entry) error {
	if collection.pathMap[entry.Path] == nil {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	if len(collection.titleMap[entry.Title]) == 0 {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	titleMapIndex := -1

	for i, existingEntry := range collection.titleMap[entry.Title] {
		if existingEntry.Path == entry.Path {
			titleMapIndex = i
			break
		}
	}

	if titleMapIndex == -1 {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	collection.titleMap[entry.Title] = removeEntry(collection.titleMap[entry.Title], titleMapIndex)
	delete(collection.pathMap, entry.Path)

	return nil
}

// copy returns a copy of the collection.
func (collection *Collection) copy() *Collection {
	newGraph := NewCollection()

	for k, v := range collection.pathMap {
		newGraph.pathMap[k] = v
	}

	for title, existingEntries := range collection.titleMap {
		entries := append([]*Entry{}, existingEntries...)
		newGraph.titleMap[title] = entries
	}

	return newGraph
}

// Filter runs the filters specified on the entries collection. It returns a copy of the entries collection.
func (collection *Collection) Filter(filters ...Filter) (*Collection, error) {
	curr := collection.copy()
	filter := FilterAnd(filters...)

	remove := []*Entry{}

	for _, entry := range collection.pathMap {
		if !filter(entry) {
			remove = append(remove, entry)
		}
	}

	for _, entry := range remove {
		err := curr.Delete(entry)
		if err != nil {
			return nil, err
		}
	}

	return curr, nil
}

// Graph gets the graphviz representation of the Collection, mainly for debugging and visualisation purposes.
func (collection *Collection) Graph() (*graphviz.Graphviz, *cgraph.Graph, error) {
	g := graphviz.New()

	viz, err := g.Graph(graphviz.Directed)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create graphviz collection: %w", err)
	}

	entryNodeMap := make(map[*Entry]*cgraph.Node)

	for path, entry := range collection.pathMap {
		n, err := viz.CreateNode(fmt.Sprintf("%s (%s)", path, entry.Title))
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't create graphviz node: %w", err)
		}

		entryNodeMap[entry] = n
	}

	for _, entry := range collection.pathMap {
		links := collection.FindLinksTo(entry)

		for _, link := range links {
			_, err := viz.CreateEdge(
				fmt.Sprintf("%s => %s", link.Parent.Path, entry.Path),
				entryNodeMap[link.Parent],
				entryNodeMap[entry],
			)

			if err != nil {
				return nil, nil, fmt.Errorf("couldn't create graphviz edge: %w", err)
			}
		}
	}

	return g, viz, nil
}

// List converts the collection into an List.
func (collection *Collection) List() List {
	l := []*Entry{}
	for _, entry := range collection.pathMap {
		l = append(l, entry)
	}

	return List{l}
}

// removeEntry is a helper function for removing an entry at a given index from a list of entries.
func removeEntry(es []*Entry, i int) []*Entry {
	es[i] = es[len(es)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return es[:len(es)-1]
}
