package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionTitleCmd represents the 'tags' action.
var ActionTitleCmd = &cobra.Command{
	Use:   "title",
	Short: "print titles",
	Long:  `title will display the titles of all matched entries`,

	Run: func(cmd *cobra.Command, args []string) {
		_, list := getFromCommand(cmd)

		for _, entry := range list.Slice() {
			fmt.Println(entry.Title)
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionTitleCmd)
}
