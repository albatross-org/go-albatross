package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ActionExportCmd represents the 'tags' action.
var ActionExportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"json"},
	Short:   "export entries into different formats",
	Long: `export will export entries in different formats, like JSON or as an EPUB file.

By default, the command will output a JSON serialised array of all the entries that were matched in the search.

	$ albatross get --sort 'date' export
	# Export all entries chronologically in JSON.
	
For help with EPUB export, see

	$ albatross get export epub --help
`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		format, err := cmd.Flags().GetString("format")
		checkArg(err)

		var out []byte

		switch format {
		case "json":
			out, err = json.Marshal(list.Slice())
		case "epub":
			fmt.Println("The correct command is: albatross get export epub")
			os.Exit(1)
		default:
			fmt.Println("Invalid output format:", format)
			fmt.Println("Currently supported are: json, epub")
			os.Exit(1)
		}

		if err != nil {
			fmt.Println("error marshalling entries:")
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(out))
	},
}

func init() {
	GetCmd.AddCommand(ActionExportCmd)

	ActionExportCmd.Flags().String("format", "json", "format to export entries in (currently only JSON is supported)")
}
