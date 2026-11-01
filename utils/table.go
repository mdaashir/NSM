package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// TableFormat identifies the output format used to render a TableRenderer.
type TableFormat int

const (
	// FormatText renders an aligned plain-text table.
	FormatText TableFormat = iota
	// FormatMarkdown renders a GitHub-flavored markdown table.
	FormatMarkdown
	// FormatJSON renders rows as an array of header-keyed objects.
	FormatJSON
	// FormatCSV renders a CSV document.
	FormatCSV
)

// ParseTableFormat converts a case-insensitive format name to a TableFormat.
func ParseTableFormat(value string) (TableFormat, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "text", "table":
		return FormatText, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	default:
		return FormatText, fmt.Errorf("unsupported format %q (expected text, markdown, json, or csv)", value)
	}
}

// TableRenderer renders tabular data in text, markdown, JSON, or CSV form.
type TableRenderer struct {
	headers   []string
	rows      [][]string
	format    TableFormat
	writer    io.Writer
	useColors bool
}

// NewTableRenderer creates a text-format table that writes to stdout.
func NewTableRenderer(headers []string) *TableRenderer {
	return &TableRenderer{
		headers:   append([]string(nil), headers...),
		rows:      make([][]string, 0),
		format:    FormatText,
		writer:    os.Stdout,
		useColors: os.Getenv("NO_COLOR") == "",
	}
}

// AddRow appends a row to the table.
func (t *TableRenderer) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// SetFormat sets the output format.
func (t *TableRenderer) SetFormat(format TableFormat) {
	t.format = format
}

// DisableColors disables ANSI colors in text output.
func (t *TableRenderer) DisableColors() {
	t.useColors = false
}

// SetWriter sets the destination writer.
func (t *TableRenderer) SetWriter(w io.Writer) {
	t.writer = w
}

// Render writes the table to the configured writer.
func (t *TableRenderer) Render() error {
	if len(t.headers) == 0 {
		return fmt.Errorf("table requires at least one header")
	}

	switch t.format {
	case FormatText:
		return t.renderText()
	case FormatMarkdown:
		return t.renderMarkdown()
	case FormatJSON:
		return t.renderJSON()
	case FormatCSV:
		return t.renderCSV()
	default:
		return fmt.Errorf("unsupported format: %d", t.format)
	}
}

// String renders the table and returns it as a string.
func (t *TableRenderer) String() string {
	var b strings.Builder
	writer := t.writer
	t.writer = &b
	err := t.Render()
	t.writer = writer
	if err != nil {
		return ""
	}
	return b.String()
}

func (t *TableRenderer) renderText() error {
	widths := t.columnWidths()
	header := formatTextRow(t.headers, widths)
	separator := formatTextSeparator(widths)

	if t.useColors {
		fmt.Fprintln(t.writer, "\033[1;36m"+header+"\033[0m")
		fmt.Fprintln(t.writer, "\033[0;37m"+separator+"\033[0m")
	} else {
		fmt.Fprintln(t.writer, header)
		fmt.Fprintln(t.writer, separator)
	}

	for _, row := range t.normalizedRows() {
		fmt.Fprintln(t.writer, formatTextRow(row, widths))
	}

	return nil
}

func (t *TableRenderer) renderMarkdown() error {
	headers := make([]string, len(t.headers))
	for i, header := range t.headers {
		headers[i] = escapeMarkdownCell(header)
	}
	fmt.Fprintf(t.writer, "| %s |\n", strings.Join(headers, " | "))

	separators := make([]string, len(t.headers))
	for i := range separators {
		separators[i] = "---"
	}
	fmt.Fprintf(t.writer, "| %s |\n", strings.Join(separators, " | "))

	for _, row := range t.normalizedRows() {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = escapeMarkdownCell(cell)
		}
		fmt.Fprintf(t.writer, "| %s |\n", strings.Join(cells, " | "))
	}

	return nil
}

func (t *TableRenderer) renderJSON() error {
	result := make([]map[string]string, 0, len(t.rows))

	for _, row := range t.normalizedRows() {
		record := make(map[string]string, len(t.headers))
		for i, header := range t.headers {
			record[header] = row[i]
		}
		result = append(result, record)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(t.writer, string(data))
	return err
}

func (t *TableRenderer) renderCSV() error {
	writer := csv.NewWriter(t.writer)

	if err := writer.Write(t.headers); err != nil {
		return err
	}

	for _, row := range t.normalizedRows() {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func (t *TableRenderer) columnWidths() []int {
	widths := make([]int, len(t.headers))
	for i, header := range t.headers {
		widths[i] = utf8.RuneCountInString(header)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			if width := utf8.RuneCountInString(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}

	return widths
}

func (t *TableRenderer) normalizedRows() [][]string {
	rows := make([][]string, 0, len(t.rows))

	for _, row := range t.rows {
		normalized := make([]string, len(t.headers))
		copy(normalized, row)
		rows = append(rows, normalized)
	}

	return rows
}

func formatTextRow(cells []string, widths []int) string {
	parts := make([]string, len(widths))

	for i, width := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}

		padding := width - utf8.RuneCountInString(cell)
		if padding < 0 {
			padding = 0
		}
		parts[i] = cell + strings.Repeat(" ", padding)
	}

	return strings.Join(parts, " | ")
}

func formatTextSeparator(widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	return strings.Join(parts, "-+-")
}

func escapeMarkdownCell(cell string) string {
	cell = strings.ReplaceAll(cell, "\\", "\\\\")
	cell = strings.ReplaceAll(cell, "|", "\\|")
	cell = strings.ReplaceAll(cell, "\r\n", "<br>")
	cell = strings.ReplaceAll(cell, "\n", "<br>")
	cell = strings.ReplaceAll(cell, "\r", "<br>")
	return cell
}
