package encryption

import (
	"fmt"

	"golang.org/x/crypto/ssh/terminal"
)

// GetPassword prompts a user for a password.
func GetPassword() (string, error) {
	fmt.Printf("Password: ")

	bytes, err := terminal.ReadPassword(0)
	if err != nil {
		return "", fmt.Errorf("couldn't read password: %w", err)
	}

	fmt.Print("\n")

	return string(bytes), nil
}
