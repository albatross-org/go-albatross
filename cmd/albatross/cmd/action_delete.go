package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ActionDeleteCmd represents the 'delete' action.
var ActionDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"rm"},
	Short:   "delete entries",
	Long: `delete will delete entries from the store
	
	$ albatross get -c "evil plan to take over the world" delete
	# Delete all entries that mention your evil plan.

By default, it will prompt you to confirm for every matched entry. This can be overriden with the 'force-delete' flag.

	$ albatross get -c "the money is located" delete --force-delete
	# Delete any entry mentioning where the money is located`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		forceDelete, err := cmd.Flags().GetBool("force-delete")
		checkArg(err)

		for _, entry := range list.Slice() {
			var confirmation bool

			if !forceDelete {
				confirmation = confirmPrompt(fmt.Sprintf("Are you sure you want to delete %s?", entry.Path))
			}

			if confirmation || forceDelete {
				fmt.Println("Deleting", entry.Path)
				err = store.Delete(entry.Path)
				if err != nil {
					fmt.Println("Error deleting", entry.Path)
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		if slice.Len() == 0 {
			fmt.Println("Nothing matched, nothing to delete.")
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionDeleteCmd)

	ActionDeleteCmd.Flags().Bool("force-delete", false, "do not prompt for deletion")
}
