package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionLinksCmd represents the 'tags' action.
var ActionLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "print links",
	Long: `links will display all the links inside an entry
	
This command is often used in conjunction with 'albatross get -i'. By default it will
display the paths to all linked entries:

    $ albatross get -p school/gcse/physics/topic7
	
	school/gcse/physics/topic7/motor-effect
    school/gcse/physics/topic7/electromagnetism
    school/gcse/physics/topic7/generator-effect
	school/gcse/physics/topic7/generators
	
You can then pipe this into 'albatross get -i' to see more info about those entries. For
example, to print the titles and dates of the linked entries:

	$ albatross get -p school/gcse/physics/topic7 | albatross get -i template '{{.Date | date "2006-01-02"}} - {{.Title}}'

	2020-08-09 The Motor Effect
	2020-08-09 Electromagnetism
	2020-08-09 The Generator Effect
	2020-08-09 Generators

Another common use case of this command could be a to-do list of things to write in the future by
checking to see links to entries that don't yet exist.

	$ albatross get links --dont-exist
	# Or, using shorthand flags:
	$ albatross get links -e
	
You can also display the entry that is linking to the other entry using the --outbound flag:

	$ albatross get links --outbound

And finally to print the link text (such as [[Link]] or {{path/to/link}}) instead of the path itself,
you can use the --text flag:

	$ albatross get links --text
`,

	Run: func(cmd *cobra.Command, args []string) {
		collection, list := getFromCommand(cmd)

		outbound, err := cmd.Flags().GetBool("outbound")
		checkArg(err)

		dontExistOnly, err := cmd.Flags().GetBool("dont-exist")
		checkArg(err)

		displayText, err := cmd.Flags().GetBool("text")
		checkArg(err)

		for _, entry := range list.Slice() {
			for _, link := range entry.OutboundLinks {
				linkedEntry := collection.ResolveLink(link)
				if linkedEntry != nil && !dontExistOnly {
					text := ""

					if displayText {
						text = entry.Contents[link.Loc[0]:link.Loc[1]]
					} else {
						text = linkedEntry.Path
					}

					if outbound {
						fmt.Printf("%s -> %s\n", entry.Path, text)
					} else {
						fmt.Println(text)
					}

				} else if linkedEntry == nil && dontExistOnly {
					text := entry.Contents[link.Loc[0]:link.Loc[1]]

					if outbound {
						fmt.Printf("%s -> %s\n", entry.Path, text)
					} else {
						fmt.Println(text)
					}
				}
			}
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionLinksCmd)

	ActionLinksCmd.Flags().BoolP("outbound", "o", false, "also show the outbound linker (i.e. the entry that's linking from) in the output")
	ActionLinksCmd.Flags().BoolP("dont-exist", "e", false, "only show links to entries which don't exist")
	ActionLinksCmd.Flags().Bool("text", false, "show the link text instead of the path, such as [[Link]] or {{path/to/linked}}")
}
