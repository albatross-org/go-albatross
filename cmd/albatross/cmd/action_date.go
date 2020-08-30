package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionDateCmd represents the 'tags' action.
var ActionDateCmd = &cobra.Command{
	Use:   "date",
	Short: "print date",
	Long:  `date will print the date of all matched entries`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)
		dateFormat, err := cmd.Flags().GetString("date-format")
		checkArg(err)

		for _, entry := range list.Slice() {
			fmt.Println(entry.Date.Format(dateFormat))
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionDateCmd)
	ActionDateCmd.Flags().String("date-format", "2006-01-02 15:04", "date format (go syntax) for dates")
}
