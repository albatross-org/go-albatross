package entries

import (
	"path/filepath"
	"strings"
)

// Tree represents entries in a tree hierarchy.
type Tree struct {
	// Entry is the entry at the root of this tree.
	Entry *Entry

	// IsEntry is set to false if this node of the tree isn't actually an entry but a 'passthrough' for other entries.
	// For example:
	// |-> school 			    -> An entry. IsEntry = true.
	//     |-> further-maths    -> Not an entry but still contains children. IsEntry = false.
	//         |-> syllabus     -> An entry. IsEntry = true.
	IsEntry bool

	// Path is the path to this node in the tree, e.g. school/further-maths. It is used to keep track of organisation
	// even if there's not an underlying entry.
	Path string

	// Children are a list of all sub-entries.
	Children []*Tree

	// Level is an integer keeping tack of how deep we are in the tree, relative to the first root entry it was parsed from.
	Level int

	// Parent is the entry that this is a child of.
	Parent *Tree
}

func listToTree(rootPath string, rootEntry *Entry, list []*Entry, parent *Tree, level int) *Tree {
	children := []*Tree{}
	passthroughs := map[string]bool{}
	tree := &Tree{}

	for _, entry := range list {
		switch {
		case filepath.Dir(entry.Path) == rootPath || (rootPath == "" && !strings.Contains(entry.Path, "/")):
			// The entry is a direct child of the root entry, such as "school/" and "school/further-maths".
			children = append(
				children,
				listToTree(entry.Path, entry, list, tree, level+1),
			)
		case strings.HasPrefix(entry.Path, rootPath+"/") && !(filepath.Dir(entry.Path) == rootPath) && rootPath != "":
			// The entry is a child of the entry, but there's a passthrough between them. For example,
			// "school/" and "school/further-maths/complex-numbers/test". Here "school/further-maths" is the passthrough.
			// We need to get the path of the passthrough without the other entry on the end.
			// school/further-maths/complex-numbers/test
			//                      ^^^^^^^^^^^^^^^^^^^^ Remove me

			// Find the first "/" without the first path (e.g. the first "/" in "further-maths/complex-numbers/test")
			firstIndex := strings.Index(strings.TrimPrefix(entry.Path, rootPath+"/"), "/")

			// Trim the end we don't want and add it to the list of paths. We use a map since there might be multiple
			// sub-entries that go through the same profile. For example:
			//
			//  school/further-maths/complex-numbers
			//  school/further-maths/argand-diagrams
			//
			// Have the same passthrough and we don't want to add them to a list twice.
			passthrough := entry.Path[:len(rootPath+"/")+firstIndex]
			passthroughs[passthrough] = true
		case rootPath == "":
			// It's a passthrough at the root of the store.
			passthrough := entry.Path[:strings.Index(entry.Path, "/")]
			passthroughs[passthrough] = true
		}
	}

	for passthrough := range passthroughs {
		children = append(
			children,
			listToTree(passthrough, nil, list, tree, level+1),
		)
	}

	// For direct children:
	// filepath.Dir(entry.Path) == root.Path

	// For passthrough entries:
	// strings.HasPrefix(entry.Path, root.Path + "/") && !(filepath.Dir(entry.Path) == root.Path)

	tree.Entry = rootEntry
	tree.IsEntry = rootEntry != nil
	tree.Path = rootPath
	tree.Children = children
	tree.Level = level
	tree.Parent = parent

	return tree
}

// ListAsTree converts an entries.List type into an entries.Tree type.
func ListAsTree(list List) *Tree {
	return listToTree("", nil, list.Slice(), nil, 0)
}
