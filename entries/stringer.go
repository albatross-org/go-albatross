package entries

import (
	"bytes"
	"fmt"
	"text/template"
)

// StringerTitle is an entry which prints a title.
type StringerTitle struct {
	*Entry
}

// String prints the entry's title.
func (e *StringerTitle) String() string {
	return e.Title
}

// StringerTemplate is an entry which can print an entry according to a template.
type StringerTemplate struct {
	entry *Entry
	tmpl  *template.Template
}

// NewStringerTemplate returns a new StringerTemplate structure.
func NewStringerTemplate(entry *Entry, tmpl *template.Template) *StringerTemplate {
	return &StringerTemplate{
		entry: entry,
		tmpl:  tmpl,
	}
}

// String prints the entry according to the template.
func (e *StringerTemplate) String() string {
	var buf bytes.Buffer

	err := e.tmpl.Execute(&buf, e.entry)
	if err != nil {
		return fmt.Sprintf("error in template: %s", err.Error())
	}

	return buf.String()
}
