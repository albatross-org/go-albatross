package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionPathCmd represents the 'tags' action.
var ActionPathCmd = &cobra.Command{
	Use:   "path",
	Short: "print paths",
	Long:  `path will display the paths of all matched entries`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		for _, entry := range list.Slice() {
			fmt.Println(entry.Path)
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionPathCmd)
}
