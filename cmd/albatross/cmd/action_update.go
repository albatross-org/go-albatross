package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
)

// ActionUpdateCmd represents the update command
var ActionUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an entry",
	Long: `update an entry from the command line
	
	$ albatross get -p food/pizza update

If multiple entries are matched, a list is displayed to choose from.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)
		var chosen *entries.Entry

		customEditor, err := cmd.Flags().GetString("editor")
		checkArg(err)

		length := len(list.Slice())

		if length == 0 {
			fmt.Println("No entries matched, nothing to update.")
			os.Exit(0)
		} else if length != 1 {
			paths := []string{}
			for _, entry := range list.Slice() {
				paths = append(paths, entry.Path)
			}

			fmt.Println("More than one entry matched, please select one.")
			prompt := promptui.Select{
				Label: "Select Entry",
				Items: paths,
			}

			_, result, err := prompt.Run()
			if err != nil {
				log.Fatalf("Couldn't choose entry: %s", err)
			}

			for _, entry := range list.Slice() {
				if entry.Path == result {
					chosen = entry
					break
				}
			}
		} else {
			chosen = list.Slice()[0]
		}

		updateEntry(chosen, customEditor)
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

	if entry.OriginalContents == content {
		fmt.Println("No change made to entry:", entry.Path)
		return
	}

	err = store.Update(entry.Path, content)
	if err != nil {
		f, tempErr := ioutil.TempFile("", "albatross-recover")
		if tempErr != nil {
			log.Fatal("Couldn't get create temporary file to save recovery entry to. You're on your own! ", err)
		}

		_, err = f.Write([]byte(content))
		if err != nil {
			log.Fatal("Error writing to temporary file to save recovery entry to. You're on your own! ", err)
		}

		fmt.Println("Error updating entry. A copy of the updated file has been saved to:", f.Name())
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated entry:", entry.Path)
}

func init() {
	GetCmd.AddCommand(ActionUpdateCmd)

	ActionUpdateCmd.Flags().StringP("editor", "e", getEditor("vim"), "Editor to use (defaults to $EDITOR, then vim)")
}
