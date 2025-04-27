package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger provides structured logging capabilities
type Logger struct {
	mu       sync.Mutex
	out      io.Writer
	level    LogLevel
	file     *os.File
	maxSize  int64
	filename string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// NewLogger creates a new logger instance
func NewLogger(filename string, level LogLevel, maxSize int64) (*Logger, error) {
	dir := filepath.Dir(filename)
	if err := EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &Logger{
		out:      io.MultiWriter(os.Stderr, file),
		level:    level,
		file:     file,
		maxSize:  maxSize,
		filename: filename,
	}, nil
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	once.Do(func() {
		logger, err := NewLogger("nsm.log", INFO, 10*1024*1024) // 10MB default size
		if err != nil {
			// Fallback to stderr if file logging fails
			defaultLogger = &Logger{out: os.Stderr, level: INFO}
			defaultLogger.Error("Failed to initialize file logger: %v", err)
			return
		}
		defaultLogger = logger
	})
	return defaultLogger
}

// log formats and writes a log message
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check file size and rotate if needed
	if l.file != nil {
		if info, err := l.file.Stat(); err == nil && info.Size() > l.maxSize {
			l.rotate()
		}
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	file = filepath.Base(file)

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Format message
	msg := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("%s [%s] %s:%d: %s\n",
		timestamp,
		levelNames[level],
		file,
		line,
		msg,
	)

	// Write to output
	if _, err := io.WriteString(l.out, logLine); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
	}
}

// rotate moves the current log file to a timestamped backup
func (l *Logger) rotate() {
	if l.file == nil {
		return
	}

	// Close current file
	l.file.Close()

	// Generate backup name with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s", l.filename, timestamp)

	// Rename current log file to backup
	if err := os.Rename(l.filename, backupName); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		return
	}

	// Open new log file
	file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create new log file: %v\n", err)
		return
	}

	// Update logger with new file
	l.file = file
	l.out = io.MultiWriter(os.Stderr, file)
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

// SetLevel changes the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Close properly closes the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %v", err)
		}
		l.file = nil
	}
	return nil
}

// Global convenience functions that use the default logger

func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}
