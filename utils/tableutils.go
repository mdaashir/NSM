package utils

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// Table implements table formatting functionality
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table with given headers
func NewTable(headers []string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// String returns the formatted table as a string
func (t *Table) String() string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', 0)

	// Write headers
	fmt.Fprintln(w, strings.Join(t.headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-", len(strings.Join(t.headers, "   "))))

	// Write rows
	for _, row := range t.rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
	return sb.String()
}
