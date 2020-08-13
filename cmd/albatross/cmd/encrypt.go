package cmd

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	albatross "github.com/albatross-org/go-albatross/pkg/core"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "encrypt an albatross store",
	Long: `encrypt will encrypt an albatross store

For example:

$ albatross encrypt
Encrypting... done in 45ms`,
	Run: func(cmd *cobra.Command, args []string) {
		encryptStore()
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)
}

// encryptStore will encrypt an albatross store.
func encryptStore() {
	fmt.Print("Encrypting... ")
	start := time.Now()

	err := store.Encrypt()
	if _, ok := err.(albatross.ErrStoreEncrypted); ok {
		fmt.Printf("Store '%s' is already encrypted.", storeName)
	} else if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("done in %s\n", time.Now().Sub(start))
}
