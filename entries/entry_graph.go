package entries

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
	return len(graph.titleMap)
}

// Add adds an entry to the entry graph.
// If it already exists, it will return an ErrEntryAlreadyExists.
func (graph *EntryGraph) Add(entry *Entry) error {
	if graph.pathMap[entry.Path] != nil {
		return ErrEntryAlreadyExists{Path: entry.Path, Title: entry.Title}
	}

	if len(graph.titleMap[entry.Title]) != 0 {
		for _, existingEntry := range graph.titleMap[entry.Title] {
			if existingEntry.Contents == entry.Contents {
				return ErrEntryAlreadyExists{Path: entry.Path, Title: entry.Title}
			}
		}
	}

	graph.pathMap[entry.Path] = entry
	graph.titleMap[entry.Path] = append(graph.titleMap[entry.Path], entry)

	return nil
}

// Delete removes an entry from the entry graph.
// If doesn't exist, it will return an ErrEntryDoesntExist.
func (graph *EntryGraph) Delete(entry *Entry) error {
	if graph.pathMap[entry.Path] == nil {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	if len(graph.titleMap[entry.Title]) == 0 {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	index := -1

	for i, existingEntry := range graph.titleMap[entry.Title] {
		if existingEntry.Contents == entry.Contents {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrEntryDoesntExist{Path: entry.Path, Title: entry.Title}
	}

	graph.titleMap[entry.Title] = append(graph.titleMap[entry.Title][:index], graph.titleMap[entry.Title][index+1:]...)
	delete(graph.pathMap, entry.Path)

	return nil
}
