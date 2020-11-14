package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
				confirmation = confirmDelete(fmt.Sprintf("Are you sure you want to delete %s?", entry.Path))
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

		if len(list.Slice()) == 0 {
			fmt.Println("Nothing matched, nothing to delete.")
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionDeleteCmd)

	ActionDeleteCmd.Flags().Bool("force-delete", false, "do not prompt for deletion")
}

// confirmDelete displays a prompt `s` to the user and returns a bool indicating yes / no
// If the lowercased, trimmed input begins with anything other than 'y', it returns false
// Courtesy https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func confirmDelete(s string) bool {
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Empty input (i.e. "\n")
		if len(res) < 2 {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(res))[0] {
		case 'y':
			return true
		case 'n':
			return false
		default:
			fmt.Println("Please enter [y/n].")
		}
	}
}
