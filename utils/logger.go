package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger handles application logging
type Logger struct {
	Out        io.Writer
	Level      LogLevel
	Prefix     string
	WithSource bool // Whether to include source file info
}

var defaultLogger *Logger

// Initialize default logger
func init() {
	defaultLogger = &Logger{
		Out:        os.Stdout,
		Level:      INFO,
		Prefix:     "",
		WithSource: false,
	}
}

// NewLogger creates a new logger instance
func NewLogger(path string, level LogLevel, maxSize int64) (*Logger, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &Logger{
		Out:        file,
		Level:      level,
		Prefix:     "",
		WithSource: true,
	}, nil
}

// ConfigureLogger configures the default logger
func ConfigureLogger(level LogLevel, output io.Writer) {
	defaultLogger.Level = level
	defaultLogger.Out = output

	// Enable source information in debug mode
	defaultLogger.WithSource = (level == DEBUG)
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Success logs a success message
func Success(format string, v ...interface{}) {
	defaultLogger.Success(format, v...)
}

// Tip logs a tip message
func Tip(format string, v ...interface{}) {
	defaultLogger.Tip(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.Level <= DEBUG {
		l.log("DEBUG", format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.Level <= INFO {
		l.log("INFO", format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.Level <= WARN {
		l.log("WARN", format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.Level <= ERROR {
		l.log("ERROR", format, v...)
	}
}

// Success logs a success message
func (l *Logger) Success(format string, v ...interface{}) {
	l.log("SUCCESS", format, v...)
}

// Tip logs a tip message
func (l *Logger) Tip(format string, v ...interface{}) {
	l.log("TIP", format, v...)
}

// getSourceInfo returns the calling file and line number
func getSourceInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// log formats and writes a log message
func (l *Logger) log(level, format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, v...)

	var logLine string
	if l.WithSource {
		source := getSourceInfo()
		logLine = fmt.Sprintf("[%s] %s (%s): %s\n", level, timestamp, source, message)
	} else {
		logLine = fmt.Sprintf("[%s] %s: %s\n", level, timestamp, message)
	}

	fmt.Fprint(l.Out, logLine)
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	return defaultLogger
}

// Close closes the logger output
func (l *Logger) Close() error {
	if closer, ok := l.Out.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// EnableDebugSourceInfo enables source file information in logs
func EnableDebugSourceInfo(enable bool) {
	defaultLogger.WithSource = enable
}
