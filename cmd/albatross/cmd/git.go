package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// GitCmd represents the git command
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Interface with git in a store",
	Long: `git lets you access git version control within the store.

Basically, it's a shorthand for doing

	$ albatross decrypt && cd $ALBATROSS_DIR && git... && albatross encrypt

For example:

	$ albatross git add .
	$ albatross git commit
	$ albatross git push

To pass flags to git, use the "--" seperator.

	$ albatross git -- log --oneline
	$ albatross git -- commit -m "commit message"`,
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

		if !store.UsingGit() {
			fmt.Printf("Store '%s' not using Git.\n", storeLocation)
			os.Exit(0)
		}

		newArgs := []string{"--git-dir", filepath.Join(storePath, "entries", ".git"), "--work-tree", filepath.Join(storePath, "entries")}
		c := exec.Command("git", append(newArgs, args...)...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		err = c.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(GitCmd)
}
