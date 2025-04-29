package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// TableFormat specifies the output format for tables
type TableFormat int

const (
	// FormatText outputs tables in plain text
	FormatText TableFormat = iota
	// FormatMarkdown outputs tables in markdown
	FormatMarkdown
	// FormatJSON outputs tables as JSON
	FormatJSON
	// FormatCSV outputs tables as CSV
	FormatCSV
)

// TableTheme contains color settings for table rendering
type TableTheme struct {
	HeaderFg string // Header foreground color
	BorderFg string // Border foreground color
	RowFg    string // Row foreground color
	AltRowFg string // Alternate row foreground color
}

// DefaultTheme is the default color theme
var DefaultTheme = TableTheme{
	HeaderFg: "\033[1;36m", // Cyan bold
	BorderFg: "\033[0;37m", // Light gray
	RowFg:    "\033[0m",    // Default
	AltRowFg: "\033[0;37m", // Light gray
}

// NoColorTheme has no colors
var NoColorTheme = TableTheme{
	HeaderFg: "",
	BorderFg: "",
	RowFg:    "",
	AltRowFg: "",
}

// Table implements table formatting functionality
type Table struct {
	Headers   []string
	Rows      [][]string
	Theme     TableTheme
	Format    TableFormat
	Writer    io.Writer
	MinWidth  int
	TabWidth  int
	Padding   int
	UseColors bool
}

// NewTable creates a new table with given headers
func NewTable(headers []string) *Table {
	return &Table{
		Headers:   headers,
		Rows:      make([][]string, 0),
		Theme:     DefaultTheme,
		Format:    FormatText,
		Writer:    os.Stdout,
		MinWidth:  0,
		TabWidth:  3,
		Padding:   1,
		UseColors: true,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

// SetFormat sets the output format
func (t *Table) SetFormat(format TableFormat) {
	t.Format = format
}

// DisableColors disables color output
func (t *Table) DisableColors() {
	t.UseColors = false
}

// SetWriter sets the output writer
func (t *Table) SetWriter(w io.Writer) {
	t.Writer = w
}

// SetTabWidth sets the tab width
func (t *Table) SetTabWidth(width int) {
	t.TabWidth = width
}

// Render renders the table to the configured writer
func (t *Table) Render() error {
	switch t.Format {
	case FormatText:
		return t.renderText()
	case FormatMarkdown:
		return t.renderMarkdown()
	case FormatJSON:
		return t.renderJSON()
	case FormatCSV:
		return t.renderCSV()
	default:
		return fmt.Errorf("unsupported format: %d", t.Format)
	}
}

// String returns the formatted table as a string
func (t *Table) String() string {
	var sb strings.Builder
	oldWriter := t.Writer
	t.Writer = &sb
	_ = t.Render()
	t.Writer = oldWriter
	return sb.String()
}

// renderText renders the table in text format
func (t *Table) renderText() error {
	w := tabwriter.NewWriter(t.Writer, t.MinWidth, t.TabWidth, t.Padding, ' ', 0)

	// Get theme
	theme := t.Theme
	if !t.UseColors {
		theme = NoColorTheme
	}

	// Write headers
	fmt.Fprintln(w, theme.HeaderFg+strings.Join(t.Headers, "\t")+"\033[0m")

	// Calculate border width
	borderWidth := 0
	for _, h := range t.Headers {
		borderWidth += len(h) + t.TabWidth
	}

	// Write border
	fmt.Fprintln(w, theme.BorderFg+strings.Repeat("-", borderWidth)+"\033[0m")

	// Write rows
	for i, row := range t.Rows {
		rowFg := theme.RowFg
		if i%2 == 1 {
			rowFg = theme.AltRowFg
		}

		// Ensure row has enough columns
		for len(row) < len(t.Headers) {
			row = append(row, "")
		}

		fmt.Fprintln(w, rowFg+strings.Join(row, "\t")+"\033[0m")
	}

	return w.Flush()
}

// renderMarkdown renders the table in markdown format
func (t *Table) renderMarkdown() error {
	// Write headers
	fmt.Fprintln(t.Writer, "| "+strings.Join(t.Headers, " | ")+" |")

	// Write separator
	var separators []string
	for range t.Headers {
		separators = append(separators, "---")
	}
	fmt.Fprintln(t.Writer, "| "+strings.Join(separators, " | ")+" |")

	// Write rows
	for _, row := range t.Rows {
		// Ensure row has enough columns
		for len(row) < len(t.Headers) {
			row = append(row, "")
		}

		// Escape pipe characters in cells
		for i, cell := range row {
			row[i] = strings.ReplaceAll(cell, "|", "\\|")
		}

		fmt.Fprintln(t.Writer, "| "+strings.Join(row, " | ")+" |")
	}

	return nil
}

// renderJSON renders the table in JSON format
func (t *Table) renderJSON() error {
	result := make([]map[string]string, 0, len(t.Rows))

	for _, row := range t.Rows {
		rowMap := make(map[string]string)
		for i, header := range t.Headers {
			if i < len(row) {
				rowMap[header] = row[i]
			} else {
				rowMap[header] = ""
			}
		}
		result = append(result, rowMap)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(t.Writer, string(jsonData))
	return err
}

// renderCSV renders the table in CSV format
func (t *Table) renderCSV() error {
	// Write headers
	fmt.Fprintln(t.Writer, escapeCSV(t.Headers))

	// Write rows
	for _, row := range t.Rows {
		// Ensure row has enough columns
		for len(row) < len(t.Headers) {
			row = append(row, "")
		}

		fmt.Fprintln(t.Writer, escapeCSV(row))
	}

	return nil
}

// escapeCSV escapes and joins values for CSV output
func escapeCSV(values []string) string {
	var escaped []string

	for _, v := range values {
		// If the value contains a comma, quote, or newline, wrap it in quotes
		if strings.ContainsAny(v, ",\"\n\r") {
			// Escape quotes by doubling them
			v = strings.ReplaceAll(v, "\"", "\"\"")
			v = "\"" + v + "\""
		}
		escaped = append(escaped, v)
	}

	return strings.Join(escaped, ",")
}

// FormatDiagnosticTable formats diagnostic results as a table
func FormatDiagnosticTable(results []DoctorResult, format TableFormat) string {
	table := NewTable([]string{"Check", "Status", "Message", "Fix"})

	if format != FormatText {
		table.DisableColors()
	}

	table.SetFormat(format)

	for _, result := range results {
		status := result.Status

		// Add color to status for text format
		if format == FormatText {
			switch result.Status {
			case StatusOK:
				status = "\033[1;32m" + status + "\033[0m" // Green
			case StatusWarning:
				status = "\033[1;33m" + status + "\033[0m" // Yellow
			case StatusError:
				status = "\033[1;31m" + status + "\033[0m" // Red
			}
		}

		table.AddRow([]string{
			result.Name,
			status,
			result.Message,
			result.Fix,
		})
	}

	return table.String()
}
