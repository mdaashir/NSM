package unit

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mdaashir/NSM/utils"
)

// captureOutput captures stdout/stderr output during test execution
func captureOutput(f func()) (string, string) {
	// Save original stdout/stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Create pipes for capturing output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Run the function that generates output
	f()

	// Close writers and restore original stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	// Read captured output
	var stdout, stderr bytes.Buffer
	io.Copy(&stdout, rOut)
	io.Copy(&stderr, rErr)

	return stdout.String(), stderr.String()
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name          string
		debug         bool
		quiet         bool
		logFunc       func(string, ...interface{})
		message       string
		expectStdout  bool
		expectStderr  bool
		expectInDebug bool
	}{
		{
			name:          "debug message with debug enabled",
			debug:         true,
			quiet:         false,
			logFunc:       utils.Debug,
			message:       "debug message",
			expectStdout:  true,
			expectStderr:  false,
			expectInDebug: true,
		},
		{
			name:          "debug message with debug disabled",
			debug:         false,
			quiet:         false,
			logFunc:       utils.Debug,
			message:       "debug message",
			expectStdout:  false,
			expectStderr:  false,
			expectInDebug: false,
		},
		{
			name:          "error message in quiet mode",
			debug:         false,
			quiet:         true,
			logFunc:       utils.Error,
			message:       "error message",
			expectStdout:  false,
			expectStderr:  true,
			expectInDebug: false,
		},
		{
			name:          "info message in quiet mode",
			debug:         false,
			quiet:         true,
			logFunc:       utils.Info,
			message:       "info message",
			expectStdout:  false,
			expectStderr:  false,
			expectInDebug: false,
		},
		{
			name:          "success message",
			debug:         false,
			quiet:         false,
			logFunc:       utils.Success,
			message:       "success message",
			expectStdout:  true,
			expectStderr:  false,
			expectInDebug: false,
		},
		{
			name:          "warning message",
			debug:         false,
			quiet:         false,
			logFunc:       utils.Warn,
			message:       "warning message",
			expectStdout:  true,
			expectStderr:  false,
			expectInDebug: false,
		},
		{
			name:          "tip message",
			debug:         false,
			quiet:         false,
			logFunc:       utils.Tip,
			message:       "tip message",
			expectStdout:  true,
			expectStderr:  false,
			expectInDebug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.ConfigureLogger(tt.debug, tt.quiet)

			stdout, stderr := captureOutput(func() {
				tt.logFunc(tt.message)
			})

			// Check stdout
			hasStdout := stdout != ""
			if hasStdout != tt.expectStdout {
				t.Errorf("stdout output = %v, want %v", hasStdout, tt.expectStdout)
			}

			// Check stderr
			hasStderr := stderr != ""
			if hasStderr != tt.expectStderr {
				t.Errorf("stderr output = %v, want %v", hasStderr, tt.expectStderr)
			}

			// Check message content
			output := stdout
			if tt.expectStderr {
				output = stderr
			}

			if tt.expectStdout || tt.expectStderr {
				if !strings.Contains(output, tt.message) {
					t.Errorf("output %q does not contain message %q", output, tt.message)
				}
			}
		})
	}
}

func TestTable(t *testing.T) {
	headers := []string{"Name", "Version"}
	rows := [][]string{
		{"gcc", "12.3.0"},
		{"python3", "3.9.0"},
	}

	stdout, _ := captureOutput(func() {
		utils.Table(headers, rows)
	})

	// Check table formatting
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 4 { // Headers + separator + 2 rows
		t.Errorf("Expected 4 lines in table output, got %d", len(lines))
	}

	// Check headers
	if !strings.Contains(lines[0], headers[0]) || !strings.Contains(lines[0], headers[1]) {
		t.Errorf("Table headers not found in output: %s", lines[0])
	}

	// Check separator line
	if !strings.Contains(lines[1], "-+-") {
		t.Errorf("Table separator not found in output: %s", lines[1])
	}

	// Check data rows
	for i, row := range rows {
		line := lines[i+2]
		for _, cell := range row {
			if !strings.Contains(line, cell) {
				t.Errorf("Row %d data %q not found in output: %s", i, cell, line)
			}
		}
	}
}

func TestNoColorOutput(t *testing.T) {
	// Save original env and restore after test
	origNoColor := os.Getenv("NO_COLOR")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("NO_COLOR", origNoColor)
		os.Setenv("TERM", origTerm)
	}()

	tests := []struct {
		name     string
		setEnv   map[string]string
		wantANSI bool
	}{
		{
			name:     "normal output",
			setEnv:   map[string]string{"NO_COLOR": "", "TERM": "xterm"},
			wantANSI: true,
		},
		{
			name:     "NO_COLOR set",
			setEnv:   map[string]string{"NO_COLOR": "1", "TERM": "xterm"},
			wantANSI: false,
		},
		{
			name:     "dumb terminal",
			setEnv:   map[string]string{"NO_COLOR": "", "TERM": "dumb"},
			wantANSI: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test environment
			for k, v := range tt.setEnv {
				os.Setenv(k, v)
			}

			stdout, _ := captureOutput(func() {
				utils.Success("test message")
			})

			hasANSI := strings.Contains(stdout, "\033[")
			if hasANSI != tt.wantANSI {
				t.Errorf("ANSI color codes present = %v, want %v", hasANSI, tt.wantANSI)
			}
		})
	}
}
