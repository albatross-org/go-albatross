package cmd

import (
	"fmt"
	"path/filepath"

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

You can also output the absolute paths using the --absolute option:

	$ albatross get -p school/gcse --absolute
	/home/david/documents/albatross/school/gcse/physics/topic7/electromagnetism
	/home/david/documents/albatross/school/gcse/physics/topic8/life-cycle-of-stars
	/home/david/documents/albatross/school/gcse/physics/topic4/nuclear-fission
	/home/david/documents/albatross/school/gcse/physics/topic8/solar-system-and-orbits
	...

Or the paths to the entry.md files instead using --to-entry-file:

	$ albatross get -p school/gcse --to-entry-file
	school/gcse/physics/topic7/electromagnetism/entry.md
	school/gcse/physics/topic8/life-cycle-of-stars/entry.md
	school/gcse/physics/topic4/nuclear-fission/entry.md
	school/gcse/physics/topic8/solar-system-and-orbits/entry.md
	...

Or a combination of both:

	$ albatross get -p school/gcse --absolute
	/home/david/documents/albatross/school/gcse/physics/topic7/electromagnetism/entry.md
	/home/david/documents/albatross/school/gcse/physics/topic8/life-cycle-of-stars/entry.md
	/home/david/documents/albatross/school/gcse/physics/topic4/nuclear-fission/entry.md
	/home/david/documents/albatross/school/gcse/physics/topic8/solar-system-and-orbits/entry.md
	...

The functionalities of this command can also be achieved by the template command:

	$ albatross get -p school/gcse template {{.Path}}
	`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		absolute, err := cmd.Flags().GetBool("absolute")
		checkArg(err)

		toEntryFile, err := cmd.Flags().GetBool("to-entry-file")
		checkArg(err)

		// It's done like this with a seperate for loop for each case instead of all
		// in loop because it might be marginally faster since it doesn't have to be checked
		// on every iteration probably neglibly so though.
		switch {
		case absolute && toEntryFile:
			for _, entry := range list.Slice() {
				fmt.Println(filepath.Join(store.Path, "entries", entry.Path, "entry.md"))
			}
		case !absolute && toEntryFile:
			for _, entry := range list.Slice() {
				fmt.Println(filepath.Join(entry.Path, "entry.md"))
			}
		case absolute && !toEntryFile:
			for _, entry := range list.Slice() {
				fmt.Println(filepath.Join(store.Path, "entries", entry.Path))
			}
		case !absolute && !toEntryFile:
			for _, entry := range list.Slice() {
				fmt.Println(entry.Path)
			}
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionPathCmd)

	ActionPathCmd.Flags().Bool("absolute", false, "show the absolute path to the entry in the store instead of the one relative to the root of the store")
	ActionPathCmd.Flags().Bool("to-entry-file", false, "show the path to the entry.md file rather than just the folder as a whole")
}
