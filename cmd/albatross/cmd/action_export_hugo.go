package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/plus3it/gorecurcopy"
	"github.com/spf13/cobra"
)

// ActionExportHugoCmd represents the 'export hugo' action.
var ActionExportHugoCmd = &cobra.Command{
	Use:   "hugo",
	Short: "output entries as content for a Hugo site",
	Long: `hugo exports matched entries in a folder structure that can be used as content for a Hugo site.

For example, if you have a Hugo site such as:

	content/
	data/
	layouts/
	public/
	resources/
	static/
	themes/
	config.toml

Running:

	$ albatross get export hugo -o content/posts/albatross

Would create the 'albatross' folder in the content/posts folder containing Hugo formatted entries.

Example: School Notes
---------------------

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

You could then use the 'export hugo' action to create a folder of Hugo posts:

	$ albatross get -p school/store export hugo -o content/posts/albatross

	config.toml
	data/
	layouts/
	public/
	resources/
	static/
	themes/
	content/
		posts/
			albatross/
				school/
					maths/
					physics/

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
		_, collection, list := getFromCommand(cmd)

		outputDest, err := cmd.Flags().GetString("output")
		checkArg(err)

		relPath, err := cmd.Flags().GetString("rel-path")
		checkArg(err)
		if relPath == "" {
			relPath = outputDest
			relPath = strings.TrimLeft(relPath, "content")
		}

		preShortcodes, err := cmd.Flags().GetStringSlice("pre-shortcodes")
		checkArg(err)

		postShortcodes, err := cmd.Flags().GetStringSlice("post-shortcodes")
		checkArg(err)

		customMetadata, err := cmd.Flags().GetStringToString("metadata")
		checkArg(err)

		adjustURLs, err := cmd.Flags().GetBool("adjust-urls")
		checkArg(err)

		showMetadata, err := cmd.Flags().GetBool("show-metadata")
		checkArg(err)

		showAttachments, err := cmd.Flags().GetBool("show-attachments")
		checkArg(err)

		if !filepath.IsAbs(outputDest) {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println("Couldn't get the current working directory:")
				fmt.Println(err)
				os.Exit(1)
			}

			outputDest = filepath.Join(cwd, outputDest)
		}

		if _, err := os.Stat(outputDest); !os.IsNotExist(err) {
			fmt.Printf("Cannot output Hugo posts to %s:\n", outputDest)
			fmt.Println("Directory/file already exists.")
			os.Exit(1)
		}

		if _, err := os.Stat(filepath.Dir(outputDest)); os.IsNotExist(err) {
			fmt.Printf("Cannot output Hugo posts to %s: \n", outputDest)
			fmt.Printf("Folder %s does not exist\n", filepath.Dir(outputDest))
			os.Exit(1)
		}

		fmt.Printf("Outputting Hugo posts to folder: %s\n", outputDest)

		os.Mkdir(outputDest, 0755)

		// The idea seems kind of simple, but there's a couple of subtleties that mean it's a bit more difficult than expected.
		for _, entry := range list.Slice() {
			// This is the path the new entry, so if you were outputting into a folder called "output" and we were currently on
			// the entry "school/a-level/maths/topics", the destPath would be "output/school/a-level/maths/topics".
			// We make the enclosing folder.
			destPath := filepath.Join(outputDest, entry.Path)
			os.MkdirAll(destPath, 0755)

			// This is the original path of to the entry in the store.
			origPath := filepath.Join(storePath, "entries", entry.Path)

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

				// This is the main difference from the 'export store' option -- we need to transform the entry.md files.
				if fi.Name() != "entry.md" {
					// If it's a non-entry file (like an attachment), just copy it normally.
					err = gorecurcopy.Copy(filepath.Join(origPath, fi.Name()), filepath.Join(destPath, fi.Name()))
					if err != nil {
						fmt.Println("Error copying file:")
						fmt.Println(err)
						os.Exit(1)
					}
				} else {
					// If it is an entry:
					// We can use the contents of the entry.md file and the value of entry.OriginalContents interchangeably. In code:
					//   fileBytes, _ := ioutil.ReadFile(filepath.Join(origPath, fi.Name()))
					//   string(fileBytes) == entry.OriginalContents => true

					newContents, err := hugoifyEntry(
						collection, entry, relPath,
						preShortcodes, postShortcodes, customMetadata, adjustURLs, showMetadata, showAttachments,
					)
					if err != nil {
						fmt.Println("Error converting entry into Hugo post:")
						fmt.Println(err)
						os.Exit(1)
					}

					err = ioutil.WriteFile(filepath.Join(destPath, fi.Name()), []byte(newContents), 0644)
					if err != nil {
						fmt.Println("Error writing Hugo post:")
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
		}
	},
}

// hugoifyEntry outputs the contents of an entry as if it was a Hugo post.
// This mainly consists of correcting the front matter and adjusting the links to the correct format.
// metadata controls whether to add the metadata as a code block at the end of the entry.
// toc controls whether to set the "toc: true" option for all the entries.
func hugoifyEntry(collection *entries.Collection, entry *entries.Entry, relPath string, preShortcodes, postShortcodes []string, customMetadata map[string]string, adjustURLs, showMetadata, showAttachments bool) (string, error) {
	var out bytes.Buffer
	var origMetadata []byte
	var err error

	// If printing metadata at the end of the entry, we need to Marshal it now so it doesn't include the changes we've made
	if showMetadata {
		origMetadata, err = yaml.Marshal(entry.Metadata)
		if err != nil {
			return "", err
		}
	}

	// Convert the date from a string to an actual date ready for YAML export.
	entry.Metadata["date"] = entry.Date

	// If adjusting URLs, rewrite the URL to the correct one:
	entry.Metadata["url"] = filepath.Join(relPath, entry.Path)
	// TODO Check if correct

	// Set all custom metadata:
	for k, v := range customMetadata {
		entry.Metadata[k] = v
	}

	// Finally, marshal the new front-matter.
	frontmatter, err := yaml.Marshal(entry.Metadata)
	if err != nil {
		return "", err
	}

	// Write the new, modified front matter.
	out.WriteString("---\n")
	out.Write(frontmatter)
	out.WriteString("\n---\n")

	entryContents := entry.Contents

	// If any pre-shortcodes are given, insert them here:
	for _, shortcode := range preShortcodes {
		out.WriteString(shortcode)
		out.WriteString("\n")
	}

	// Handle links properly by replacing '{{ }}' and '[[ ]]' with actual markdown links.
	for _, link := range entry.OutboundLinks {
		linkedEntry := collection.ResolveLink(link)
		text := entry.Contents[link.Loc[0]:link.Loc[1]]

		if linkedEntry != nil {
			hugoLink := fmt.Sprintf(`[%s]({{< relref "/%s/%s/entry" >}})`, markdownEscape(text), relPath, linkedEntry.Path)
			entryContents = strings.ReplaceAll(entryContents, text, hugoLink)
		} else {
			hugoLink := fmt.Sprintf(`[%s](/404.html)`, markdownEscape(text))
			entryContents = strings.ReplaceAll(entryContents, text, hugoLink)
		}
	}

	// Write the modified contents of the entry.
	out.WriteString(entryContents)

	// If any post shortcodes are given, insert them here:
	for _, shortcode := range postShortcodes {
		out.WriteString(shortcode)
		out.WriteString("\n")
	}

	// If showing metadata or attachments, create a seperator.
	if showMetadata || showAttachments {
		out.WriteString("\n___\n")
	}

	if showMetadata {
		out.WriteString("###### Metadata\n")
		out.WriteString("```yaml\n")
		out.WriteString(string(origMetadata))
		out.WriteString("```\n")
	}

	if showAttachments && len(entry.Attachments) != 0 {
		out.WriteString("###### Attachments\n")
		for _, attachment := range entry.Attachments {
			out.WriteString(fmt.Sprintf("- [%s](/%s)\n", attachment.Name, filepath.Join(relPath, attachment.RelPath)))
		}
	}

	// Return the final modified entry.
	return out.String(), nil
}

func markdownEscape(str string) string {
	str = strings.ReplaceAll(str, "[", "\\[")
	str = strings.ReplaceAll(str, "]", "\\]")
	return str
}

func init() {
	ActionExportCmd.AddCommand(ActionExportHugoCmd)

	ActionExportHugoCmd.Flags().StringP("output", "o", "entries", "output location of the posts, a path. If a folder is specified which doesn't exist, it will be created")
	ActionExportHugoCmd.Flags().String("rel-path", "", "path to make relative links from, e.g. 'school/a-level/physics' -> 'posts/notes/school/a-level/physics' if set to 'posts/notes'. Defaults to the same as output but trims off a leading 'content' if it exists")

	ActionExportHugoCmd.Flags().StringSlice("pre-shortcodes", []string{}, "insert these shortcodes before entry content")
	ActionExportHugoCmd.Flags().StringSlice("post-shortcodes", []string{}, "insert these shortcodes after entry content")
	ActionExportHugoCmd.Flags().StringToStringP("metadata", "m", map[string]string{}, "insert this metadata into every entry, such as '-m draft=true'")

	ActionExportHugoCmd.Flags().Bool("adjust-urls", true, "adjust post URLs so that relative links from entries (such as referencing a photo in the same folder) will work properly")
	ActionExportHugoCmd.Flags().Bool("show-metadata", false, "print a section at the end of every entry showing the original metadata")
	ActionExportHugoCmd.Flags().Bool("show-attachments", false, "print a section at the end of every entry showing links to attachments")

}
