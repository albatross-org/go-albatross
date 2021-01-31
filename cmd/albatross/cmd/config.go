package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ConfigCmd represents the "config" command.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the current configuration",
	Long: `This command will print the current configuration being used by the store.
	
For example:
`,
	Aliases: []string{"configuration"},
	Run: func(cmd *cobra.Command, args []string) {
		bs, err := json.MarshalIndent(store.Config, "", "\t")
		if err != nil {
			fmt.Println("Error converting the config struct to JSON:")
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(bs))
	},
}

func init() {
	rootCmd.AddCommand(ConfigCmd)
}
