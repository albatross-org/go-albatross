package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path"
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

If you're looking to build or view a standalone HTML-formatted version of a store, you are likely looking for 'albatross export html'
instead which is similar /but handles all configuration automatically.


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

	$ albatross get -p school/ export hugo -o content/posts/albatross

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

Now if you ran:

	$ hugo serve

You could access the entries at a url like http://localhost:1313/posts/albatross/school/maths


Custom Options
--------------

A few options are provided that give a little bit of control over how the entries are generated.

--pre-insert will write some text before the entry content. This is useful for adding shortcodes:

	$ albatross get -p school/ export hugo -o content/posts/albatross --pre-insert "{{< toc >}}"
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	# {{< toc >}}
	# ...entry contents...

--post-insert does the same but writes it at the end of the entry instead.

	$ albatross get -p school/ export hugo -o content/posts/albatross --post-insert "{{< comments >}}"
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	# ...entry contents...
	# {{< comments >}}

--show-metadata appends a small section to each entry's page which shows the original metadata of each entry:

	$ albatross get -p school/ export hugo -o content/posts/albatross
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	#
	# ...entry contents...
	#
	# ###### Metadata
	#  title: 'Original Title'
	#  date: '2021-01-14 09:51'
	#  tags: ["@?tags"]

--show-attachments appends a small section to each entry's page with links to any attachments to the entry:

	$ albatross get -p school/ export hugo -o content/posts/albatross
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	#
	# ...entry contents...
	#
	# ###### Attachments
	# - image.jpg
	# - movie.mov

--metadata can be used to specify custom metadata that will be added to each Hugo post. For example:

	$ albatross get -p school/ export hugo -o content/posts/albatross --metadata toc=true --metadata draft=true
	# Entries now look like:
	# ---
	# ...metadata...
	# toc: true
	# draft: true
	# ---
	# ...entry contents...

Custom metadata given will not show up in the metadata section if --show-metadata is passed.


Relative Path Config
--------------------

--rel-path is an option that controls how links are handeled. If you run this command from inside the Hugo site folder,
then this is automatically handeled because the location of the posts is alread known. However, say you were in a directory
organised like this:
	
	README.md           < You're in this directory.
	file.txt
	hugo-site/
		contents/       < You need to generate posts in contents/posts/albatross

You could run the command:

	$ albatross get -p school/ export hugo -o hugo-site/content/posts/albatross

This would put the posts in the correct folder, but the links would be incorrect because the code generated would assume entries
could be accessed like so:

	http://localhost:1313/hugo-site/content/posts/albatross/...entry...
	
Whereas in reality you have to access posts like so:
	
	http://localhost:1313/posts/albatross/...entry...

So in this case you would set --rel-path to posts/albatross:

	$ albatross get -p school/ export hugo -o hugo-site/content/posts/albatross --rel-path posts/albatross

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

These are the same rules that 'albatross export store' follow.

	  `,
	Run: func(cmd *cobra.Command, args []string) {
		_, collection, list := getFromCommand(cmd)

		outputDest, err := cmd.Flags().GetString("output")
		checkArg(err)

		// Post generation settings.
		relPath, err := cmd.Flags().GetString("rel-path")
		checkArg(err)

		preInsert, err := cmd.Flags().GetStringSlice("pre-insert")
		checkArg(err)

		postInsert, err := cmd.Flags().GetStringSlice("post-insert")
		checkArg(err)

		customMetadata, err := cmd.Flags().GetStringToString("metadata")
		checkArg(err)

		showMetadata, err := cmd.Flags().GetBool("show-metadata")
		checkArg(err)

		showAttachments, err := cmd.Flags().GetBool("show-attachments")
		checkArg(err)

		// If relPath is unset, default to the same as the outputDest but trim any 'content' folder from the beginning.
		if relPath == "" {
			relPath = strings.TrimPrefix(outputDest, "content")
		}

		err = checkOutputDest(outputDest)
		if err != nil {
			fmt.Printf("Cannot output Hugo posts to %s:\n", outputDest)
			fmt.Println(err)
			os.Exit(1)
		}

		// Adjust any relative outputDest to an absolute path.
		if !filepath.IsAbs(outputDest) {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println("Could not establish current working directory:")
				fmt.Println(err)
				os.Exit(1)
			}

			outputDest = filepath.Join(cwd, outputDest)
		}

		contentDir := outputDest

		fmt.Printf("Outputting Hugo posts to folder: %s\n", contentDir)
		generatePostsFolder(contentDir, list, collection, relPath, preInsert, postInsert, customMetadata, showMetadata, showAttachments)
	},
}

// generatePostsFolder generates a Hugo posts folder with the given options.
func generatePostsFolder(outputDest string, list entries.List, collection *entries.Collection, relPath string, preInsert, postInsert []string, customMetadata map[string]string, showMetadata, showAttachments bool) {
	var err error
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

			// This is the main difference from the 'export store' option -- we need to transform the entry.md files.
			switch fi.Name() {
			case "entry.md":
				// If it is an entry:
				// We can use the contents of the entry.md file and the value of entry.OriginalContents interchangeably. In code:
				//   fileBytes, _ := ioutil.ReadFile(filepath.Join(origPath, fi.Name()))
				//   string(fileBytes) == entry.OriginalContents => true

				newContents, err := hugoifyEntry(
					collection, entry, relPath,
					preInsert, postInsert, customMetadata, showMetadata, showAttachments,
				)
				if err != nil {
					fmt.Println("Error converting entry into Hugo post:")
					fmt.Println(err)
					os.Exit(1)
				}

				err = ioutil.WriteFile(filepath.Join(destPath, "entry.md"), []byte(newContents), 0644)
				if err != nil {
					fmt.Println("Error writing Hugo post:")
					fmt.Println(err)
					os.Exit(1)
				}

			case "_index.md":
				// This is an unlikely case but it should be handeled regardless, even if it's not handeled very well.
				// Here we create a new filename by generating a random filename... it's not fast or pretty but it should prevent
				// most problems. One potential issue here is that if this file was referenced in the main `entry.md` file, the reference
				// will no longer be correct.
				filename := fmt.Sprintf("_index_%d.md", rand.Int())
				err = gorecurcopy.Copy(filepath.Join(origPath, fi.Name()), filepath.Join(destPath, filename))
				if err != nil {
					fmt.Println("Error copying file:")
					fmt.Println(err)
					os.Exit(1)
				}

			default:
				// If it's a non-entry file (like an attachment), and isn't named "_index.md" (what Hugo recognises as the main page for a folder),
				// we can just copy it normally.
				err = gorecurcopy.Copy(filepath.Join(origPath, fi.Name()), filepath.Join(destPath, fi.Name()))
				if err != nil {
					fmt.Println("Error copying file:")
					fmt.Println(err)
					os.Exit(1)
				}

			}
		}
	}
}

// hugoifyEntry outputs the contents of an entry as if it was a Hugo post.
// This mainly consists of correcting the front matter and adjusting the links to the correct format.
// metadata controls whether to add the metadata as a code block at the end of the entry.
// toc controls whether to set the "toc: true" option for all the entries.
func hugoifyEntry(collection *entries.Collection, entry *entries.Entry, relPath string, preInsert, postInsert []string, customMetadata map[string]string, showMetadata, showAttachments bool) (string, error) {
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

	// Set the URL parameter so that every URL doesn't have "/entry" on the end of it. This means that you can access the entry located
	// at a path just by going https://example.com/posts/<path>
	entry.Metadata["url"] = filepath.Join(relPath, entry.Path)

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
	out.WriteString("---\n")

	// If any pre inserts are given, insert them here:
	for _, shortcode := range preInsert {
		out.WriteString(shortcode)
		out.WriteString("\n")
	}

	entryContents := entry.Contents

	// Handle links properly by replacing '{{ }}' and '[[ ]]' with actual markdown links.
	for _, link := range entry.OutboundLinks {
		linkedEntry := collection.ResolveLink(link)
		text := entry.Contents[link.Loc[0]:link.Loc[1]]

		if linkedEntry != nil {
			var hugoLink string
			if relPath == "" {
				hugoLink = fmt.Sprintf(`[%s]({{< relref "/%s/entry.md" >}})`, markdownEscape(text), linkedEntry.Path)
			} else {
				hugoLink = fmt.Sprintf(`[%s]({{< relref "/%s/%s/entry.md" >}})`, markdownEscape(text), relPath, linkedEntry.Path)
			}

			entryContents = strings.ReplaceAll(entryContents, text, hugoLink)
		} else {
			hugoLink := fmt.Sprintf(`[%s](/404.html)`, markdownEscape(text))
			entryContents = strings.ReplaceAll(entryContents, text, hugoLink)
		}
	}

	// Write the modified contents of the entry.
	out.WriteString(entryContents)

	// If any post inserts are given, insert them here:
	for _, shortcode := range postInsert {
		out.WriteString(shortcode)
		out.WriteString("\n")
	}

	// If showing metadata or attachments, create a seperator.
	if showMetadata || showAttachments {
		out.WriteString("\n___\n")
	}

	// Write a metadata section.
	if showMetadata {
		out.WriteString("###### Metadata\n")
		out.WriteString("```yaml\n")
		out.WriteString(string(origMetadata))
		out.WriteString("```\n")
	}

	// If we want to show attachments and the entry has attachments, then do so.
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

func createHugoSite(outputDest, theme, relPath string) {
	themeURL, err := url.Parse(theme)
	if err != nil {
		fmt.Println("Please specify --theme as a valid URL to a git repository.")
		fmt.Println(err)
		os.Exit(1)
	}
	themeName := path.Base(themeURL.Path)

	if !commandExists("hugo") {
		fmt.Println("--standalone option given but hugo command not on path. Is Hugo installed?")
		os.Exit(1)
	}

	if !commandExists("git") {
		fmt.Println("--standalone option given but git not on path. Why on earth isn't git installed?")
		os.Exit(1)
	}

	fmt.Println("# Creating new site directory via Hugo:")
	cmd := exec.Command("hugo", "new", "site", outputDest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("\n# Error creating site with Hugo... there is probably output above.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("\n# Adding theme via git:")
	cmd = exec.Command("git", "clone", theme, filepath.Join(outputDest, "themes", themeName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("\n# Couldn't add theme... there is probably output above.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("\n# Updating config:")
	f, err := os.OpenFile(filepath.Join(outputDest, "config.toml"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("\ntheme = '%s'\n", themeName)); err != nil {
		log.Println(err)
	}
	fmt.Println("Appended", fmt.Sprintf("\n theme = '%s'\n", themeName), "to config.toml.")

	fmt.Println("")
}

func init() {
	ActionExportCmd.AddCommand(ActionExportHugoCmd)

	ActionExportHugoCmd.Flags().StringP("output", "o", "entries", "output location of the posts, a path. If a folder is specified which doesn't exist, it will be created")
	ActionExportHugoCmd.Flags().String("rel-path", "", "path to make relative links from, e.g. 'school/a-level/physics' -> 'posts/notes/school/a-level/physics' if set to 'posts/notes'. Defaults to the same as output but trims off a leading 'content' if it exists")

	ActionExportHugoCmd.Flags().StringSlice("pre-insert", []string{}, "insert these strings before entry content, useful for inserting shortcodes")
	ActionExportHugoCmd.Flags().StringSlice("post-insert", []string{}, "insert these strings after entry content, useful for inserting shortcodes")
	ActionExportHugoCmd.Flags().StringToStringP("metadata", "m", map[string]string{}, "insert this metadata into every entry, such as '-m draft=true'")

	ActionExportHugoCmd.Flags().Bool("show-metadata", false, "print a section at the end of every entry showing the original metadata")
	ActionExportHugoCmd.Flags().Bool("show-attachments", false, "print a section at the end of every entry showing links to attachments")
}
