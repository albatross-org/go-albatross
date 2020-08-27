package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionLinksCmd represents the 'tags' action.
// TODO: option lump all the links together, like a set
// TODO: show links and thier locations
// TODO: show links without an actual location
var ActionLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "print links",
	Long:  `links will display all the links inside an entry`,

	Run: func(cmd *cobra.Command, args []string) {
		collection, list := getFromCommand(cmd)

		for _, entry := range list.Slice() {
			for _, link := range entry.OutboundLinks {
				linkedEntry := collection.ResolveLink(link)
				if linkedEntry != nil {
					fmt.Println(linkedEntry.Path)
				}
			}
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionLinksCmd)
}
