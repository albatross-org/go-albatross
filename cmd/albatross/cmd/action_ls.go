package cmd

import (
	"fmt"
	"strings"

	"github.com/disiqueira/gotree"
	"github.com/spf13/cobra"
)

// mapToTree converts a map[string]interface{} into a gotree.Tree.
// maxDepth can be disabled by setting to -1.
// By default, path should be an empty string.
// renderFunc should take a path to an entry and should return how it should be displayed.
func mapToTree(rootKey string, stringTree map[string]interface{}, maxDepth int, path string, renderFunc func(string) string) gotree.Tree {
	tree := gotree.New(rootKey)

	if len(stringTree) == 0 {
		return tree
	} else if maxDepth == 0 {
		tree.Add("...")
		return tree
	}

	for key, subStringTree := range stringTree {
		subtree := mapToTree(key, subStringTree.(map[string]interface{}), maxDepth-1, path+"/"+key, renderFunc)

		if len(subtree.Items()) == 0 {
			tree.Add(renderFunc(strings.TrimLeft(path+"/"+key, "/")))
		} else {
			tree.AddTree(subtree)
		}
	}

	return tree
}

// stripLongestCommonPrefix will remove the longest prefix that all slice items share.
// If they do not share any prefix, it will do nothing.
func stipLongestCommonPrefix(xs []string) []string {
	return xs
}

// ActionLsCmd represents the 'ls' action.
// This is some pretty ugly code, TODO.
//
// Basically, it first converts a list of paths into a nested map[string]interface{}, like parsing a list of files into a tree.
// Then it uses the mapToTree command to recursivly convert the nested map structure into a GoTree structure. It uses the renderFunc
// to determine how it should display entries.
//
// BUG: When using the --display-title option, if an entry itself contains other entries, the path will be printed instead of the title.
// TODO: Currently no sorting.
var ActionLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"tree"},
	Short:   "list entries returned in a tree format",
	Long:    `list entries returned in a tree format`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		depth, err := cmd.Flags().GetInt("depth")
		checkArg(err)

		displayTitle, err := cmd.Flags().GetBool("display-title")
		checkArg(err)

		paths := [][]string{}
		stringTree := map[string]interface{}{}
		renderFunc := func(path string) string { return path[strings.LastIndex(path, "/")+1:] }

		if displayTitle {
			renderFunc = func(path string) string {
				for _, entry := range list.Slice() {
					if entry.Path == path {
						return entry.Title
					}
				}

				return "<not found>"
			}
		}

		for _, entry := range list.Slice() {
			paths = append(paths, strings.Split(entry.Path, "/"))
		}

		// Convert the entries into a nested map[*Entry]interface{}
		for _, path := range paths {
			subtree := stringTree

			for _, curr := range path {
				newSubtree, ok := subtree[curr].(map[string]interface{})

				if !ok {
					if subtree == nil {
						subtree = make(map[string]interface{})
					}

					subtree[curr] = map[string]interface{}{}
					subtree, _ = subtree[curr].(map[string]interface{})
				} else {
					subtree = newSubtree
				}
			}
		}

		tree := mapToTree(".", stringTree, depth, "", renderFunc)
		fmt.Println(tree.Print())
	},
}

func init() {
	GetCmd.AddCommand(ActionLsCmd)

	ActionLsCmd.Flags().Int("depth", -1, "max depth to print, -1 prints all levels")
	ActionLsCmd.Flags().Bool("display-title", false, "print the title for each entry rather than the path")
}
