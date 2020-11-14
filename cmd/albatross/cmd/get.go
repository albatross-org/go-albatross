package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

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
	Use:     "get <filters> [action]",
	Short:   "get entries matching specific criteria and perform actions on them",
	Aliases: []string{"search", "query", "g"},
	Long: `get finds entries matching specific criteria and allows you to run actions on them, such as

	- Printing their links
	- Printing their paths
	- Exporting them as JSON or YAML
	- Generating flashcards

Some examples:

	# Get the entry about pizza.
	$ albatross get --path food/pizza

	# Get all journal entries between Jan 2020 and Feb 2020.
	$ albatross get --tag "@!journal" --from "2020-01-01 00:00AM" --until "2020-02-01 00:00AM"

	# Get all entries tagged "gcse", "physics" and "ankify".
	$ albatross get --tag "@?gcse" --tag "@?physics" --tag "@?ankify"

	# Sort all entries where you mention cats in reverse alphabetical order.
	$ albatross get --substring "cat" --sort "alpha" --rev

The syntax of a get command is:

	albatross get --<filters> [action]

Filters are flags which allow or disallow entries. For example:

	--tag "@?food"

Will only allow entries containing the tag "@?food". Furthermore,

	--tag "@?food" --path "notes"

Will only allow entries containing the tag "@?food" AND that are from the "notes/" path. So each subsequent
filter further restricts the amount of entries that can be matched. However, this leads to a difficulty: what
if you wish to specify multiple filters which will allow entries matching either criteria? In other words, what if
you want OR instead of AND?

In order to achieve this, some flags allow syntax like this:

	--path "notes OR school"

This will match entries from the path "notes/" or "school/". The filters that support this feature are:

	--path     --path-exact     --path-not     --path-exact-not
	--title    --title-exact    --title-not    --title-exact-not
	--contents --contents-exact --contents-not --contents-exact-not

You can also change the delimeter used from " OR " using the --delimeter flag.

By default, the command will print all the entries to all the paths that it matched. However, you can do
much more. 'Actions' are mini-programs that operate on lists of entries. For all available entries, see
the available subcommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		ActionPathCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(GetCmd)

	// Filters
	GetCmd.PersistentFlags().IntP("number", "n", -1, "number of entries to return, -1 means all")
	GetCmd.PersistentFlags().StringP("from", "f", "", "only show entries with creation dates after this")
	GetCmd.PersistentFlags().StringP("until", "u", "", "only show entries with creation dates before this")

	GetCmd.PersistentFlags().Int("min-length", 0, "minimum length to allow")
	GetCmd.PersistentFlags().Int("max-length", 0, "maximum length to allow")

	GetCmd.PersistentFlags().StringSliceP("tag", "a", []string{}, "tags to allow")
	GetCmd.PersistentFlags().StringSlice("tag-not", []string{}, "tags to disallow")

	GetCmd.PersistentFlags().StringSliceP("path", "p", []string{}, "paths to allow, substring")
	GetCmd.PersistentFlags().StringSliceP("title", "t", []string{}, "titles to allow, substring")
	GetCmd.PersistentFlags().StringSliceP("contents", "c", []string{}, "contents to allow, substring")

	GetCmd.PersistentFlags().StringSlice("path-exact", []string{}, "paths to allow, exact")
	GetCmd.PersistentFlags().StringSlice("title-exact", []string{}, "titles to allow, exact")
	GetCmd.PersistentFlags().StringSlice("contents-exact", []string{}, "substrings to allow, exact")

	GetCmd.PersistentFlags().StringSlice("path-not", []string{}, "paths to disallow, substring")
	GetCmd.PersistentFlags().StringSlice("title-not", []string{}, "titles to disallow, substring")
	GetCmd.PersistentFlags().StringSlice("contents-not", []string{}, "contents to disallow, substring")

	GetCmd.PersistentFlags().StringSlice("path-exact-not", []string{}, "paths to disallow, exact")
	GetCmd.PersistentFlags().StringSlice("title-exact-not", []string{}, "titles to disallow, exact")
	GetCmd.PersistentFlags().StringSlice("contents-exact-not", []string{}, "substrings to disallow, exact")

	GetCmd.PersistentFlags().BoolP("stdin", "i", false, "read list of exact paths from stdin")

	// Misc
	GetCmd.PersistentFlags().BoolP("rev", "r", false, "reverse the list returned")
	GetCmd.PersistentFlags().String("sort", "", "sorting scheme ('alpha', 'date' or '' for random)")
	GetCmd.PersistentFlags().String("date-format", "2006-01-02 15:04", "date format for parsing from and until")
	GetCmd.PersistentFlags().String("delimeter", " OR ", "delimeter to use for splitting up arguments")
}

// multiSplit is like strings.Split except it splits a slice of strings into a slice of slices.
func multiSplit(strs []string, delimeter string) [][]string {
	res := [][]string{}

	for _, str := range strs {
		res = append(res, strings.Split(str, delimeter))
	}

	return res
}

// getFromCommand runs a get query by parsing a command for flags.
func getFromCommand(cmd *cobra.Command) (collection *entries.Collection, filtered *entries.Collection, list entries.List) {
	encrypted, err := store.Encrypted()
	if err != nil {
		log.Fatal(err)
	} else if encrypted {
		decryptStore()

		if !leaveDecrypted {
			defer encryptStore()
		}
	}

	// Get the misc flags
	dateFormat, err := cmd.Flags().GetString("date-format")
	checkArg(err)

	rev, err := cmd.Flags().GetBool("rev")
	checkArg(err)

	sort, err := cmd.Flags().GetString("sort")
	checkArg(err)

	delimeter, err := cmd.Flags().GetString("delimeter")
	checkArg(err)

	// Get the filter flags, generic
	number, err := cmd.Flags().GetInt("number")
	checkArg(err)

	from, err := cmd.Flags().GetString("from")
	checkArg(err)

	until, err := cmd.Flags().GetString("until")
	checkArg(err)

	minLength, err := cmd.Flags().GetInt("min-length")
	checkArg(err)

	maxLength, err := cmd.Flags().GetInt("max-length")
	checkArg(err)

	tags, err := cmd.Flags().GetStringSlice("tag")
	checkArg(err)

	tagsExclude, err := cmd.Flags().GetStringSlice("tag-not")
	checkArg(err)

	// Get the filter flags, match vs not
	pathsMatch, err := cmd.Flags().GetStringSlice("path")
	checkArg(err)

	pathsExact, err := cmd.Flags().GetStringSlice("path-exact")
	checkArg(err)

	pathsMatchNot, err := cmd.Flags().GetStringSlice("path-not")
	checkArg(err)

	pathsExactNot, err := cmd.Flags().GetStringSlice("path-exact-not")
	checkArg(err)

	titlesMatch, err := cmd.Flags().GetStringSlice("title")
	checkArg(err)

	titlesExact, err := cmd.Flags().GetStringSlice("title-exact")
	checkArg(err)

	titlesMatchNot, err := cmd.Flags().GetStringSlice("title-not")
	checkArg(err)

	titlesExactNot, err := cmd.Flags().GetStringSlice("title-exact-not")
	checkArg(err)

	contentsMatch, err := cmd.Flags().GetStringSlice("contents")
	checkArg(err)

	contentsExact, err := cmd.Flags().GetStringSlice("contents-exact")
	checkArg(err)

	contentsMatchNot, err := cmd.Flags().GetStringSlice("contents-not")
	checkArg(err)

	contentsExactNot, err := cmd.Flags().GetStringSlice("contents-exact-not")
	checkArg(err)

	stdin, err := cmd.Flags().GetBool("stdin")
	checkArg(err)

	// Parse dates using format
	var fromDate, untilDate time.Time

	if from != "" {
		fromDate, err = time.Parse(dateFormat, from)
		if err != nil {
			log.Fatalf("Can't parse %s using format %s: %s", from, dateFormat, err)
		}
	}

	if until != "" {
		untilDate, err = time.Parse(dateFormat, until)
		if err != nil {
			log.Fatalf("Can't parse %s using format %s: %s", until, dateFormat, err)
		}
	}

	// Build the query
	query := entries.Query{
		From:  fromDate,
		Until: untilDate,

		MinLength: minLength,
		MaxLength: maxLength,

		Tags:        tags,
		TagsExclude: tagsExclude,

		ContentsExact:        multiSplit(contentsExact, delimeter),
		ContentsMatch:        multiSplit(contentsMatch, delimeter),
		ContentsExactExclude: multiSplit(contentsExactNot, delimeter),
		ContentsMatchExclude: multiSplit(contentsMatchNot, delimeter),

		PathsExact:        multiSplit(pathsExact, delimeter),
		PathsMatch:        multiSplit(pathsMatch, delimeter),
		PathsExactExclude: multiSplit(pathsExactNot, delimeter),
		PathsMatchExclude: multiSplit(pathsMatchNot, delimeter),

		TitlesExact:        multiSplit(titlesExact, delimeter),
		TitlesMatch:        multiSplit(titlesMatch, delimeter),
		TitlesExactExclude: multiSplit(titlesExactNot, delimeter),
		TitlesMatchExclude: multiSplit(titlesMatchNot, delimeter),
	}

	// Get stdin paths
	if stdin {
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Can't read stdin: %s", err)
		}

		query.PathsExact = append(query.PathsExact, strings.Split(string(stdin), "\n"))
	}

	if globalLog.IsLevelEnabled(logrus.TraceLevel) {
		queryJSON, err := json.MarshalIndent(query, "", "\t")
		if err != nil {
			log.Errorf("couldn't marshal query as JSON for tracing: %s", err)
		}

		log.Tracef("Query created from command: %s", string(queryJSON))
	}

	collection, err = store.Collection()
	if err != nil {
		log.Fatalf("Couldn't parse Albatross store to collection: %s", err)
	}

	start := time.Now()

	filtered, err = collection.Filter(query.Filter())
	if err != nil {
		log.Fatalf("Couldn't run filter on Albatross store: %s", err)
	}

	end := time.Now()

	list = filtered.List()

	switch sort {
	case "alpha":
		list = list.Sort(entries.SortAlpha)
	case "date":
		list = list.Sort(entries.SortDate)
	}

	if rev {
		list = list.Reverse()
	}

	if number != -1 {
		list = list.First(number)
	}

	if globalLog.IsLevelEnabled(logrus.DebugLevel) {
		log.Debugf("Query matched %d entries in %s.", len(list.Slice()), end.Sub(start))
	}

	return collection, filtered, list
}
