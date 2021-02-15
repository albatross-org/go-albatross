package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	albatross "github.com/albatross-org/go-albatross/albatross"

	"github.com/spf13/cobra"
)

var defaultEntry = `---
title: "<(default "Title" .title)>"
date: "<(.date | date "2006-01-02 15:04")>"
---

`

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new entry",
	Aliases: []string{"new"},
	Long: `This command lets you create a new entry from the command line
	
	$ albatross create food/pizza "The Most Amazing Pizza"

You can define templates which are placed in a "templates/" directory at the root of the store:

	.
	├── config.yaml
	├── entries/
	└── templates/
		└── exercise.tmpl

In the above example, the template could be used when creating the entry like so:

	$ albatross create logs/exercise/2020/08/30 -t exercise

You can use Go's template strings within the template files themselves to auto populate some values
to save typing:

	(exercise.tmpl)
	---
	title: 'Exercise Log'
	date: '<(.date | date "2006-01-02 15:04")>'
	---

	### [[Running]]
	* <(.distance)> mi @ <(.pace)>/mi

Notice the alternate syntax for templates, "<(" and ")>", opposed to Go's default "{{" and "}}". This is
to prevent interference with Albatross' path links.

As a context for the template, you can pass key values with the -c flag:

	$ albatross create logs/exercise/2020/08/30 -t exercise -c distance=3.24 -c pace=7:47

.date, as shown above, is set automatically to the current time. Sprig (https://github.com/Masterminds/sprig) helper
functions/pipelines are available, such as:

	- date
	- toJSON
	- upper

The default template is:

	---
	title: "<(default "Title" .title)>"
	date: "<(.date | date "2006-01-02 15:04")>"
	---

The --from-file option allows you to create an entry from a file. You are prompted to edit the file first, though this can
be disabled using the --no-edit flag.

	(file.md)
	---
	title: "My Awesome Copied File"
	date: "2020-12-23 13:57"
	---

	(shell)
	$ albatross new junk/copied-file --from-file file.md

The above will create an entry at the path junk/copied-file that has the exact contents of file.md.

	`,
	Run: func(cmd *cobra.Command, args []string) {
		encrypted, err := store.Encrypted()
		if err != nil {
			log.Fatal(err)
		} else if encrypted {
			decryptStore()

			if !leaveDecrypted {
				defer encryptStore()
			}
		}

		editorName := getEditor("vim")
		customEditor, err := cmd.Flags().GetString("editor")
		checkArg(err)

		if customEditor != "" {
			editorName = customEditor
		}

		noEdit, err := cmd.Flags().GetBool("no-edit")
		checkArg(err)

		templateFile, err := cmd.Flags().GetString("template")
		checkArg(err)

		contextStrings, err := cmd.Flags().GetStringToString("context")
		checkArg(err)

		fromFile, err := cmd.Flags().GetString("from-file")
		checkArg(err)

		if len(args) == 0 {
			fmt.Println("Expecting exactly one or more arguments: path to entry and optional title")
			fmt.Println("For example:")
			fmt.Println("")
			fmt.Println("$ albatross create food/pizza Pizza")
		}

		contextStrings["title"] = strings.Join(args[1:], " ")

		var contents, defaultContents string
		if fromFile != "" {
			bytes, err := ioutil.ReadFile(fromFile)
			if err != nil {
				fmt.Printf("Error creating entry from file: %s\n", err)
				os.Exit(1)
			}

			contents = string(bytes)
		} else {
			contents, defaultContents = getTemplate(templateFile, contextStrings)
		}

		// Here we create an empty entry first, then update it.
		// This means that an error like "EntryAlreadyExists" will come up now rather than
		// after the entry has been created, which could lead to data loss and be frustrating in general.
		err = store.Create(args[0], contents)
		if err != nil {
			if _, ok := err.(albatross.ErrEntryAlreadyExists); ok {
				fmt.Printf("Entry %s already exists.\n", args[0])
				os.Exit(1)
			}

			log.Fatal("Couldn't create entry: ", err)
		}

		var editedContent string

		if !noEdit {
			editedContent, err = edit(editorName, contents)
			if err != nil {
				log.Fatal("Couldn't get content from editor: ", err)
			}
		} else {
			editedContent = contents
		}

		// The user didn't actually make any changes from the default value of the template. This means that
		// we shouldn't actually create the entry.
		if editedContent == defaultContents && fromFile == "" {
			err = store.Delete(args[0])
			if err != nil {
				log.Fatal("Couldn't delete blank entry: ", err)
			}

			fmt.Printf("Entry %s left blank, not creating.\n", args[0])
			os.Exit(0)
		}

		if !noEdit {
			err = store.Update(args[0], editedContent)
			if err != nil {
				f, err := ioutil.TempFile("", "albatross-recover")
				if err != nil {
					log.Fatal("Couldn't get create temporary file to save recovery entry to. You're on your own! ", err)
				}

				_, err = f.Write([]byte(editedContent))
				if err != nil {
					log.Fatal("Error writing to temporary file to save recovery entry to. You're on your own! ", err)
				}

				fmt.Println("Error creating entry. A copy has been saved to:", f.Name())
				os.Exit(1)
			}
		}

		fmt.Println("Successfully created entry", args[0])
	},
}

// getTemplate takes a template name and a map containing values to populate the given template with. It returns two things,
// the output of executing the template (i.e. what the user wants) and also the output of executing the template with no
// contextStrings. This is done so that the empty version can be compared later to see if the user actually wrote anything
// and therefore whether it's needed to actually create an entry.
func getTemplate(name string, contextStrings map[string]string) (string, string) {
	var context = make(map[string]interface{})
	for k, v := range contextStrings {
		context[k] = v
	}

	context["date"] = time.Now()

	templates, err := ioutil.ReadDir(filepath.Join(store.Path, "templates"))
	if err != nil && name != "" {
		log.Fatalf("Error reading templates directory: %s", err)
		return "", ""
	}

	var match string

	if name != "" {
		for _, info := range templates {
			templateName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			if templateName == name {
				matchBytes, err := ioutil.ReadFile(filepath.Join(store.Path, "templates", info.Name()))
				if err != nil {
					log.Fatalf("error reading template file %s: %s", filepath.Join(store.Path, "templates", info.Name()), err)
				}

				match = string(matchBytes)
			}
		}

		if len(match) == 0 {
			log.Fatalf("Template '%s' doesn't exist.", name)
		}
	} else {
		match = defaultEntry
	}

	tmpl := template.New("template").Delims("<(", ")>").Funcs(sprig.TxtFuncMap())
	tmpl, err = tmpl.Parse(match)
	if err != nil {
		log.Fatalf("Error parsing template: %s", err)
	}

	var out bytes.Buffer
	var outDefault bytes.Buffer

	err = tmpl.Execute(&out, context)
	if err != nil {
		log.Fatalf("Error executing template: %s", err)
	}

	err = tmpl.Execute(&outDefault, map[string]string{})
	if err != nil {
		log.Fatalf("Error executing template: %s", err)
	}

	return out.String(), outDefault.String()
}

func init() {
	rootCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().StringP("editor", "e", "", "Editor to use (defaults to $EDITOR, then vim)")
	CreateCmd.Flags().Bool("no-edit", false, "don't edit the entry, create it straight away")
	CreateCmd.Flags().StringP("template", "t", "", "Template file to use")
	CreateCmd.Flags().StringP("from-file", "f", "", "If set to a file, create the entry from the file's contents")
	CreateCmd.Flags().StringToStringP("context", "c", map[string]string{}, "Context for template")
}
