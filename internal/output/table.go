package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Table is a thin wrapper over text/tabwriter for aligned columnar output.
type Table struct {
	tw *tabwriter.Writer
}

// NewTable creates a Table with the given headers, written immediately.
func NewTable(w io.Writer, headers []string) *Table {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if len(headers) > 0 {
		fmt.Fprintln(tw, strings.Join(headers, "\t"))
	}
	return &Table{tw: tw}
}

// Row prints one row. Each cell is converted with fmt.Sprint.
func (t *Table) Row(cells ...any) {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = fmt.Sprint(c)
	}
	fmt.Fprintln(t.tw, strings.Join(parts, "\t"))
}

// Flush flushes the tabwriter; call before writing further unaligned output.
func (t *Table) Flush() error { return t.tw.Flush() }
