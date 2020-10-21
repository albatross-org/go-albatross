package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/bmaupin/go-epub"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

// ActionExportEpubCmd represents the 'tags' action.
var ActionExportEpubCmd = &cobra.Command{
	Use:   "epub",
	Short: "convert entries to an EPUB format",
	Long: `epub converts matched entries into an EPUB format, ready for loading on to a device like a Kindle.
	
It's often a good idea to sort entries when creating an EPUB because then the Chapters will be in the correct order in the export.

	$ albatross get -p school --sort export epub`,

	Run: func(cmd *cobra.Command, args []string) {
		_, collection, list := getFromCommand(cmd)
		command := "albatross " + strings.Join(os.Args[1:], " ")

		author, err := cmd.Flags().GetString("book-author")
		checkArg(err)

		title, err := cmd.Flags().GetString("book-title")
		checkArg(err)

		if author == "" {
			author = command
		}

		if title == "" {
			title = "Albatross " + time.Now().Format("2006-01-02 15:01")
		}

		outputDest, err := cmd.Flags().GetString("output")
		checkArg(err)

		output, err := convertToEpub(collection, list, title, author, command)
		if err != nil {
			fmt.Println("Error when creating the EPUB:")
			fmt.Println(err)
			os.Exit(1)
		}

		if outputDest != "" {
			err = ioutil.WriteFile(outputDest, output, 0644)
			if err != nil {
				fmt.Println("Couldn't write to output destination:")
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Println(string(output))
		}
	},
}

// convertToEpub returns an EPUB file built from the list of entries specified. It also takes an argument
// for the title and author.
func convertToEpub(collection *entries.Collection, list entries.List, title, author, command string) ([]byte, error) {
	e := epub.NewEpub(title)
	e.SetAuthor(author)

	md := goldmark.New(
		goldmark.WithRendererOptions(html.WithXHTML()),
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
	)

	info := `<h1>Info</h1>
	<p>This EPUB was generated <pre>%s</pre> by the command <pre>%s</pre>matching<pre>%d</pre> entries.</p>
	<ul>
		<li><a href="toc.xhtml">Table of Contents</a></li>
		<li><a href="tags.xhtml">Tags Search</a></li>
		<li><a href="paths.xhtml">Path Search</a></li>
	</ul>`

	_, err := e.AddSection(fmt.Sprintf(info, time.Now().Format(time.RFC3339), command, len(list.Slice())), "Info", "info.xhtml", "")
	if err != nil {
		return nil, err
	}

	toc, err := epubBuildTableOfContents(list)
	if err != nil {
		return nil, err
	}
	_, err = e.AddSection(toc, "Table of Contents", "toc.xhtml", "")
	if err != nil {
		return nil, err
	}

	tags, err := epubBuildTagSearch(collection, list)
	if err != nil {
		return nil, err
	}
	_, err = e.AddSection(tags, "Tags", "tags.xhtml", "")
	if err != nil {
		return nil, err
	}

	paths := epubBuildPathSearch(list)
	_, err = e.AddSection(paths, "Paths", "paths.xhtml", "")
	if err != nil {
		return nil, err
	}

	for _, entry := range list.Slice() {
		contents, title, path, err := epubEntryToXHTML(md, collection, entry)
		if err != nil {
			return nil, err
		}

		_, err = e.AddSection(contents, title, path, "")
		if err != nil {
			return nil, fmt.Errorf("error adding section for entry %s: %w", entry.Path, err)
		}
	}

	_, err = e.AddSection(`<h1>Unknown</h1><p>The entry you linked to either doesn't exist or wasn't matched.</p>`, "Unknown", "unknown.xhtml", "")
	if err != nil {
		return nil, err
	}

	dir, err := ioutil.TempDir("", "")
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("Error removing temp directory %s: %s", dir, err)
			return
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory for EPUB: %s", err)
	}

	output := path.Join(dir, "book.epub")

	err = e.Write(output)
	if err != nil {
		return nil, fmt.Errorf("error writing epub file to %s: %s", output, err)
	}

	return ioutil.ReadFile(output)
}

// epubBuildTableOfContents creates the XHTML for a Table of Contents, built from a list of entries.
func epubBuildTableOfContents(list entries.List) (string, error) {
	sorted := list.Sort(entries.SortDate)

	var out bytes.Buffer
	var curr string

	out.WriteString("<h1>Table of Contents</h1>")

	// Here we loop through the entries and print headings with the months.
	for i, entry := range sorted.Slice() {
		month := entry.Date.Format("January") // Using Go's date format syntax.

		// If the current month we're printing entries for has changed, write a new heading.
		if month != curr {
			if i != 0 {
				// We need to make sure to close the last list unless it's the first entry because
				// in that case we don't have anything to close.
				out.WriteString("</ul>")
			}

			out.WriteString("<h2>")
			out.WriteString(month)
			out.WriteString("</h2>")

			curr = month
			out.WriteString("<ul>")
		}

		path := hashString(entry.Path)
		title := fmt.Sprintf("%s: %s", entry.Date.Format("Mon 2006-01-02"), entry.Title)

		// Write something like "<li><a href='aasfhsadkjhf3.xhtml'>Mon 2006-01-02</a></li>"
		out.WriteString("<li><a href='")
		out.WriteString(path)
		out.WriteString("'>")
		out.WriteString(title)
		out.WriteString("</a></li>")
	}

	// Make sure to close the last unordered list.
	out.WriteString("</ul>")

	return out.String(), nil
}

// epubBuildTagSearch creates the XHTML for a tag search page, where all tags are listed along with all
// of the entries with those tags. It's useful for quickly hopping around.
func epubBuildTagSearch(collection *entries.Collection, list entries.List) (string, error) {
	var out bytes.Buffer
	tags := make(map[string]bool)

	for _, entry := range list.Slice() {
		for _, tag := range entry.Tags {
			tags[tag] = true
		}
	}

	out.WriteString("<h1>Tags</h1><ul>")
	for tag := range tags {
		out.WriteString("<li><pre><a href='#")
		out.WriteString(hashString(tag))
		out.WriteString("'>")
		out.WriteString(tag)
		out.WriteString("</a></pre></li>")
	}
	out.WriteString("</ul>")

	for tag := range tags {
		out.WriteString("<h2 id='")
		out.WriteString(hashString(tag))
		out.WriteString("'><pre>")
		out.WriteString(tag)
		out.WriteString("</pre></h2><ul>")

		filtered, err := collection.Filter(entries.FilterTags(tag))
		if err != nil {
			return "", err
		}

		for _, entry := range filtered.List().Sort(entries.SortDate).Slice() {
			path := hashString(entry.Path)
			title := fmt.Sprintf("%s: %s", entry.Date.Format("Mon 2006-01-02"), entry.Title)

			out.WriteString("<li><a href='")
			out.WriteString(path)
			out.WriteString("'>")
			out.WriteString(title)
			out.WriteString("</a></li>")
		}

		out.WriteString("</ul>")
	}

	return out.String(), nil
}

// epubBuildPathSearch creates XHTML for a path search page, a sequential list of all paths sorted alphabetically.
// It's useful for quickly hopping around.
// TODO: have this generate a tree like strucutre, like the `ls` command does.
func epubBuildPathSearch(list entries.List) string {
	sorted := list.Sort(entries.SortPath)
	var out bytes.Buffer

	out.WriteString("<h1>Paths</h1><ul>")

	for _, entry := range sorted.Slice() {
		out.WriteString("<li><a href='")
		out.WriteString(hashString(entry.Path))
		out.WriteString("'><kbd>")
		out.WriteString(entry.Path)
		out.WriteString("</kbd></a></li>")
	}

	return out.String()
}

// epubEntryToXHTML creates the XHTML for an entry, ready to be placed into an EPUB.
// This function returns the XHTML, the title and the path it should be written to, then an error if there
// was one.
func epubEntryToXHTML(md goldmark.Markdown, collection *entries.Collection, entry *entries.Entry) (xhtml string, title string, path string, err error) {
	var buf bytes.Buffer

	err = md.Convert([]byte(entry.Contents), &buf)
	if err != nil {
		return "", "", "", fmt.Errorf("couldn't convert entry %s to markdown: %s", entry.Path, err)
	}

	path = hashString(entry.Path)
	title = fmt.Sprintf("%s: %s", entry.Date.Format("Mon 2006-01-02"), entry.Title)

	metadata, err := yaml.Marshal(entry.Metadata)
	if err != nil {
		log.Errorf("Error marshalling metadta to YAML for entry %s: %s", entry.Path, err)
		metadata = []byte("(error marshalling metadata)")
	}

	entryContents := buf.String()

	for _, link := range entry.OutboundLinks {
		linkedEntry := collection.ResolveLink(link)
		text := entry.Contents[link.Loc[0]:link.Loc[1]]

		if linkedEntry == nil {
			entryContents = strings.ReplaceAll(entryContents, text, "<a href='unknown.xhtml'><kbd>"+text+"</kbd></a>")
		} else {
			location := hashString(linkedEntry.Path)
			entryContents = strings.ReplaceAll(entryContents, text, "<a href='"+location+"'><kbd>"+text+"</kbd></a>")
		}
	}

	contents := fmt.Sprintf("<h1>%s</h1>\n%s\n<hr />", title, entryContents)

	backlinksText := `<h5>Links to this entry</h5><ul>`
	backlinks := collection.FindLinksTo(entry)

	if len(backlinks) != 0 {
		for _, backlink := range backlinks {
			backlinksText += "<li><a href='" + hashString(backlink.Parent.Path) + "'><kbd>" + backlink.Parent.Title + "</kbd></a>"
		}

		contents += "\n" + backlinksText + "</ul><hr />"
	}

	contents += "\n<pre>" + string(metadata) + "</pre>"

	return contents, title, path, nil
}

func init() {
	ActionExportCmd.AddCommand(ActionExportEpubCmd)

	ActionExportEpubCmd.Flags().String("book-author", "", "set the author of the output EPUB, by default the command used to search")
	ActionExportEpubCmd.Flags().String("book-title", "", "set the title of the output EPUB, by default a timestamp")
	ActionExportEpubCmd.Flags().StringP("output", "o", "", "output location of the EPUB, by default the file contents are printed to stdout")
}
