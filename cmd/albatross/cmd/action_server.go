package cmd

import (
	"github.com/albatross-org/go-albatross/server"
	"github.com/spf13/cobra"
)

// responseMatchedMultiple is the response sent when a request matches multiple or zero entries.
type responseMatchedMultiple struct {
	Message string   `json:"message"`
	Entries []string `json:"entries"`
}

// ActionServerCmd represents the 'server' action.
var ActionServerCmd = &cobra.Command{
	Use:   "server",
	Short: "start HTTP server which serves JSON entries",
	Long:  `date will print the date of all matched entries`,

	Run: func(cmd *cobra.Command, args []string) {
		_, collection, _ := getFromCommand(cmd)

		port, err := cmd.Flags().GetInt("port")
		checkArg(err)

		s := server.NewServer(collection)
		s.Serve(port)
	},
}

func init() {
	GetCmd.AddCommand(ActionServerCmd)
	ActionServerCmd.Flags().Int("port", 2718, "port to run server")
}
