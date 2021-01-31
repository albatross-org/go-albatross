package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/albatross-org/go-albatross/albatross"
)

// EncryptCmd represents the encrypt command
var EncryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt a store",
	Long: `encrypt will encrypt a store

For example:

	$ albatross encrypt
	Encrypting... done in 45ms`,
	Run: func(cmd *cobra.Command, args []string) {
		encryptStore()
	},
}

func init() {
	rootCmd.AddCommand(EncryptCmd)
}

// encryptStore will encrypt an albatross store.
func encryptStore() {
	fmt.Print("Encrypting... ")
	start := time.Now()

	err := store.Encrypt()
	if _, ok := err.(albatross.ErrStoreEncrypted); ok {
		fmt.Printf("Store '%s' is already encrypted.", storeLocation)
	} else if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("done in %s\n", time.Since(start))
}
