package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var defaultEntry = `---
title: "%s"
date: "%s"
---

`

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "create a new entry",
	Aliases: []string{"new"},
	Long: `create a new entry from the command line
	
$ albatross create food/pizza`,
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
			fmt.Println("$ albatross create food/pizza")
		}

		// Here we create an empty entry first, then update it.
		// This means that an error like "EntryAlreadyExists" will come up now rather than
		// after the entry has been created, which could lead to data loss.
		err = store.Create(args[0], fmt.Sprintf(defaultEntry, "Temp", time.Now().Format("2006-01-02 15:04")))
		if err != nil {
			log.Fatal("Couldn't create entry: ", err)
		}

		content, err := edit(
			editorName,
			fmt.Sprintf(defaultEntry, "", time.Now().Format("2006-01-02 15:04")),
		)
		if err != nil {
			log.Fatal("Couldn't get content from editor: ", err)
		}

		err = store.Update(args[0], content)
		if err != nil {
			f, err := ioutil.TempFile("", "albatross-recover")
			if err != nil {
				logrus.Fatal("Couldn't get create temporary file to save recovery entry to. You're on your own! ", err)
			}

			f.Write([]byte(content))

			fmt.Println("Error creating entry. A copy has been saved to:", f.Name())
			os.Exit(1)
		}

		fmt.Println("Successfully created entry", args[0])
	},
}

func init() {
	rootCmd.AddCommand(CreateCmd)
	CreateCmd.Flags().StringP("editor", "e", "", "Editor to use (defaults to $EDITOR, then vim)")
}
