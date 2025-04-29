package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed troubleshooting
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for situations that might cause problems
	WARN
	// ERROR level for failures that should be addressed
	ERROR
	// FATAL level for critical failures that require immediate attention
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger represents a logger with multiple outputs and levels
type Logger struct {
	level      LogLevel
	outputs    map[string]io.Writer
	mu         sync.Mutex
	timeFormat string
	fileInfo   bool
}

// DefaultLogger is the global instance used by package-level functions
var DefaultLogger *Logger
var once sync.Once

// ConfigureLogger initializes the logger with specified settings
func ConfigureLogger(level LogLevel, logFilePath string, enableConsole bool) error {
	var err error
	once.Do(func() {
		DefaultLogger = &Logger{
			level:      level,
			outputs:    make(map[string]io.Writer),
			timeFormat: "2006-01-02 15:04:05",
			fileInfo:   level == DEBUG,
		}

		if enableConsole {
			DefaultLogger.outputs["console"] = os.Stdout
		}

		if logFilePath != "" {
			if err = os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
				return
			}

			var file *os.File
			file, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return
			}

			DefaultLogger.outputs["file"] = file
		}
	})
	return err
}

// AddLogRotation sets up log rotation based on file size or time period
func (l *Logger) AddLogRotation(maxSizeMB int, maxAgeDays int, logDir string, baseFilename string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Implementation of log rotation logic would go here
	// In production, you'd typically use a library like lumberjack or zap
	// This is a simplified placeholder

	// For now, just ensure the log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	return nil
}

// log formats and outputs a log message to all configured outputs
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(l.timeFormat)
	message := fmt.Sprintf(format, args...)

	var fileInfo string
	if l.fileInfo {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			file = filepath.Base(file)
			fileInfo = fmt.Sprintf(" [%s:%d]", file, line)
		}
	}

	logLine := fmt.Sprintf("[%s] %s%s: %s\n", timestamp, levelNames[level], fileInfo, message)

	for _, writer := range l.outputs {
		fmt.Fprint(writer, logLine)
	}

	// Auto-flush on fatal
	if level == FATAL {
		for name, writer := range l.outputs {
			if f, ok := writer.(*os.File); ok && f != os.Stdout && f != os.Stderr {
				f.Sync()
			}
			if name == "file" {
				if f, ok := writer.(*os.File); ok {
					f.Close()
				}
			}
		}
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// SetLevel changes the minimum level of messages to log
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	l.fileInfo = level == DEBUG
}

// Package-level functions that use the DefaultLogger

func init() {
	// Initialize with sensible defaults if not explicitly configured
	if DefaultLogger == nil {
		err := ConfigureLogger(INFO, "", true)
		if err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}
	}
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debug(format, args...)
	}
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warn(format, args...)
	}
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Error(format, args...)
	}
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Fatal(format, args...)
	} else {
		log.Fatalf(format, args...)
	}
}

// PromptUser asks the user a yes/no question and returns their response
func PromptUser(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/n): ", question)
	response, err := reader.ReadString('\n')
	if err != nil {
		Error("Error reading input: %v", err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// PromptContinue asks the user if they want to continue to the next step
// and returns their response
func PromptContinue(nextAction string) bool {
	return PromptUser(fmt.Sprintf("Continue to %s?", nextAction))
}
