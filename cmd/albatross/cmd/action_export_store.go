package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plus3it/gorecurcopy"
	"github.com/spf13/cobra"
)

// ActionExportStoreCmd represents the 'export store' action.
var ActionExportStoreCmd = &cobra.Command{
	Use:     "store",
	Aliases: []string{"albatross", "folder", "dir"},
	Short:   "output entries in the Albatross store format",
	Long: `store converts the matched entries into the Albatross Store format, a folder containing 'entry.md' files.

If you have the following folders:

	entries/
		school/
			maths/
			physics/
		notes/
			books/
			videos/

You could match all the entries in 'school/' using the command

	$ albatross get -p school/
	school/maths/...
	school/physics/...

You could then use the 'export store' action to create a new folder containing just the school entries:

	$ albatross get -p school/ export store
	# Outputs a folder, by default 'store' in the current directory:

	./
		store/
			school/
				maths/
				physics/

One way of thinking about it is creating filtered stores -- exporting a subset of your entries.

Sub-Entries that Don't Match
----------------------------

Consider the case where you have one top-level entry you _want_ to match that contains a folder with attachments and also more
entries that you _don't_ want to match, like this:

For example:

  an-entry
    entry.md
    attachments/
      hello.jpg              < These are fine to include in the output.
      message.txt            <
    super-secret/            < We don't want to copy this folder!
      entry.md
    attachments-and-secret/
      some-attachment.jpg    < This is fine
      even-more-secret/      < This isn't!
		entry.md
		
We should be left with something that looks like this:

  an-entry/
    entry.md
    attachments/
      hello.jpg
      message.txt
    attachments-and-secret/
	  some-attachment.jpg    < This is fine.
	  `,
	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		outputDest, err := cmd.Flags().GetString("output")
		checkArg(err)

		err = checkOutputDest(outputDest)
		if err != nil {
			fmt.Printf("Cannot output store to %s:\n", outputDest)
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Outputting store to folder: %s\n", outputDest)

		os.Mkdir(outputDest, 0755)

		// The idea seems kind of simple, but there's a couple of subtleties that mean it's a bit more difficult than expected.
		for _, entry := range list.Slice() {
			// This is the path the new entry, so if you were outputting into a folder called "output" and we were currently on
			// the entry "school/a-level/maths/topics", the destPath would be "output/school/a-level/maths/topics".
			// We make the enclosing folder.
			destPath := filepath.Join(outputDest, entry.Path)
			os.MkdirAll(destPath, 0755)

			// This is the original path of to the entry in the store.
			origPath := filepath.Join(store.Path, "entries", entry.Path)

			f, _ := os.Open(origPath)
			fis, _ := f.Readdir(-1)
			f.Close()

			for _, fi := range fis {
				// If it is a directory we need to be careful; we can't just blindly copy the folder because it may contain
				// additional entries that weren't matched in the search.
				//
				// For example:
				//   <current entry folder>
				//     entry.md
				//     attachments/
				//       hello.jpg     < These are fine to include in the output.
				//       message.txt   <
				//
				//     super-secret/   < We don't want to copy this folder!
				//       entry.md
				//
				//     attachments-and-secret/
				//       some-attachment.jpg  < This is fine
				//       even-more-secret/    < This isn't!
				//         entry.md
				//
				// We should be left with something that looks like this:
				//   <current entry folder>
				//     entry.md
				//     attachments/
				//       hello.jpg
				//       message.txt
				//
				//     attachments-and-secret/
				//       some-attachment.jpg  < This is fine.
				//
				// The function copyFolderWithoutEntries handles this.
				if fi.IsDir() {
					copyFolderWithoutEntries(filepath.Join(origPath, fi.Name()), filepath.Join(destPath, fi.Name()))
					continue
				}

				err = gorecurcopy.Copy(filepath.Join(origPath, fi.Name()), filepath.Join(destPath, fi.Name()))
				if err != nil {
					fmt.Println("Error copying file:")
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	},
}

func folderContainsEntry(folder string) (bool, error) {
	containsEntry := false

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "entry.md" {
			containsEntry = true
		}

		if info.IsDir() && path != folder {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return containsEntry, nil
}

// copyFolderWithoutEntries will copy a folder and all it's subdirectories from src to dest but omitting subdirectories that contain entries themselves.
func copyFolderWithoutEntries(src, dest string) error {
	f, _ := os.Open(src)
	fis, _ := f.Readdir(-1)
	f.Close()

	containsEntry, err := folderContainsEntry(src)
	if err != nil {
		return err
	}

	if containsEntry {
		return nil
	}

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		origPath := filepath.Join(src, fi.Name())
		destPath := filepath.Join(dest, fi.Name())

		containsEntries, err := folderContainsEntry(origPath)
		if err != nil {
			return err
		}

		if containsEntries {
			continue
		}

		if fi.IsDir() {
			err = copyFolderWithoutEntries(origPath, destPath)
			if err != nil {
				return err
			}
		}

		err = gorecurcopy.Copy(origPath, destPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// checkOutputDest checks if the argument given is a valid output destination
// An error is returned if it is invalid, otherwise it is nil.
func checkOutputDest(outputDest string) error {
	if !filepath.IsAbs(outputDest) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("couldn't get the current working directory %w", err)
		}

		outputDest = filepath.Join(cwd, outputDest)
	}

	if _, err := os.Stat(outputDest); !os.IsNotExist(err) {
		return fmt.Errorf("directory or file already exists")
	}

	if _, err := os.Stat(filepath.Dir(outputDest)); os.IsNotExist(err) {
		return fmt.Errorf("folder %s does not exist", filepath.Dir(outputDest))
	}

	return nil
}

func init() {
	ActionExportCmd.AddCommand(ActionExportStoreCmd)

	ActionExportStoreCmd.Flags().StringP("output", "o", "entries", "output location of the store, a path. If a folder is specified which doesn't exist, it will be created")
}
