package entries

import "fmt"

// ErrEntryReadFailed is returned when an entry cannot be read.
type ErrEntryReadFailed struct {
	Path string
	Err  error
}

// Error returns a string representing the error.
func (e ErrEntryReadFailed) Error() string {
	return fmt.Sprintf("could not read entry file %q: %s", e.Path, e.Err)
}

// Unwrap returns the error embedded in the ErrEntryReadFailed.
func (e ErrEntryReadFailed) Unwrap() error { return e.Err }

// ErrEntryParseFailed is returned when the entry.md file cannot be parsed.
type ErrEntryParseFailed struct {
	Path string
	Err  error
}

// Error returns a string representing the error.
func (e ErrEntryParseFailed) Error() string {
	return fmt.Sprintf("could not parse entry file %q: %s", e.Path, e.Err)
}

// Unwrap returns the error embedded in the ErrEntryParseFailed.
func (e ErrEntryParseFailed) Unwrap() error { return e.Err }

// ErrEntryAlreadyExists is returned by an EntryGraph when you attempt to add a duplicate entry.
type ErrEntryAlreadyExists struct {
	Path  string
	Title string
}

// Error returns a string representing the error.
func (e ErrEntryAlreadyExists) Error() string {
	return fmt.Sprintf("entry '%s' (%s) already exists", e.Title, e.Path)
}

// ErrEntryDoesntExist is returned by an EntryGraph when you attempt to delete a non-existant entry.
type ErrEntryDoesntExist struct {
	Path  string
	Title string
}

// Error returns a string representing the error.
func (e ErrEntryDoesntExist) Error() string {
	return fmt.Sprintf("entry '%s' (%s) doesnt exist", e.Title, e.Path)
}
