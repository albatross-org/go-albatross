package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionTitleCmd represents the 'tags' action.
var ActionTitleCmd = &cobra.Command{
	Use:   "title",
	Short: "print titles",
	Long: `title will display the titles of all matched entries.

	$ albatross get -p school/a-level/further-maths title
	Further Maths - Lessons
	Further Maths - Solving Systems of Equations Using Triangle Method
	Further Maths - Cubics
	Further Maths - Integration
	Further Maths - Sums of Squares
	
The functionalities of this command can be achieved with the template command:

	$ albatross get -p school/a-level/further-maths template "{{.Title}}"`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		for _, entry := range list.Slice() {
			fmt.Println(entry.Title)
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionTitleCmd)
}
