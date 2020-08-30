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
	Long:    `export will export entries in different formats, like JSON or TOML`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		metadata, err := cmd.Flags().GetBool("metadata")
		checkArg(err)

		format, err := cmd.Flags().GetString("format")
		checkArg(err)

		var out []byte
		var toMarshal interface{}

		if metadata {
			metadatas := []map[string]interface{}{}
			for _, entry := range list.Slice() {
				metadatas = append(metadatas, entry.Metadata)
			}

			toMarshal = metadatas
		} else {
			toMarshal = list.Slice()
		}

		switch format {
		case "json":
			out, err = json.Marshal(toMarshal)
		case "yaml":
			out, err = yaml.Marshal(toMarshal)
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

		fmt.Println(string(out))
	},
}

func init() {
	GetCmd.AddCommand(ActionExportCmd)

	ActionExportCmd.Flags().Bool("metadata", false, "only export metadata")
	ActionExportCmd.Flags().String("format", "json", "format to export entries in (json, yaml)")
}
