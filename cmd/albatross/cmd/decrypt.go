package cmd

import (
	"fmt"
	"os"

	"github.com/albatross-org/go-albatross/encryption"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	albatross "github.com/albatross-org/go-albatross/pkg/core"
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "decrypt an albatross store",
	Long:  `decrypt will decrypt an albatross store`,
	Run: func(cmd *cobra.Command, args []string) {
		decryptStore()
		fmt.Println("Decrypted.")
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
}

// decryptStore is a utility function for decrypting the store, asking for a password three times.
// It will exit if authentication fails three times.
func decryptStore() {
	var failCount int

	for i := 0; i < 3; i++ {
		err := store.Decrypt(encryption.GetPassword)
		if _, ok := err.(encryption.ErrPrivateKeyDecryptionFailed); ok {
			fmt.Printf("Invalid password. Try again...\n\n")
			failCount++
			continue
		} else if _, ok = err.(albatross.ErrStoreDecrypted); ok {
			fmt.Printf("Store '%s' is already decrypted.\n", storeName)
			break
		} else if err != nil {
			logrus.Fatal(err)
		}

		break
	}

	if failCount == 3 {
		fmt.Println("Decryption failed three times. Exiting.")
		os.Exit(1)
	}
}
