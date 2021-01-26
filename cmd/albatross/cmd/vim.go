package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// VimCmd represents the vim command
var VimCmd = &cobra.Command{
	Use:   "vim",
	Short: "Integration between Vim and Albatross",
	Long: `vim is the command that allows the Vim plugin, vim-albatross, to work.

This command shouldn't need to be used directly. Instead, install the Vim plugin:

    (init.vim/vimrc)
    Plug 'albatross-org/vim-albatross' " Using vim-plugged for example.

Information about how the plugin works is avaliable through Vim using :h vim-albatross.

I Actually Want To Use This Command
-----------------------------------

    COMMAND                FUNCTION
    vim open               Creates and returns a temporary file for Vim to edit.

    vim close              Takes the path to a temporary file and makes the changes in the
                           file.

How It Works
------------

This command is used to start and maintain a conversation with vim-albatross:

    vim-albatross                                   'albatross vim' command
    =============                                   =======================

    "Can I have a temporary file to edit
    the entry school/physics?"
    | VimAlbatross#OpenEntry (vim) 
    | albatross vim open school/physics (shell)


                                                    "Sure thing. Here it is:
                                                    /tmp/vim-albatross/<path>/entry.md"
                                                    (internal logic in this command)


    ...Makes some edits to temp file...
    On close:
    "Thanks, I'm done with this file."
    | VimAlbatross#CloseEntry (vim)
    | albatross vim close school/physics (shell)
    

                                                    "Ok, I'll reflect those changes in the
                                                    the store."
                                                    ...Makes updates to actual entry...
                                                    ...Updates Git...
                                                    ...Makes sure nothing is corrupted or lost...

    `,
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

	},
}

func init() {
	rootCmd.AddCommand(VimCmd)
}
