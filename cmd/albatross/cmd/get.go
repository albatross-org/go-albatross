package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get entries matching specific criteria",
	Long: `get finds entries matching specific criteria
	
$ albatross get -path food/pizza # Get the entry about pizza.
$ albatross get -tag "@!journal" -from "2020-01-01 00:00AM" -until "2020-02-01 00:00AM" # Get all journal entries between Jan 2020 and Feb 2020.
$ albatross get -tag "@?gcse" -tag "@?physics" -tag "@?anki" # Get all entries tagged "gcse", "physics" and "anki".
$ albatross get -substring "cat" -sort "alpha" -rev # Sort all entries where you mention cats in reverse alphabetical order.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the misc flags
		dateFormat, err := cmd.Flags().GetString("date-format")
		checkArg(err)

		rev, err := cmd.Flags().GetBool("rev")
		checkArg(err)

		sort, err := cmd.Flags().GetString("sort")
		checkArg(err)

		customEditor, err := cmd.Flags().GetString("editor")
		checkArg(err)

		// Get the filter flags
		fNumber, err := cmd.Flags().GetInt("f-number")
		checkArg(err)

		fFrom, err := cmd.Flags().GetString("f-from")
		checkArg(err)

		fUntil, err := cmd.Flags().GetString("f-until")
		checkArg(err)

		fPath, err := cmd.Flags().GetStringSlice("f-path")
		checkArg(err)

		fTitle, err := cmd.Flags().GetStringSlice("f-title")
		checkArg(err)

		fSubstring, err := cmd.Flags().GetStringSlice("f-substring")
		checkArg(err)

		fTag, err := cmd.Flags().GetStringSlice("f-tag")
		checkArg(err)

		// Get the action flags
		read, err := cmd.Flags().GetBool("read")
		checkArg(err)

		path, err := cmd.Flags().GetBool("path")
		checkArg(err)

		links, err := cmd.Flags().GetBool("links")
		checkArg(err)

		title, err := cmd.Flags().GetBool("title")
		checkArg(err)

		date, err := cmd.Flags().GetBool("date")
		checkArg(err)

		export, err := cmd.Flags().GetBool("export")
		checkArg(err)

		metadata, err := cmd.Flags().GetBool("metadata")
		checkArg(err)

		ankify, err := cmd.Flags().GetBool("ankify")
		checkArg(err)

		update, err := cmd.Flags().GetBool("update")
		checkArg(err)

		attach, err := cmd.Flags().GetBool("attach")
		checkArg(err)

		tags, err := cmd.Flags().GetBool("tags")
		checkArg(err)

		// Check multiple actions weren't given
		var alreadySet bool
		var actions = []bool{read, path, links, title, date, export, metadata, ankify, update, attach, tags}
		for _, action := range actions {
			if action {
				if alreadySet {
					fmt.Println("Multiple actions have been set. Please only set one.")
					os.Exit(1)
				} else {
					alreadySet = true
				}
			}
		}

		if !alreadySet {
			read = true
		}

		// Parse dates using format
		var fFromDate, fUntilDate time.Time

		if fFrom != "" {
			fFromDate, err = time.Parse(dateFormat, fFrom)
			if err != nil {
				log.Fatalf("Can't parse %s using format %s: %s", fFrom, dateFormat, err)
			}
		}

		if fUntil != "" {
			fUntilDate, err = time.Parse(dateFormat, fFrom)
			if err != nil {
				log.Fatalf("Can't parse %s using format %s: %s", fUntil, dateFormat, err)
			}
		}

		// Get the correct editor
		editorName := getEditor("vim")
		if customEditor != "" {
			editorName = customEditor
		}

		// Get the list
		collection, list := get(fFromDate, fUntilDate, fPath, fTitle, fTag, fSubstring)

		switch sort {
		case "alpha":
			list = list.Sort(entries.SortAlpha)
		case "date":
			list = list.Sort(entries.SortDate)
		}

		if rev {
			list = list.Reverse()
		}

		if fNumber != -1 {
			list = list.First(fNumber)
		}

		switch {
		case read:
			for _, entry := range list.Slice() {
				fmt.Println(entry.OriginalContents)
				fmt.Printf("\n\n\n")
			}

		case path:
			for _, entry := range list.Slice() {
				fmt.Println(entry.Path)
			}
		case links:
			for _, entry := range list.Slice() {
				for _, link := range entry.OutboundLinks {
					linkedEntry := collection.ResolveLink(link)
					if linkedEntry != nil {
						fmt.Println(linkedEntry.Path)
					}
				}
			}
		case tags:
			for _, entry := range list.Slice() {
				fmt.Println(strings.Join(entry.Tags, ", "))
			}

		case title:
			for _, entry := range list.Slice() {
				fmt.Println(entry.Title)
			}
		case date:
			for _, entry := range list.Slice() {
				fmt.Println(entry.Date.Format(dateFormat))
			}
		case export:
			j, err := json.Marshal(list.Slice())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(j))
		case metadata:
			metadatas := []map[string]interface{}{}
			for _, entry := range list.Slice() {
				metadatas = append(metadatas, entry.Metadata)
			}

			j, err := json.Marshal(metadatas)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(j))

		case ankify:
			entries := list.Slice()
			generateAnkiFlashcards(entries, true)

		case update:
			var chosen *entries.Entry

			length := len(list.Slice())

			if length == 0 {
				fmt.Println("No entries matched, nothing to update.")
			} else if length != 1 {
				paths := []string{}
				for _, entry := range list.Slice() {
					paths = append(paths, entry.Path)
				}

				fmt.Println("More than one entry matched, please select one.")
				prompt := promptui.Select{
					Label: "Select Entry",
					Items: paths,
				}

				_, result, err := prompt.Run()
				if err != nil {
					log.Fatalf("Couldn't choose entry: %s", err)
				}

				for _, entry := range list.Slice() {
					if entry.Path == result {
						chosen = entry
						break
					}
				}
			} else {
				chosen = list.Slice()[0]
			}

			updateEntry(chosen, editorName)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Filters
	getCmd.Flags().IntP("f-number", "n", -1, "number of entries to return, -1 means all")
	getCmd.Flags().StringP("f-from", "f", "", "only show entries with creation dates after this")
	getCmd.Flags().StringP("f-until", "u", "", "only show entries with creation dates before this")
	getCmd.Flags().StringSliceP("f-path", "p", []string{}, "paths to allow")
	getCmd.Flags().StringSliceP("f-title", "t", []string{}, "titles to allow")
	getCmd.Flags().StringSliceP("f-substring", "s", []string{}, "substrings to allow")
	getCmd.Flags().StringSliceP("f-tag", "a", []string{}, "tags to allow")

	// Actions
	getCmd.Flags().Bool("read", false, "print the entries (default)")
	getCmd.Flags().Bool("path", false, "prints the path to the entries")
	getCmd.Flags().Bool("links", false, "prints all the links in each entry")
	getCmd.Flags().Bool("title", false, "prints the titles in the entries")
	getCmd.Flags().Bool("date", false, "prints the date of the entries")
	getCmd.Flags().Bool("export", false, "prints all the information about the entries, in JSON format")
	getCmd.Flags().Bool("metadata", false, "returns all the metadata in all entries, in JSON format")

	getCmd.Flags().Bool("ankify", false, "output a TSV-formatted file ready for an import to anki (see albatross tool ankify --help)")

	getCmd.Flags().Bool("update", false, "update an entry")
	getCmd.Flags().Bool("attach", false, "attach a file to the entry")
	getCmd.Flags().Bool("tags", false, "prints all the tags")

	// getCmd.Flags().Bool("delete", false, "print the entries")

	// Misc
	getCmd.Flags().String("date-format", "2006-01-02 15:04", "date format (go syntax) for dates")
	getCmd.Flags().BoolP("rev", "r", false, "reverse the list returned")
	getCmd.Flags().String("sort", "", "sorting scheme ('alpha', 'date' or '' for none)")
	getCmd.Flags().StringP("editor", "e", "", "editor to use (defaults to $EDITOR, then vim)")
}

// get gets all the entries matching certain criteria, as an entries.List. It also returns the original entries.Collection
func get(from, until time.Time, paths []string, titles []string, tags []string, substrings []string) (*entries.Collection, entries.List) {
	var err error
	collection, err := store.Collection()

	var originalCollection = collection

	if err != nil {
		log.Fatalf("Error parsing the Albatross store: %s", err)
	}

	if from != (time.Time{}) {
		collection, err = collection.Filter(entries.FilterFrom(from))
		if err != nil {
			log.Fatalf("Error filtering 'from': %s", err)
		}
	}

	if until != (time.Time{}) {
		collection, err = collection.Filter(entries.FilterUntil(until))
		if err != nil {
			log.Fatalf("Error filtering 'until': %s", err)
		}
	}

	if len(paths) != 0 {
		collection, err = collection.Filter(entries.FilterPathsInclude(paths...))
		if err != nil {
			log.Fatalf("Error filtering 'paths': %s", err)
		}
	}

	if len(titles) != 0 {
		collection, err = collection.Filter(entries.FilterTitlesInclude(titles...))
		if err != nil {
			log.Fatalf("Error filtering 'titles': %s", err)
		}
	}

	if len(tags) != 0 {
		collection, err = collection.Filter(entries.FilterTagsInclude(tags...))
		if err != nil {
			log.Fatalf("Error filtering 'tags': %s", err)
		}
	}

	if len(substrings) != 0 {
		collection, err = collection.Filter(entries.FilterMatchInclude(substrings...))
		if err != nil {
			log.Fatalf("Error filtering 'substrings': %s", err)
		}
	}

	return originalCollection, collection.List()
}
