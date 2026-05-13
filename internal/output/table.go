package output

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleWidth(s string) int {
	return len([]rune(ansiRE.ReplaceAllString(s, "")))
}

// Table buffers rows and writes them with ANSI-aware column alignment on Flush.
type Table struct {
	w       io.Writer
	rows    [][]string
	widths  []int
	padding int
}

// NewTable creates a Table with the given headers.
func NewTable(w io.Writer, headers []string) *Table {
	t := &Table{w: w, padding: 2}
	if len(headers) > 0 {
		t.addRow(headers)
	}
	return t
}

func (t *Table) addRow(cells []string) {
	for len(t.widths) < len(cells) {
		t.widths = append(t.widths, 0)
	}
	for i, c := range cells {
		if w := visibleWidth(c); w > t.widths[i] {
			t.widths[i] = w
		}
	}
	t.rows = append(t.rows, cells)
}

// Row adds one data row.
func (t *Table) Row(cells ...any) {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = fmt.Sprint(c)
	}
	t.addRow(parts)
}

// Flush writes all buffered rows with aligned columns.
func (t *Table) Flush() error {
	for _, row := range t.rows {
		var sb strings.Builder
		for i, cell := range row {
			sb.WriteString(cell)
			if i < len(row)-1 {
				pad := t.widths[i] - visibleWidth(cell) + t.padding
				sb.WriteString(strings.Repeat(" ", pad))
			}
		}
		fmt.Fprintln(t.w, sb.String())
	}
	return nil
}

// RelTime renders "N{s,m,h,d} ago" for a duration.
func RelTime(d time.Duration) string {
	s := int64(d.Seconds())
	if s < 60 {
		return fmt.Sprintf("%ds ago", s)
	}
	if s < 3600 {
		return fmt.Sprintf("%dm ago", s/60)
	}
	if s < 86400 {
		return fmt.Sprintf("%dh ago", s/3600)
	}
	return fmt.Sprintf("%dd ago", s/86400)
}
