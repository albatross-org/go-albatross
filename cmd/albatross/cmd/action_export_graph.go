package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/albatross-org/go-albatross/entries"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
)

// ActionExportGraphCmd represents the 'export graph' action.
var ActionExportGraphCmd = &cobra.Command{
	Use:     "graph",
	Aliases: []string{"dot", "graphviz"},
	Short:   "print a DOT-formatted graph representing the entries",
	Run: func(cmd *cobra.Command, args []string) {
		_, collection, list := getFromCommand(cmd)

		var tmpl = template.New("input").Funcs(sprig.TxtFuncMap())

		tmplStr := "{{.Path}}"
		tmpl, err := tmpl.Parse(tmplStr)
		if err != nil {
			fmt.Println("Error parsing template:")
			fmt.Println(err)
			os.Exit(1)
		}

		var out bytes.Buffer
		out.WriteString("digraph {\n")

		for _, entry := range list.Slice() {
			// TODO: this is a rough escape
			stringerEntry := entries.NewStringerTemplate(entry, tmpl)
			escapedOutput := strings.ReplaceAll(stringerEntry.String(), `"`, `\"`)

			for _, link := range entry.OutboundLinks {
				linkedEntry := collection.ResolveLink(link)
				stringerLink := entries.NewStringerTemplate(linkedEntry, tmpl)
				escapedLink := strings.ReplaceAll(stringerLink.String(), `"`, `\"`)

				out.WriteString(fmt.Sprintf("\t\"%s\" -> \"%s\"\n", escapedOutput, escapedLink))
			}
		}

		out.WriteString("}")

		fmt.Println(out.String())
	},
}

func init() {
	ActionExportCmd.AddCommand(ActionExportGraphCmd)
}
