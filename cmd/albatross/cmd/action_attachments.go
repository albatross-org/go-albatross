package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ActionAttachmentsCmd represents the 'attachments' action.
var ActionAttachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "print attachments",
	Long: `attachments will print any attachments to entries.

	$ albatross get -p school/a-level/further-maths attachments
	/home/david/documents/.../argand-diagram.png
	
You can specify what kind of path to print with these flags:

	--name	    argand-diagram.png
	--rel-path  school/further-maths/argand-diagrams/argand-diagram.png
	--abs-path  /home/david/documents/.../argand-diagram.png`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		name, err := cmd.Flags().GetBool("name")
		checkArg(err)

		relPath, err := cmd.Flags().GetBool("rel-path")
		checkArg(err)

		absPath, err := cmd.Flags().GetBool("abs-path")
		checkArg(err)

		// Since absPath is the default, if either of the other two are set we assume
		// absPath to be false.
		if absPath && (name || relPath) {
			absPath = false
		}

		// Only allow one to be set explicitely.
		if name && relPath {
			fmt.Println("Please specify one flag:")
			fmt.Println("--name, --rel-path or --abs-path")
			os.Exit(1)
		}

		for _, entry := range list.Slice() {
			for _, attachment := range entry.Attachments {
				switch {
				case name:
					fmt.Println(attachment.Name)
				case relPath:
					fmt.Println(attachment.RelPath)
				case absPath:
					fmt.Println(attachment.AbsPath)
				}
			}
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionAttachmentsCmd)

	ActionAttachmentsCmd.Flags().Bool("name", false, "print only the filename (e.g. argand-diagram.png)")
	ActionAttachmentsCmd.Flags().Bool("rel-path", false, "print the path relative to the store (e.g. school/a-level/further-maths/argand-diagrams/argand-diagram.png)")
	ActionAttachmentsCmd.Flags().Bool("abs-path", true, "print the absolute path (e.g. /home/dave/documents/.../argand-diagram.png)")
}
