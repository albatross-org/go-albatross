package cmd

import (
	"time"

	"github.com/albatross-org/go-albatross/entries"

	"github.com/spf13/cobra"
)

// These are global variables that is set once the get command is called.
// This is to allow commands in the `actions/` subdirectory to access filtered entries.
var (
	GetCollection *entries.Collection
	GetList       entries.List
)

// GetCmd represents the get command
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "get entries matching specific criteria",
	Long: `get finds entries matching specific criteria
	
Some examples:

	# Get the entry about pizza.
	$ albatross get --path food/pizza
	
	# Get all journal entries between Jan 2020 and Feb 2020.
	$ albatross get --tag "@!journal" --from "2020-01-01 00:00AM" --until "2020-02-01 00:00AM"
	
	# Get all entries tagged "gcse", "physics" and "ankify".
	$ albatross get --tag "@?gcse" --tag "@?physics" --tag "@?ankify"
	
	# Sort all entries where you mention cats in reverse alphabetical order.
	$ albatross get --substring "cat" --sort "alpha" --rev
	
By default, the command will print all the entries to all the paths that it matched. However, you can do
much more. 'Actions' are mini-programs that operate on lists of entries. For all available entries, see
the available subcommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		ActionPathCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(GetCmd)

	GetCmd.PersistentFlags().String("date-format", "2006-01-02", "date format for parsing from and until")

	// Filters
	GetCmd.PersistentFlags().IntP("number", "n", -1, "number of entries to return, -1 means all")
	GetCmd.PersistentFlags().StringP("from", "f", "", "only show entries with creation dates after this")
	GetCmd.PersistentFlags().StringP("until", "u", "", "only show entries with creation dates before this")
	GetCmd.PersistentFlags().StringSliceP("path", "p", []string{}, "paths to allow")
	GetCmd.PersistentFlags().StringSliceP("title", "t", []string{}, "titles to allow")
	GetCmd.PersistentFlags().StringSliceP("substring", "s", []string{}, "substrings to allow")
	GetCmd.PersistentFlags().StringSliceP("tag", "a", []string{}, "tags to allow")

	// Misc
	GetCmd.PersistentFlags().BoolP("rev", "r", false, "reverse the list returned")
	GetCmd.PersistentFlags().String("sort", "", "sorting scheme ('alpha', 'date' or '' for none)")
}

// getFromCommand runs a get query by parsing a command for flags.
func getFromCommand(cmd *cobra.Command) (*entries.Collection, entries.List) {
	// Get the misc flags
	dateFormat, err := cmd.Flags().GetString("date-format")
	checkArg(err)

	rev, err := cmd.Flags().GetBool("rev")
	checkArg(err)

	sort, err := cmd.Flags().GetString("sort")
	checkArg(err)

	// Get the filter flags
	fNumber, err := cmd.Flags().GetInt("number")
	checkArg(err)

	fFrom, err := cmd.Flags().GetString("from")
	checkArg(err)

	fUntil, err := cmd.Flags().GetString("until")
	checkArg(err)

	fPath, err := cmd.Flags().GetStringSlice("path")
	checkArg(err)

	fTitle, err := cmd.Flags().GetStringSlice("title")
	checkArg(err)

	fSubstring, err := cmd.Flags().GetStringSlice("substring")
	checkArg(err)

	fTag, err := cmd.Flags().GetStringSlice("tag")
	checkArg(err)

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

	return collection, list
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
