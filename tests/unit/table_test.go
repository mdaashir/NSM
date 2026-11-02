package unit

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mdaashir/NSM/utils"
)

func TestTableTextFormat(t *testing.T) {
	table := utils.NewTableRenderer([]string{"A", "B"})
	table.DisableColors()
	table.AddRow([]string{"1", "2"})

	got := table.String()
	want := "A | B\n--+--\n1 | 2\n"
	if got != want {
		t.Fatalf("unexpected table output:\ngot  %q\nwant %q", got, want)
	}
}

func TestTableTextPadsShortRows(t *testing.T) {
	table := utils.NewTableRenderer([]string{"A", "B"})
	table.DisableColors()
	table.AddRow([]string{"only"})

	lines := strings.Split(strings.TrimRight(table.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if !strings.HasPrefix(lines[2], "only |") {
		t.Errorf("short row was not padded: %q", lines[2])
	}

	if strings.Contains(table.String(), "\033[") {
		t.Error("disabled colors still produced ANSI escapes")
	}
}

func TestTableMarkdownFormat(t *testing.T) {
	table := utils.NewTableRenderer([]string{"Name", "Value"})
	table.SetFormat(utils.FormatMarkdown)
	table.AddRow([]string{"pipe", "a|b"})
	table.AddRow([]string{"newline", "a\nb"})

	got := table.String()
	want := "| Name | Value |\n| --- | --- |\n| pipe | a\\|b |\n| newline | a<br>b |\n"
	if got != want {
		t.Fatalf("unexpected markdown output:\ngot  %q\nwant %q", got, want)
	}
}

func TestTableJSONFormat(t *testing.T) {
	table := utils.NewTableRenderer([]string{"Package", "Status"})
	table.SetFormat(utils.FormatJSON)
	table.AddRow([]string{"gcc", "installed"})
	table.AddRow([]string{"python3", "pending"})

	var rows []map[string]string
	if err := json.Unmarshal([]byte(table.String()), &rows); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0]["Package"] != "gcc" || rows[0]["Status"] != "installed" {
		t.Errorf("unexpected first row: %v", rows[0])
	}

	if rows[1]["Package"] != "python3" || rows[1]["Status"] != "pending" {
		t.Errorf("unexpected second row: %v", rows[1])
	}
}

func TestTableJSONEmpty(t *testing.T) {
	table := utils.NewTableRenderer([]string{"A"})
	table.SetFormat(utils.FormatJSON)

	if got := strings.TrimSpace(table.String()); got != "[]" {
		t.Errorf("empty table JSON = %q, want []", got)
	}
}

func TestTableCSVFormat(t *testing.T) {
	table := utils.NewTableRenderer([]string{"Name", "Note"})
	table.SetFormat(utils.FormatCSV)
	table.AddRow([]string{"gcc", "has,comma"})
	table.AddRow([]string{"quote", `say "hi"`})
	table.AddRow([]string{"multi", "line1\nline2"})

	var buf bytes.Buffer
	table.SetWriter(&buf)
	if err := table.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	records, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV output: %v", err)
	}

	want := [][]string{
		{"Name", "Note"},
		{"gcc", "has,comma"},
		{"quote", `say "hi"`},
		{"multi", "line1\nline2"},
	}

	if !reflect.DeepEqual(records, want) {
		t.Errorf("CSV records = %v, want %v", records, want)
	}
}

func TestParseTableFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    utils.TableFormat
		wantErr bool
	}{
		{"", utils.FormatText, false},
		{"text", utils.FormatText, false},
		{"TABLE", utils.FormatText, false},
		{"md", utils.FormatMarkdown, false},
		{"markdown", utils.FormatMarkdown, false},
		{"json", utils.FormatJSON, false},
		{"CSV", utils.FormatCSV, false},
		{"yaml", utils.FormatText, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := utils.ParseTableFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseTableFormat(%q) error = %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("ParseTableFormat(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTableRenderErrors(t *testing.T) {
	table := utils.NewTableRenderer(nil)
	table.SetWriter(io.Discard)
	if err := table.Render(); err == nil {
		t.Error("expected an error for a table without headers")
	}

	table = utils.NewTableRenderer([]string{"A"})
	table.SetWriter(io.Discard)
	table.SetFormat(utils.TableFormat(99))
	if err := table.Render(); err == nil {
		t.Error("expected an error for an unsupported format")
	}
}

func TestTableRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	table := utils.NewTableRenderer([]string{"A"})
	table.AddRow([]string{"1"})

	if strings.Contains(table.String(), "\033[") {
		t.Error("NO_COLOR table output contains ANSI escapes")
	}
}
