package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ReadCmd represents the read command.
var ReadCmd = &cobra.Command{
	Use:     "read",
	Aliases: []string{"print", "contents"},
	Short:   "print entries",
	Long:    `read will print entries to stdout`,

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
	GetCmd.AddCommand(ReadCmd)

	ReadCmd.Flags().Bool("raw", false, "include front matter when printing")
	ReadCmd.Flags().String("between", "", "what to print between entries")
}
