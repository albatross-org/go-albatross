package entries

import "github.com/spf13/afero"

var (
	// Os points to the real OS filesystem.
	Os = &afero.OsFs{}
)

// NewBaseFs returns a new file system for an albatross folder.
// It is read-only (to prevent accidental side-effects) and deals with relative paths behind the scenes.
// For example, "food/pizza" will be converted into "/path/to/albatross/entries/food/pizza"
func NewBaseFs(basePath string) afero.Fs {
	return afero.NewBasePathFs(afero.NewReadOnlyFs(Os), basePath)
}
