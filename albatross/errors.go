package albatross

import "fmt"

// ErrStoreEncrypted is returned when an action is attempted that requires a decrypted store but the store is decrypted.
type ErrStoreEncrypted struct {
	Path string
}

// Error returns the error message.
func (e ErrStoreEncrypted) Error() string {
	return fmt.Sprintf("store %s is currently encrypted", e.Path)
}

// ErrStoreDecrypted is returned when a store is asked to be decrypted but it's already decrypted.
type ErrStoreDecrypted struct {
	Path string
}

// Error returns the error message.
func (e ErrStoreDecrypted) Error() string {
	return fmt.Sprintf("store %s is already decrypted", e.Path)
}

// ErrEntryDoesntExist is returned when the entry requested doesn't exist.
type ErrEntryDoesntExist struct {
	Path string
}

// Error returns the error message.
func (e ErrEntryDoesntExist) Error() string {
	return fmt.Sprintf("entry %s doesn't exist", e.Path)
}

// ErrEntryAlreadyExists is returned when the entry requested already exists.
type ErrEntryAlreadyExists struct {
	Path string
}

// Error returns the error message.
func (e ErrEntryAlreadyExists) Error() string {
	return fmt.Sprintf("entry %s already exists", e.Path)
}

// ErrLoadStoreInvalid is returned if you try and load an invalid store.
// For example, trying to load a store named "phd" when a store of that name isn't present in the config.
type ErrLoadStoreInvalid struct {
	Name       string
	ConfigPath string
}

// Error returns the error message.
func (e ErrLoadStoreInvalid) Error() string {
	return fmt.Sprintf("unknown store %s in config %s", e.Name, e.ConfigPath)
}
