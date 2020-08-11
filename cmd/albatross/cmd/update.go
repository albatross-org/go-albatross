package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/albatross-org/go-albatross/entries"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an entry",
	Long: `update an entry from the command line
	
$ albatross update food/pizza`,
	Run: func(cmd *cobra.Command, args []string) {
		encrypted, err := store.Encrypted()
		if err != nil {
			log.Fatal(err)
		} else if encrypted {
			decryptStore()

			if !leaveDecrypted {
				defer encryptStore()
			}
		}

		editorName := getEditor("vim")
		customEditor, err := cmd.Flags().GetString("editor")
		if err != nil {
			log.Fatal("Couldn't get custom editor: ", err)
		}

		if customEditor != "" {
			editorName = customEditor
		}

		if len(args) != 1 {
			fmt.Println("Expecting exactly one argument: path to entry")
			fmt.Println("For example:")
			fmt.Println("")
			fmt.Println("$ albatross update food/pizza")
		}

		collection, err := store.Collection()
		if err != nil {
			log.Fatal("Error parsing the Albatross store: ", err)
		}

		es, err := collection.Filter(entries.FilterPathsExact(args[0]))
		if err != nil {
			log.Fatalf("Error filtering entries for ones with the path %s: %s", args[0], err)
		}

		if es.Len() == 0 {
			fmt.Println("Couldn't find any entries with the path", args[0])
			os.Exit(1)
		} else if es.Len() > 1 {
			fmt.Println("Multiple entires have the path")
		}

		entry := es.List().Slice()[0]

		updateEntry(entry, editorName)
	},
}

func updateEntry(entry *entries.Entry, editorName string) {
	content, err := edit(
		editorName,
		entry.OriginalContents,
	)
	if err != nil {
		log.Fatal("Couldn't get content from editor: ", err)
	}

	err = store.Update(entry.Path, content)
	if err != nil {
		f, tempErr := ioutil.TempFile("", "albatross-recover")
		if tempErr != nil {
			logrus.Fatal("Couldn't get create temporary file to save recovery entry to. You're on your own! ", err)
		}

		f.Write([]byte(content))

		fmt.Println("Error updating entry. A copy of the updated file has been saved to:", f.Name())
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated entry:", entry.Path)
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("editor", "e", "", "Editor to use (defaults to $EDITOR, then vim)")
}
