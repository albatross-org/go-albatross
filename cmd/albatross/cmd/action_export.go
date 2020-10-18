package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// ActionExportCmd represents the 'tags' action.
var ActionExportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"json"},
	Short:   "export entries",
	Long: `export will export entries in different formats, like JSON or YAML.
	
You can also export entries to an EPUB file but as this has additional options it is a subcommand:

	$ albatross get -p school --sort 'date' export epub`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		format, err := cmd.Flags().GetString("format")
		checkArg(err)

		var out []byte

		switch format {
		case "json":
			out, err = json.Marshal(list.Slice())
		case "yaml":
			out, err = yaml.Marshal(list.Slice())
		case "epub":
			fmt.Println("The correct command is: albatross get export epub")
			os.Exit(1)
		default:
			fmt.Println("Invalid output format:", format)
			fmt.Println("Currently supported are: json, yaml")
			os.Exit(1)
		}

		if err != nil {
			fmt.Println("error marshalling entries:")
			fmt.Println(err)
			os.Exit(1)
		}

		// _ = out
		fmt.Println(string(out))
	},
}

func init() {
	GetCmd.AddCommand(ActionExportCmd)

	ActionExportCmd.Flags().String("format", "json", "format to export entries in (json, yaml)")
}
