package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
)

// ActionTemplateCmd represents the 'tags' action.
var ActionTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "fill a template",
	Long: `template executes Go templates

This can be used to manipulate entries in a variety of ways. For example:

    # Print titles for all entries.
	$ albatross get template "{.Title}"
	
	# Print all entries in the format
	# YYYY-MM-DD (PATH) TITLE
	$ albatross get template '{{.Date | date "2006-01-02"}} ({{.Path}}) {{.Title}}'

You can also read templates in from stdin:

	$ cat template.txt | albatross get template
	
Sprig (https://github.com/Masterminds/sprig) helper functions/pipelines are available, such as:

    - date
    - toJSON
    - upper
	
By default, the template is run against every entry matched sequentially. If you wish to access the list
of entries itself, use the --all flag.

	$ albatross get template --all '{{range .}}{{.Title}}{{end}}'
	
Context
-------

The template context is an Entry struct. This contains the following fields:

	- .Path, string
	  The path to the entry file.

	- .Contents, string
	  The contents of the file without front matter.

	- .OriginalContents, string
	  The contents of the file, includiong the front matter.

	- .Tags, []string
	  All the tags present in the document. For example, "@!journal".

	- .OutboundLinks, []Link
	  Links going from this entry to another one.

	- .Date, time.Time
	  The date extracted from the entry.

	- .ModTime, time.Time
	  The modification time for the entry. This is not always accurate, since encrypting and decryting all the
	  files will "modify" them. Therefore it cannot be used for sorting accurately.

	- .Title, string
	  The title of the entry.

	- .Metadata, map[string]interface{}
	  All of the front matter. `,

	Run: func(cmd *cobra.Command, args []string) {
		_, _, list := getFromCommand(cmd)

		all, err := cmd.Flags().GetBool("all")
		checkArg(err)

		var tmpl = template.New("input").Funcs(sprig.TxtFuncMap())

		fi, err := os.Stdin.Stat()
		if err != nil {
			panic(err)
		}

		if len(args) == 1 {
			tmpl, err = tmpl.Parse(args[0])
		} else if fi.Mode()&os.ModeNamedPipe == 0 {
			fmt.Println("template takes one arg, the template, or reads from stdin:")
			fmt.Println("")
			fmt.Println("    $ albatross get template '{{.Title}}'")
			fmt.Println("    $ cat template.txt | albatross get template")
			os.Exit(1)
		} else {
			stdin, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println("Error reading template from stdin:")
				fmt.Println(err)
				os.Exit(1)
			}

			tmpl, err = tmpl.Parse(string(stdin))
			if err != nil {
				fmt.Println("Error parsing template from stdin:")
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if err != nil {
			fmt.Println("Error parsing your template:")
			fmt.Println(err)
			fmt.Println("")
			fmt.Println("Please consult https://golang.org/pkg/text/template/ for valid template syntax.")
			os.Exit(1)
		}

		if all {
			err = tmpl.Execute(os.Stdout, list.Slice())
			if err != nil {
				fmt.Println("Error executing template:")
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			for _, entry := range list.Slice() {
				err = tmpl.Execute(os.Stdout, entry)
				if err != nil {
					fmt.Println("Error executing template:")
					fmt.Println(err)
					os.Exit(1)
				}

				fmt.Println("")
			}
		}
	},
}

func init() {
	GetCmd.AddCommand(ActionTemplateCmd)
	ActionTemplateCmd.Flags().Bool("all", false, "Run a template on all entries instead of each one sequentially")
}
