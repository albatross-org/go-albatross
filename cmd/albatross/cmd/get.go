package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/albatross-org/go-albatross/entries"

	"github.com/spf13/cobra"
)

// GetCmd represents the get command
var GetCmd = &cobra.Command{
	Use:     "get <filters> [action]",
	Short:   "Get entries matching specific criteria and perform actions on them",
	Aliases: []string{"search", "query", "g"},
	Long: `get finds entries matching specific criteria and allows you to run actions on them, such as

	- Viewing their contents
	- Updating their contents
	- Printing their links
	- Printing their paths
	- Exporting them as JSON or YAML
	- Showing or creating attachments
	- Generating flashcards
	- Creating a static HTML website to browse entries matched

For a full list, see the subcommands section.

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
the available subcommands.


Advanced: Parsing
-----------------

Parsing is a feature that allows you to run commands on pretend entries. For some context of why this might
be useful, the vim-albatross plugin is a good example. vim-albatross provides a function where you can see
all the Albatross links that are present in the current buffer.

Doing this in vim-albatross in especially difficult because the text inside the current buffer isn't neccesarily
in the store because if the file is unsaved then when an Albatross client reads the filesystem the changes won't be
present. Furthermore, due to how vim-albatross works (as detailed in 'albatross vim --help') even if we write to the
current buffer, the changes won't be reflected.

This could be achieved by manually searching for all the [[Links]] and {{Paths}} that are present in the text,
running a command like 'albatross get title | grep -q "Link Name"' or 'albatross get title | grep -q "path/to/link"'
having Vim check the status code, but then why reimplement an entry parser to perform this when one already exists.

The solution is parsing pretend entries. Now we can run a command like:

	# Print all valid links in the entry:
	$ albatross get --parse-content "Is [[this]] a link? This is definitely a link: [[Pizza]]" links
	food/pizza

	# Print links that don't exist:
	$ albatross get --parse-content "Is [[this]] a link? This is definitely a link: [[Pizza]]" links -e
	[[this]]

In the case of vim-albatross specifically:

	$ albatross get --parse-file /tmp/vim-albatross/... links

This may seem overkill, and it probably is. But the technique is much more general and can be used for other purposes:

	# Generate flashcards that aren't in an existing entry:
	$ albatross get --parse-file flashcards.txt ankify > ready_to_import_to_anki.tsv

So in summary: this feature can be used to add a pretend entry to the store which can be queried and interacted with
just like any other entry.

	FLAG					FUNCTION
	--parse-content			Parse the text in this flag like it's an entry.
	--parse-file			Parse the file in this flag like it's an entry.

Only one of these flags can be used at a time.

By default, if you invoke a --parse flag, the search will automatically be filtered to only include that entry. For example:

	# What you might expect to happen:
	$ albatross get -p school --parse-content "I'm a pretend entry!" title
	I'm a pretend entry! 			(here the entry content is the same as the title seeing as there's no metadata)
	Computing - Communications
	Computing - Networking
	Further Maths - Argand Diagrams
	...


	# What actually happens:
	$ albatross get -p school --parse-content "I'm a pretend entry!" title
	I'm a pretend entry!

You can disable this using the --parse-dont-restrict flag. Pretend entries will have a path like "_parsing/I53Buamaskx5GZv7"
in order to prevent conflicts with other paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		for _, entry := range list.Slice() {
			fmt.Println(entry.Path)
		}
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

	GetCmd.PersistentFlags().StringSliceP("path", "p", []string{}, "paths to allow, prefix")
	GetCmd.PersistentFlags().StringSliceP("title", "t", []string{}, "titles to allow, substring")
	GetCmd.PersistentFlags().StringSliceP("contents", "c", []string{}, "contents to allow, substring")

	GetCmd.PersistentFlags().StringSlice("path-exact", []string{}, "paths to allow, exact")
	GetCmd.PersistentFlags().StringSlice("title-exact", []string{}, "titles to allow, exact")
	GetCmd.PersistentFlags().StringSlice("contents-exact", []string{}, "substrings to allow, exact")

	GetCmd.PersistentFlags().StringSlice("path-not", []string{}, "paths to disallow, prefix")
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

	// Parsing
	GetCmd.PersistentFlags().String("parse-content", "", "parse the text given as an entry that was matched as part of the search. See help for details")
	GetCmd.PersistentFlags().String("parse-file", "", "parse this file like an entry that was matched as part of the search. See help for details")
	GetCmd.PersistentFlags().Bool("parse-dont-restrict", false, "don't automatically restrict the entries matched to be parsed entries if any were given. See help for details")
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
	checkArgVerbose(cmd, "date-format", err)

	rev, err := cmd.Flags().GetBool("rev")
	checkArgVerbose(cmd, "rev", err)

	sort, err := cmd.Flags().GetString("sort")
	checkArgVerbose(cmd, "sort", err)

	delimeter, err := cmd.Flags().GetString("delimeter")
	checkArgVerbose(cmd, "delimeter", err)

	// Get the filter flags, generic
	number, err := cmd.Flags().GetInt("number")
	checkArgVerbose(cmd, "number", err)

	from, err := cmd.Flags().GetString("from")
	checkArgVerbose(cmd, "from", err)

	until, err := cmd.Flags().GetString("until")
	checkArgVerbose(cmd, "until", err)

	minLength, err := cmd.Flags().GetInt("min-length")
	checkArgVerbose(cmd, "min-length", err)

	maxLength, err := cmd.Flags().GetInt("max-length")
	checkArgVerbose(cmd, "max-length", err)

	tags, err := cmd.Flags().GetStringSlice("tag")
	checkArgVerbose(cmd, "tag", err)

	tagsExclude, err := cmd.Flags().GetStringSlice("tag-not")
	checkArgVerbose(cmd, "tag-not", err)

	// Get the filter flags, match vs not
	pathsMatch, err := cmd.Flags().GetStringSlice("path")
	checkArgVerbose(cmd, "path", err)

	pathsExact, err := cmd.Flags().GetStringSlice("path-exact")
	checkArgVerbose(cmd, "path-exact", err)

	pathsMatchNot, err := cmd.Flags().GetStringSlice("path-not")
	checkArgVerbose(cmd, "path-not", err)

	pathsExactNot, err := cmd.Flags().GetStringSlice("path-exact-not")
	checkArgVerbose(cmd, "path-exact-not", err)

	titlesMatch, err := cmd.Flags().GetStringSlice("title")
	checkArgVerbose(cmd, "title", err)

	titlesExact, err := cmd.Flags().GetStringSlice("title-exact")
	checkArgVerbose(cmd, "title-exact", err)

	titlesMatchNot, err := cmd.Flags().GetStringSlice("title-not")
	checkArgVerbose(cmd, "title-not", err)

	titlesExactNot, err := cmd.Flags().GetStringSlice("title-exact-not")
	checkArgVerbose(cmd, "title-exact-not", err)

	contentsMatch, err := cmd.Flags().GetStringSlice("contents")
	checkArgVerbose(cmd, "contents", err)

	contentsExact, err := cmd.Flags().GetStringSlice("contents-exact")
	checkArgVerbose(cmd, "contents-exact", err)

	contentsMatchNot, err := cmd.Flags().GetStringSlice("contents-not")
	checkArgVerbose(cmd, "contents-not", err)

	contentsExactNot, err := cmd.Flags().GetStringSlice("contents-exact-not")
	checkArgVerbose(cmd, "contents-exact-not", err)

	stdin, err := cmd.Flags().GetBool("stdin")
	checkArgVerbose(cmd, "stdin", err)

	// Get flags for parsing fake entries
	parseContent, err := cmd.Flags().GetString("parse-content")
	checkArgVerbose(cmd, "parse-content", err)

	parseFile, err := cmd.Flags().GetString("parse-file")
	checkArgVerbose(cmd, "parse-file", err)

	parseDontRestrict, err := cmd.Flags().GetBool("parse-dont-restrict")
	checkArgVerbose(cmd, "parse-dont-restrict", err)

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

	if parseContent != "" || parseFile != "" {
		if parseContent != "" && parseFile != "" {
			fmt.Println("Error: Please use either --parse-content or --parse-file, not both at the same time.")
			fmt.Println(err)
			os.Exit(1)
		}

		parser, err := entries.NewParser(store.Config.DateFormat, store.Config.TagPrefix)
		if err != nil {
			fmt.Println("Error: Couldn't create a new Parser struct. Something's gone wrong.")
			fmt.Println(err)
			os.Exit(1)
		}

		var pretendEntry *entries.Entry
		fakePath := fmt.Sprint("_parsing", "/", randomString(16))

		if parseContent != "" {
			pretendEntry, err = parser.Parse(fakePath, parseContent)
		} else if parseFile != "" {
			pretendContentBytes, err := ioutil.ReadFile(parseFile)
			if err != nil {
				fmt.Println("Error: Couldn't read the --parse-file ", parseFile)
				fmt.Println(err)
				os.Exit(1)
			}
			pretendEntry, err = parser.Parse(fakePath, string(pretendContentBytes))
		}

		if err != nil {
			fmt.Println("Error: Couldn't parse the entry given:")
			fmt.Println(err)
			os.Exit(1)
		}

		pretendEntry.Path = fakePath
		pretendEntry.Synthetic = true

		collection.Add(pretendEntry)

		if !parseDontRestrict {
			query.PathsExact = [][]string{[]string{fakePath}}
		}
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
