package cmd

import (
	"github.com/spf13/cobra"
)

// toolCmd represents the tool command
var toolCmd = &cobra.Command{
	Use:   "tool [subcommand]",
	Short: "tool runs other tools on an Albatross store",
	Long: `tool runs other tools on an Albatross store. These include:
	
- ankify: Generate anki flashcards`,
}

func init() {
	rootCmd.AddCommand(toolCmd)
}
