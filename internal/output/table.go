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

// RelTime renders "N{s,m,h,d} ago" for a duration in seconds.
func RelTime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds ago", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm ago", seconds/60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh ago", seconds/3600)
	}
	return fmt.Sprintf("%dd ago", seconds/86400)
}
