package entries

import (
	"fmt"

	"github.com/goccy/go-graphviz/cgraph"

	"github.com/goccy/go-graphviz"
)

// EntryGraph represents a graph of entries.
type EntryGraph struct {
	titleMap map[string][]*Entry // entries can share titles
	pathMap  map[string]*Entry   // paths are unique
}

// NewEntryGraph returns a new, initialised EntryGraph.
func NewEntryGraph() *EntryGraph {
	return &EntryGraph{
		titleMap: make(map[string][]*Entry),
		pathMap:  make(map[string]*Entry),
	}
}

// In returns true if the entry exists in the graph. It returns false otherwise.
func (graph *EntryGraph) In(entry *Entry) bool {
	return graph.pathMap[entry.Path] != nil
}

// Len returns the amount of entries in the graph.
func (graph *EntryGraph) Len() int {
	return len(graph.pathMap)
}

// Add adds an entry to the entry graph.
// If it already exists, it will return an ErrEntryAlreadyExists.
// It will update all the inbound links for entries present in the graph if the entry being added has outbound links to that entry.
// Inbound links can be removed with RemoveInboundLinks
func (graph *EntryGraph) Add(entry *Entry) error {
	if graph.pathMap[entry.Path] != nil {
		return ErrEntryAlreadyExists{Path: entry.Path, Title: entry.Title}
	}

	if len(graph.titleMap[entry.Title]) != 0 {
		for _, existingEntry := range graph.titleMap[entry.Title] {
			if existingEntry.Path == entry.Path {
				return ErrEntryAlreadyExists{Path: entry.Path, Title: entry.Title}
			}
		}
	}

	// Add inbound links by iterating through all existing entries and checking if they link to this one.
	for _, existingEntry := range graph.pathMap {
		for _, link := range existingEntry.OutboundLinks {
			if link.Path == entry.Path {
				entry.InboundLinks = append(entry.InboundLinks, existingEntry)
			}
		}
	}

	graph.pathMap[entry.Path] = entry

	if graph.titleMap[entry.Title] == nil {
		graph.titleMap[entry.Title] = []*Entry{}
	}
	graph.titleMap[entry.Title] = append(graph.titleMap[entry.Title], entry)

	return nil
}

// AddMany adds multiple Entries to the graph at once. It is a shorthand for manually calling .Add on every argument.
func (graph *EntryGraph) AddMany(entries ...*Entry) error {
	for _, entry := range entries {
		err := graph.Add(entry)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes an entry from the entry graph.
// If doesn't exist, it will return an ErrEntryDoesntExist.
// Delete will preserve inbound links, to remove them use RemoveInboundLinks
func (graph *EntryGraph) Delete(entry *Entry) error {
	if graph.pathMap[entry.Path] == nil {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	if len(graph.titleMap[entry.Title]) == 0 {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	titleMapIndex := -1

	for i, existingEntry := range graph.titleMap[entry.Title] {
		if existingEntry.Path == entry.Path {
			titleMapIndex = i
			break
		}
	}

	if titleMapIndex == -1 {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	graph.titleMap[entry.Title] = removeEntry(graph.titleMap[entry.Title], titleMapIndex)
	delete(graph.pathMap, entry.Path)

	return nil
}

// copy returns a copy of the graph.
func (graph *EntryGraph) copy() *EntryGraph {
	newGraph := NewEntryGraph()

	for k, v := range graph.pathMap {
		newGraph.pathMap[k] = v
	}

	for title, existingEntries := range graph.titleMap {
		entries := append([]*Entry{}, existingEntries...)
		newGraph.titleMap[title] = entries
	}

	return newGraph
}

// Filter runs the filters specified on the entries graph. It returns a copy of the entries graph.
func (graph *EntryGraph) Filter(filters ...Filter) (*EntryGraph, error) {
	curr := graph

	for _, filter := range filters {
		curr = curr.copy()
		err := filter(curr)
		if err != nil {
			return nil, err
		}
	}

	return curr, nil
}

// RemoveInboundLinks takes an entry and removes it from all InboundLinks in the graph.
// For example, if "food/pizza" links to "moods/hunger", "moods/hunger" will have a inbound link for "food/pizza".
// Running graph.RemoveInboundLinks(pizzaEntry) will get rid of this.
func (graph *EntryGraph) RemoveInboundLinks(entry *Entry) {
	for _, link := range entry.OutboundLinks {
		linkedEntry := graph.pathMap[link.Path]
		if linkedEntry == nil {
			continue // skip linked entries which don't exist
		}

		entryIndex := -1
		for i, inboundLink := range linkedEntry.InboundLinks {
			if inboundLink != nil && inboundLink.Path == entry.Path {
				entryIndex = i
				break
			}
		}

		if entryIndex != -1 {
			linkedEntry.InboundLinks = removeEntry(linkedEntry.InboundLinks, entryIndex)
		}
	}
}

// Graph gets the graphviz representation of the EntryGraph, mainly for debugging and visualisation purposes.
// TODO: this really isn't optimal.
func (graph *EntryGraph) Graph() (*graphviz.Graphviz, *cgraph.Graph, error) {
	g := graphviz.New()

	viz, err := g.Graph(graphviz.StrictUnDirected)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create graphviz graph: %w", err)
	}

	entryNodeMap := make(map[*Entry]*cgraph.Node)

	for path, entry := range graph.pathMap {
		n, err := viz.CreateNode(path)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't create graphviz node: %w", err)
		}

		entryNodeMap[entry] = n
	}

	for entry, node := range entryNodeMap {
		for _, link := range entry.InboundLinks {
			_, err := viz.CreateEdge(fmt.Sprintf("%s => %s", entry.Path, link.Path), node, entryNodeMap[link])
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't create graphviz edge: %w", err)
			}
		}
	}

	return g, viz, nil
}

// removeEntry is a helper function for removing an entry at a given index from a list of entries.
func removeEntry(es []*Entry, i int) []*Entry {
	es[i] = es[len(es)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return es[:len(es)-1]
}
