package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ActionTagsCmd represents the 'tags' action.
// TODO: option to lump all the tags together
var ActionTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "print tags",
	Long: `tags will display the tags in entries.
	
	$ albatross get -p school/a-level/further-maths tags
	@?further-maths, @?latex-block-alt
	@?further-maths, @?latex-block-alt
	...

The functionalities of this command can be achieved with the template command:

	$ albatross get -p school/a-level/further-maths template '{{.Tags | join ", "}}'
	@?further-maths, @?latex-block-alt
	@?further-maths, @?latex-block-alt
	...
`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		for _, entry := range list.Slice() {
			fmt.Println(strings.Join(entry.Tags, ", "))
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionTagsCmd)
}
