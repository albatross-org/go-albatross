package cmd

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ActionExportGraphCmd represents the 'export graph' action.
var ActionExportGraphCmd = &cobra.Command{
	Use:     "graph",
	Aliases: []string{"dot", "graphviz"},
	Short: `print a DOT-formatted graph representing the entries
	
This command can be used to visualise a store using Graphviz software by outputting a DOT formatted representation
of all entries and their links.

For example, to export all entries:

	$ albatross get export graph | dot -T svg -o /tmp/out.svg
	digraph {
		"notes/books/flow" -> "notes/books/flow/chapters/1-happiness-revisited"
		...
	}
	
You can specify graph attributes using the --graph-attributes/-g flag:

	$ albatross get export graph -g overlap=scale -g fontname=Times-Roman | dot -T svg -o /tmp/out.svg
	# digraph {
		overlap=scale
		fontname=Times-Roman
		"notes/books/flow" -> "notes/books/flow/chapters/1-happiness-revisited"
		...	
	}
	
Currently, only displaying entry paths is supported. Entries that link to other entries outside the scope
of the search will not be displayed.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, collection, list := getFromCommand(cmd)

		attributes, err := cmd.Flags().GetStringToString("graph-attributes")
		checkArg(err)

		var out bytes.Buffer

		out.WriteString("digraph {\n")
		for key, val := range attributes {
			out.WriteString("\t")
			out.WriteString(key)
			out.WriteString("=")
			out.WriteString(val)
			out.WriteString("\n")
		}

		for _, entry := range list.Slice() {
			escapedOutput := strconv.Quote(entry.Path)

			for _, link := range entry.OutboundLinks {
				linkedEntry := collection.ResolveLink(link)
				if linkedEntry == nil {
					continue
				}
				escapedLink := strconv.Quote(linkedEntry.Path)

				out.WriteString(fmt.Sprintf("\t%s -> %s\n", escapedOutput, escapedLink))
			}
		}

		out.WriteString("}")

		fmt.Println(out.String())
	},
}

func init() {
	ActionExportCmd.AddCommand(ActionExportGraphCmd)
	ActionExportGraphCmd.Flags().StringToStringP("graph-attributes", "g", map[string]string{}, "graph attributes (https://graphviz.org/doc/info/attrs.html)")
}
