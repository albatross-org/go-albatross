package encryption

import "fmt"

// ErrPrivateKeyDecryptionFailed occurs when the private key cannot be encrypted. This error is normally the fault of the user
// and means that the program should ask for the password again.
type ErrPrivateKeyDecryptionFailed struct {
	PathToPrivateKey string
	Err              error
}

// Error returns the error message.
func (e ErrPrivateKeyDecryptionFailed) Error() string {
	return fmt.Sprintf("couldn't decrypt private key (%s): %s", e.PathToPrivateKey, e.Err)
}
