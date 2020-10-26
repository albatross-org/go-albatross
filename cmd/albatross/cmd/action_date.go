package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ActionDateCmd represents the 'date' action.
var ActionDateCmd = &cobra.Command{
	Use:   "date",
	Short: "print entry dates",
	Long: `date will print the date of all matched entries. If you want more complex formatting, it may be a good idea to
use the template command instead.
	
	$ albatross get date
	2020-03-16 17:28
	2020-05-31 15:42
	2020-08-31 07:27
	2020-07-27 10:32

Is equivalent too...

	$ albatross get template '{{.Date | date "2006-01-02 15:01"}}'
	2020-03-16 17:28
	2020-05-31 15:42
	2020-08-31 07:27
	2020-07-27 10:32

You can specify a date format using the --print-date-format flag. This is not to be confused with --date-format,
which specifies how "albatross get" should parse the --from and --until flags.`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)
		dateFormat, err := cmd.Flags().GetString("print-date-format")
		checkArg(err)

		for _, entry := range list.Slice() {
			fmt.Println(entry.Date.Format(dateFormat))
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionDateCmd)
	ActionDateCmd.Flags().String("print-date-format", "2006-01-02 15:04", "date format (go syntax) for dates")
}
