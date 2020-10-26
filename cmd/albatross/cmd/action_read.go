package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ContentsCmd represents the contents command.
var ContentsCmd = &cobra.Command{
	Use:     "contents",
	Aliases: []string{"print", "contents"},
	Short:   "print entry contents",
	Long: `contents (also 'print' or 'read') will print the contents of entries to standard output.
	
	$ albatross get -p school/gcse/physics/topic4/nuclear-fusion contents | less
	# Opens a specific entry in less.
	
If you're using this to read entries, you'd probably be better off just simply using the update command
and not editing it whilst you're reading it. That way, you can get syntax highlighting.

This is more useful for doing stuff like word counts:

	$ albatross get contents | wc -w
	513362
	
	$ albatross get contents | tr -c '[:alnum:]' '[\n*]' | sort | uniq -c | sort -nr | head -100
	# Get the most common 100 words in the search`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		raw, err := cmd.Flags().GetBool("raw")
		checkArg(err)

		between, err := cmd.Flags().GetString("between")
		checkArg(err)

		between += "\n"

		for _, entry := range list.Slice() {
			if raw {
				fmt.Println(entry.OriginalContents)
			} else {
				fmt.Println(entry.Contents)
			}

			fmt.Print(between)
		}
	},
}

func init() {
	GetCmd.AddCommand(ContentsCmd)

	ContentsCmd.Flags().Bool("raw", false, "include front matter when printing")
	ContentsCmd.Flags().String("between", "", "what to print between entries")
}
