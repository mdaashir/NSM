package utils

import (
	"fmt"
	"os"
	"strings"
)

var (
	debugEnabled bool
	quietEnabled bool
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelSuccess
	LevelWarn
	LevelError
	LevelTip
)

// ConfigureLogger sets up the logger with the given options
func ConfigureLogger(debug, quiet bool) {
	debugEnabled = debug
	quietEnabled = quiet
}

// formatMessage formats a message with optional arguments and color
func formatMessage(level LogLevel, format string, args ...interface{}) string {
	var prefix string
	var color string

	switch level {
	case LevelDebug:
		prefix = "[DEBUG] "
		color = "\033[36m" // Cyan
	case LevelInfo:
		prefix = ""
		color = "\033[0m" // Default
	case LevelSuccess:
		prefix = "✓ "
		color = "\033[32m" // Green
	case LevelWarn:
		prefix = "⚠️ "
		color = "\033[33m" // Yellow
	case LevelError:
		prefix = "✗ "
		color = "\033[31m" // Red
	case LevelTip:
		prefix = "💡 "
		color = "\033[35m" // Magenta
	}

	message := fmt.Sprintf(format, args...)
	if strings.HasSuffix(message, "\n") {
		message = strings.TrimSuffix(message, "\n")
	}

	// Check if we should use colors
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return prefix + message
	}

	return color + prefix + message + "\033[0m"
}

// logMessage outputs a message to the appropriate destination
func logMessage(level LogLevel, format string, args ...interface{}) {
	// Skip debug messages unless debug is enabled
	if level == LevelDebug && !debugEnabled {
		return
	}

	// Skip non-error messages if quiet mode is enabled
	if quietEnabled && level != LevelError {
		return
	}

	message := formatMessage(level, format, args...)

	// Write to the appropriate output
	if level == LevelError {
		_, err := fmt.Fprintln(os.Stderr, message)
		if err != nil {
			return
		}
	} else {
		fmt.Println(message)
	}
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	logMessage(LevelDebug, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	logMessage(LevelInfo, format, args...)
}

// Success logs a success message
func Success(format string, args ...interface{}) {
	logMessage(LevelSuccess, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	logMessage(LevelWarn, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	logMessage(LevelError, format, args...)
}

// Tip logs a tip/hint message
func Tip(format string, args ...interface{}) {
	logMessage(LevelTip, format, args...)
}

// Table formats and prints tabular data
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-*s", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i, w := range widths {
		if i > 0 {
			fmt.Print("-+-")
		}
		fmt.Print(strings.Repeat("-", w))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Print(" | ")
			}
			if i < len(widths) {
				fmt.Printf("%-*s", widths[i], cell)
			}
		}
		fmt.Println()
	}
}
