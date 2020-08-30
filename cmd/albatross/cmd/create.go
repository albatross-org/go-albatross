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
	"github.com/sirupsen/logrus"

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
	Short:   "create a new entry",
	Aliases: []string{"new"},
	Long: `create a new entry from the command line
	
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
	* <(.distance)> mi @ !(.pace)!/mi

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

		templateFile, err := cmd.Flags().GetString("template")
		checkArg(err)

		contextStrings, err := cmd.Flags().GetStringToString("context")
		checkArg(err)

		if len(args) == 0 {
			fmt.Println("Expecting exactly one or more arguments: path to entry and optional title")
			fmt.Println("For example:")
			fmt.Println("")
			fmt.Println("$ albatross create food/pizza Pizza")
		}

		contextStrings["title"] = strings.Join(args[1:], " ")

		// Here we create an empty entry first, then update it.
		// This means that an error like "EntryAlreadyExists" will come up now rather than
		// after the entry has been created, which could lead to data loss.
		err = store.Create(args[0], fmt.Sprintf(defaultEntry, "Temp", time.Now().Format("2006-01-02 15:04")))
		if err != nil {
			log.Fatal("Couldn't create entry: ", err)
		}

		content, err := edit(editorName, getTemplate(templateFile, contextStrings))
		if err != nil {
			log.Fatal("Couldn't get content from editor: ", err)
		}

		err = store.Update(args[0], content)
		if err != nil {
			f, err := ioutil.TempFile("", "albatross-recover")
			if err != nil {
				logrus.Fatal("Couldn't get create temporary file to save recovery entry to. You're on your own! ", err)
			}

			f.Write([]byte(content))

			fmt.Println("Error creating entry. A copy has been saved to:", f.Name())
			os.Exit(1)
		}

		fmt.Println("Successfully created entry", args[0])
	},
}

func getTemplate(name string, contextStrings map[string]string) string {
	var context = make(map[string]interface{})
	for k, v := range contextStrings {
		context[k] = v
	}

	context["date"] = time.Now()

	templates, err := ioutil.ReadDir(filepath.Join(storePath, "templates"))
	if err != nil {
		logrus.Fatalf("error reading templates directory: %s", err)
		return ""
	}

	var match string

	if name != "" {
		for _, info := range templates {
			templateName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			if templateName == name {
				matchBytes, err := ioutil.ReadFile(filepath.Join(storePath, "templates", info.Name()))
				if err != nil {
					logrus.Fatalf("error reading template file %s: %s", filepath.Join(storePath, "templates", info.Name()), err)
				}

				match = string(matchBytes)
			}
		}

		if len(match) == 0 {
			logrus.Fatalf("Template '%s' doesn't exist.", name)
		}
	} else {
		match = defaultEntry
	}

	tmpl := template.New("template").Delims("<(", ")>").Funcs(sprig.TxtFuncMap())
	tmpl, err = tmpl.Parse(match)
	if err != nil {
		logrus.Fatalf("Error parsing template: %s", err)
	}

	var out bytes.Buffer

	err = tmpl.Execute(&out, context)
	if err != nil {
		logrus.Fatalf("Error executing template: %s", err)
	}

	return out.String()
}

func init() {
	rootCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().StringP("editor", "e", "", "Editor to use (defaults to $EDITOR, then vim)")
	CreateCmd.Flags().StringP("template", "t", "", "Template file to use")
	CreateCmd.Flags().StringToStringP("context", "c", map[string]string{}, "Context for template")
}
