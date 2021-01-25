package cmd

import (
	"fmt"
	"strings"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/disiqueira/gotree"

	"github.com/spf13/cobra"
)

// Tree characters, courtesty of https://github.com/DiSiqueira/GoTree/.
const (
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
	blankItem    = "─── "
)

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
	Short:   "list entries returned",
	Long: `list prints the entries matched in a search in a tree like format. For example:
	
	$ albatross get -p school/gcse ls
	.
	└── school
		└── gcse
			└── physics
			│   ├── topic7
			│   │   ├── electromagnetism
			│   │   ├── revision-questions
			│   │   ├── motor-effect
			│   │   ├── motors
			│   │   ├── transformers
			│   │   ├── generators
			│   │   ├── generator-effect
			│   │   ├── loudspeakers
			│   │   ├── microphones
			│   │   ├── magnets
			│   ├── topic8
			│   │   ├── red-shift-and-big-bang
			│   │   ├── solar-system-and-orbits
			│   │   ├── life-cycle-of-stars
			│   │   ├── revision-questions
			│   ├── topic4
			│       └── nuclear-fission
			│       └── nuclear-fusion
			└── results

Currently, only printing the paths for entries is supported. In a future version you should be able to show more information.
`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)
		tree := entries.ListAsTree(list)
		printableTree := buildPrintableTree(tree)

		fmt.Println(printableTree.Print())
	},
}

func buildPrintableTree(tree *entries.Tree) gotree.Tree {
	root := gotree.New(tree.Path)

	for _, subtree := range tree.Children {
		if len(subtree.Children) != 0 {
			root.AddTree(buildPrintableTree(subtree))
		} else {
			root.Add(subtree.Path)
		}
	}

	return root
}

func printTree(trees []*entries.Tree) {
	length := len(trees)

	for i, tree := range trees {
		fmt.Print(strings.Repeat(emptySpace, tree.Level))

		switch {
		case i == length-1:
			fmt.Print(lastItem)
		case len(tree.Children) > 0:
			fmt.Print(blankItem)
		default:
			fmt.Print(middleItem)
		}

		fmt.Print(tree.Path)
		fmt.Print("\n")

		printTree(tree.Children)
	}
}

func init() {
	GetCmd.AddCommand(ActionLsCmd)

	ActionLsCmd.Flags().Int("depth", -1, "max depth to print, -1 prints all levels")
	ActionLsCmd.Flags().Bool("display-title", false, "print the title for each entry rather than the path")
}
