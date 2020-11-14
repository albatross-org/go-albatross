package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ActionAttachCmd represents the 'date' action.
var ActionAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "attach files",
	Long: `attach attaches files to entries.
	
	$ albatross get --path-exact roadtrip attach ~/documents/photos/roadtrip
	# Attach the '~/documents/roadtrip/photos' folder to the roadtrip entry

Multiple attachments can be specified as multiple arguments:

	$ albatross get --path-exact roadtrip attach ~/documents/photos/roadtrip/2020-11-14.jpg ~/documents/photos/roadtrip/2020-11-15.jpg
	# Attach the files 2020-11-14.jpg and 2020-11-15.jpg to the 'roadtrip' entry.
	
Attachments are added in a slightly confusing way. By default the file that is being attached will be copied to the 'attachments/'
folder in the root of the Albatross store:

	store/
		entries/
		templates/
		attachments/   <--
		config.yaml
		
And then a symlink will be created inside the matched entries that points from the attachments folder into the entry itself.

For example, consider the 'roadtrip' entry from above:

	store/
		entries/
			roadtrip/
				entry.md
				
		attachments/
		
Running the command

	$ albatross get --path-exact roadtrip attach ~/documents/photos/roadtrip/2020-11-14.jpg

Copies the photo 2020-11-14.jpg into the 'attachments/' folder, named after the SHA-256 hash of the file and a
symbolic link is then created in the 'roadtrip/' entry which looks like this:

	store/
		entries/
			roadtrip/
				2020-11-14.jpg => ../../attachments/attachment-d1716f9f598eeb6922f6aeca390a1b27c3075901ed2216906bd03e8ef90b963b
				
		attachments/
			attachment-d1716f9f598eeb6922f6aeca390a1b27c3075901ed2216906bd03e8ef90b963b
			
For a folder, the process is the same apart from the hash is calculated from the concatenation of all the hashes of the subdirectories
it contains.

This has a few side effects. One of them is that if you synchronise two stores on different computers, media and attachments present on one
computer will not be synced via Git. This is on purpose and the main reason for the slightly strange implementation; it means massive files
such as PDFs, photos and videos are not tracked and stored in Git.

Furthermore, encrypting a store will not encrypt the attachments. This is again on purpose because otherwise it would mean waiting a long time
every time the store is encrypted and decrypted.

If you would like the files synced via Git and encrypted, you can instead make direct copies using the 'copy' flag.

	$ albatross get --path-exact roadtrip attach --copy 2020-11-14.jpg

`,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		copyAttachment, err := cmd.Flags().GetBool("copy")
		checkArg(err)

		noConfirm, err := cmd.Flags().GetBool("no-confirm-multiple-attach")
		checkArg(err)

		if len(args) == 0 {
			fmt.Println("Please give something to attach. For example:")
			fmt.Println("$ albatross get --path-exact roadtrip attach --copy 2020-11-14.jpg")
			os.Exit(1)
		}

		for _, arg := range args {
			_, err := os.Stat(arg)
			if err != nil && os.IsNotExist(err) {
				fmt.Println(arg, "is not a valid path.")
				os.Exit(1)
			}
		}

		slice := list.Slice()

		if len(slice) > 1 && !noConfirm {
			fmt.Printf("You're about to attach the same thing to %d entries.\n", len(slice))
			confirm := confirmPrompt("Are you sure you wish to continue?")

			if !confirm {
				fmt.Println("Not continuing.")
			}
		}

		for _, arg := range args {
			for _, entry := range list.Slice() {
				if copyAttachment {
					err := store.AttachCopy(entry.Path, arg)
					if err != nil {
						fmt.Printf("Error attaching a copy of %s to %s:\n", arg, entry.Path)
						fmt.Println(err)
					}
				} else {
					err := store.AttachSymlink(entry.Path, arg)
					if err != nil {
						fmt.Printf("Error attaching a copy of %s to %s:\n", arg, entry.Path)
						fmt.Println(err)
					}
				}
			}
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionAttachCmd)

	ActionAttachCmd.Flags().Bool("copy", false, "move the original file rather than copying")
	ActionAttachCmd.Flags().Bool("no-confirm-multiple-attach", false, "whether to prevent confirming attaching the same file to multiple entries")
}
