package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/icza/dyno"

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

You can use the 'metadata' flag to only export metadata rather than the whole entry including it's contents:

	$ albatross get --sort 'date' export --metadata
	{
		"title": "Example Entry",
		"date": "2020-11-12 22:07",
		"some_piece_of_data": 10
	}
	
For help with EPUB export, see

	$ albatross get export epub --help
`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		format, err := cmd.Flags().GetString("format")
		checkArg(err)

		metadataOnly, err := cmd.Flags().GetBool("metadata")
		checkArg(err)

		var out []byte

		switch format {
		case "json":
			slice := list.Slice()
			metadatas := []map[string]interface{}{}

			if metadataOnly {
				for _, entry := range slice {
					metadatas = append(metadatas, entry.Metadata)
				}

				out, err = json.Marshal(metadatas)
			} else {
				out, err = json.Marshal(slice)
			}

			// Sometimes if the metadata contains numerical keys or nested dictionaries, we will get this error:
			// json: unsupported type: map[interface {}]interface {}
			// This is due to the implementation of the YAML parser and the JSON parser. More info can be found here:
			// https://stackoverflow.com/a/40737676
			// Therefore we convert the map[interface {}]interface {} to a map[string]interface{}, which the json Marshaller
			// understands. Understandably, this does add a slight performance overhead.
			if _, ok := err.(*json.UnsupportedTypeError); ok {
				if metadataOnly {
					for i, metadata := range metadatas {
						metadatas[i], ok = dyno.ConvertMapI2MapS(metadata).(map[string]interface{})
					}
					out, err = json.Marshal(metadatas)

				} else {
					for _, entry := range slice {
						entry.Metadata, ok = dyno.ConvertMapI2MapS(entry.Metadata).(map[string]interface{})
						if !ok {
							fmt.Println("error converting entry metadata into a format that can be marshalled into JSON")
							os.Exit(1)
						}
					}

					out, err = json.Marshal(slice)
				}

			}

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
	ActionExportCmd.Flags().Bool("metadata", false, "only export metadata, not contents")
}
