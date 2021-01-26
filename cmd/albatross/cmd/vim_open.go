package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/albatross-org/go-albatross/entries"

	"github.com/spf13/cobra"
)

// VimOpenCmd represents the vim open command
var VimOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Create and returns a temporary file for Vim to edit.",
	Long:  `This command creates and returns a temporary file for Vim to edit.`,
	Run: func(cmd *cobra.Command, args []string) {
		encrypted, err := store.Encrypted()
		if err != nil {
			log.Fatal(err)
		} else if encrypted {
			fmt.Println("ERROR: It's not currently possible to use the Vim integration while the store is encrypted. This may change in a future version.")
			fmt.Println("\nTo allow access, run:")
			fmt.Println("    $ albatross decrypt")
			os.Exit(1)
		}

		path, err := cmd.Flags().GetString("path")
		checkArgVerbose(cmd, "path", err)

		title, err := cmd.Flags().GetString("title")
		checkArgVerbose(cmd, "title", err)

		if len(path) > 0 && len(title) > 0 {
			fmt.Println("ERROR: Both a path and a title argument have been given. Please only specify one.")
			os.Exit(1)
		}

		if len(path) == 0 && len(title) == 0 {
			fmt.Println("ERROR: Neither a path nor a title have been given. Please give one.")
			os.Exit(1)
		}

		collection, err := store.Collection()
		if err != nil {
			fmt.Println("ERROR: Couldn't get the collection representation of the store.")
			fmt.Println(err)
			os.Exit(1)
		}

		var entry *entries.Entry

		if len(path) > 0 {
			filteredCollection, err := collection.Filter(entries.FilterPathsExact(path))
			if err != nil {
				fmt.Println("ERROR: Couldn't filter the store for the path", path)
				fmt.Println(err)
				os.Exit(1)
			}

			list := filteredCollection.List().Slice()
			if len(list) == 0 {
				fmt.Println("ERROR: That entry doesn't exist.")
				os.Exit(1)
			}

			// Since we're matching a path exactly, there should only be one item in this list.
			entry = list[0]
		}

		if len(title) > 0 {
			filteredCollection, err := collection.Filter(entries.FilterTitlesExact(title))
			if err != nil {
				fmt.Println("ERROR: Couldn't filter the store for the title", title)
				fmt.Println(err)
				os.Exit(1)
			}

			// Since many entries might have the same title, we default to the most recently edited one.
			list := filteredCollection.List().Sort(entries.SortDate).Reverse().Slice()
			if len(list) == 0 {
				fmt.Println("ERROR: That entry doesn't exist.")
				os.Exit(1)
			}

			entry = list[0]
		}

		// We use a 16 digit string to make it easer to determine if it's safe to delete the temporary folder later
		// in the 'vim close' command.
		// For example, if a user opened /tmp/vim-albatross/food/pizza and /tmp/vim-albatross/food/avocados, when
		// finished with one it would be difficult to determine if it was safe to delete the /tmp/vim-albatross/food folder
		// as well as the specific entry.
		temporaryDir := filepath.Join(os.TempDir(), "vim-albatross", randomString(16), entry.Path)
		err = os.MkdirAll(temporaryDir, 0755)
		if err != nil {
			fmt.Println("ERROR: Couldn't create temporary directory", temporaryDir)
			fmt.Println(err)
			os.Exit(1)
		}

		temporaryFile := filepath.Join(temporaryDir, "entry.md")

		err = ioutil.WriteFile(temporaryFile, []byte(entry.OriginalContents), 0755)
		if err != nil {
			fmt.Println("ERROR: Couldn't create temporary file", temporaryFile)
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(temporaryFile)
	},
}

func init() {
	VimOpenCmd.Flags().String("path", "", "path to the entry Vim needs to edit")
	VimOpenCmd.Flags().String("title", "", "title to the entry vim needs to edit")

	VimCmd.AddCommand(VimOpenCmd)
}
