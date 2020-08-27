package cmd

import (
	"bytes"
	"encoding/csv"
	"os"
	"regexp"
	"strings"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var regexLatexSingle = regexp.MustCompile(`[^$](\${1})[^$]`)
var regexLatexDouble = regexp.MustCompile(`(\${2})`)

// toolAnkifyCmd represents the ankify command
var toolAnkifyCmd = &cobra.Command{
	Use:   "ankify",
	Short: "ankify converts notes into Anki flashcards",
	Args:  cobra.ExactArgs(1),
	Long: `ankify converts entries into anki flashcards.
	
Given a path, it will find all entries tagged @?ankify and convert headings with two
question marks (??) into flashcards. For example:

    $ albatross tool ankify path/to/entries

    ##### What's the point in living??
    The point in living is to...

Will become:

    1st side:
    ┌──────────────────────────────┐
    │ What's the point in living?  │
    └──────────────────────────────┘

    2nd side:
    ┌──────────────────────────────┐
    │ The point in living is to... │
	└──────────────────────────────┘

You can also use ankify with the 'get' command for more powerful searches:

    $ albatross get --title "My Flashcards" --ankify 
	
It outputs a TSV file, which can then be redirected into a file:

	$ albatross tool ankify path/to/entries > ~/.local/decks/entries.tsv

The format of the TSV file is:

<HEADING>	<QUESTION>	<PATH>

In order to import this into Anki, open the application and click "Import File" at the bottom.
You will need to create a new Note Type so that Anki handles the path correctly before you import. To do this:

- Click 'Add' at the top.
- Press 'Manage', then Add.
- Click 'Add: Basic', and enter a suitable name like 'Ankify'.
- This should open the note entry window.
- Then click 'Fields...', press 'Add' and name it 'Path'.

That should be it. Now when you import the TSV file, select the Note Type as being 'Ankify', or the name that you entered.

As a suggestion, decks should cover broad topics as a whole. So instead of creating "School::A-Level::Physics::Topic 1::Electromagnetism",
it's better to have a deck that's more like "School::A-Level::Physics". If you need to study a specific section of an Albatross store,
you can leverage the search field and create a filtered deck (Tools->Create Filtered Deck):

	# Revise a single topic
	path:*school/a-level/physics/topic1*
	
	# Revise a specific piece of knowledge
	path:*school/a-level/physics/topic8/electromagnetism
`,
	Run: func(cmd *cobra.Command, args []string) {
		collection, err := store.Collection()
		if err != nil {
			log.Fatalf("Error parsing the Albatross store: %s", err)
		}

		path := args[0]

		ignoreTagRequirement, err := cmd.Flags().GetBool("ignore-required-tag")
		if err != nil {
			log.Fatalf("Error getting 'ignore-required-tag' flag: %s", err)
		}

		fixLatex, err := cmd.Flags().GetBool("fix-latex")
		if err != nil {
			log.Fatalf("Error getting 'fix-latex' flag: %s", err)
		}

		filtered, err := collection.Filter(entries.FilterPathsInclude(path))
		if err != nil {
			log.Fatalf("Error filtering entries for exact path %q: %s", path, err)
		}

		if !ignoreTagRequirement {
			filtered, err = collection.Filter(entries.FilterTagsInclude("@?ankify"))
			if err != nil {
				log.Fatalf("Error filtering entries for tag %q: %s", "@?ankify", err)
			}
		}

		entries := filtered.List().Slice()
		generateAnkiFlashcards(entries, fixLatex)
	},
}

func generateAnkiFlashcards(entries []*entries.Entry, fixLatex bool) {
	csvw := csv.NewWriter(os.Stdout)
	csvw.Comma = '\t'

	for _, entry := range entries {
		flashcards, err := extractFlashcards(entry)
		if err != nil {
			log.Error("Error parsing markdown for entry %q: %s", entry.Path, err)
			continue
		}

		for _, flashcard := range flashcards {
			row := []string{flashcard[0], strings.Join(flashcard[1:], ""), entry.Path}
			if fixLatex {
				row = fixFlashcardLatex(row)
			}

			csvw.Write(row)
		}
	}

	csvw.Flush()
}

func extractFlashcards(entry *entries.Entry) ([][]string, error) {
	md := goldmark.New()
	parser := md.Parser()
	renderer := md.Renderer()

	contents := []byte(entry.Contents)
	flashcards := [][]string{}

	// var buf bytes.Buffer
	rootAst := parser.Parse(text.NewReader(contents))
	child := rootAst.FirstChild()

	state := "none"
	flashcard := []string{}

	for child != nil {
		if child.Kind() == ast.KindHeading {
			heading := child.(*ast.Heading)
			text := heading.Text(contents)

			if len(flashcard) != 0 {
				flashcards = append(flashcards, flashcard)
			}

			if strings.HasSuffix(string(text), "??") {
				state = "flashcard"
			} else {
				state = "none"
				child = child.NextSibling()
				continue
			}

			flashcard = []string{strings.ReplaceAll(string(text), "\n", "")}
			child = child.NextSibling()

		} else if state == "flashcard" {
			var buf bytes.Buffer
			renderer.Render(&buf, contents, child)
			flashcard = append(flashcard, strings.ReplaceAll(buf.String(), "\n", ""))
			child = child.NextSibling()

		} else {
			child = child.NextSibling()
		}

	}

	return flashcards, nil
}

// fixFlashcardLatex replaces '$' and '$$' with '[$]' and '[$$]' respectively.
// It does this in a very hacky way.
func fixFlashcardLatex(flashcard []string) []string {
	for i := range flashcard {
		flashcard[i] = strings.ReplaceAll(flashcard[i], "$$", "[@@]")
		flashcard[i] = strings.ReplaceAll(flashcard[i], "$", "[$]")
		flashcard[i] = strings.ReplaceAll(flashcard[i], "[@@]", "[$$]")
	}

	return flashcard
}

func init() {
	toolCmd.AddCommand(toolAnkifyCmd)

	toolAnkifyCmd.Flags().BoolP("ignore-required-tag", "i", false, "Don't require the @?ankify tag to generate flashcards for that entry")
	toolAnkifyCmd.Flags().Bool("fix-latex", true, "converts '$' and '$$' to '[$]' and '[$$]'")
}
