package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionPathCmd represents the 'tags' action.
var ActionPathCmd = &cobra.Command{
	Use:     "path",
	Aliases: []string{"paths"},
	Short:   "print paths",
	Long: `path will display the paths of all matched entries. For example:
	
	$ albatross get -p school/gcse path
	school/gcse/physics/topic7/electromagnetism
	school/gcse/physics/topic8/life-cycle-of-stars
	school/gcse/physics/topic4/nuclear-fission
	school/gcse/physics/topic8/solar-system-and-orbits
	...

Printing paths is the default when no subcommand is given to albatross get. You can always omit the 'path' part of the command:

	$ albatross get -p school/gcse
	school/gcse/physics/topic7/electromagnetism
	school/gcse/physics/topic8/life-cycle-of-stars
	school/gcse/physics/topic4/nuclear-fission
	school/gcse/physics/topic8/solar-system-and-orbits
	...

The functionalities of this command can also be achieved by the template command:

	$ albatross get -p school/gcse template {{.Path}}
	`,

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
