package unit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"../../utils"
)

func TestLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create new logger with small max size to test rotation
	logger, err := utils.NewLogger(logFile, utils.DEBUG, 100)
	require.NoError(t, err)
	defer logger.Close()

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Read log file contents
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logStr := string(content)
	assert.Contains(t, logStr, "[DEBUG]")
	assert.Contains(t, logStr, "debug message")
	assert.Contains(t, logStr, "[INFO]")
	assert.Contains(t, logStr, "info message")
	assert.Contains(t, logStr, "[WARN]")
	assert.Contains(t, logStr, "warn message")
	assert.Contains(t, logStr, "[ERROR]")
	assert.Contains(t, logStr, "error message")
}

func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotation.log")

	// Create logger with small max size
	logger, err := utils.NewLogger(logFile, utils.DEBUG, 50)
	require.NoError(t, err)
	defer logger.Close()

	// Write enough logs to trigger rotation
	for i := 0; i < 10; i++ {
		logger.Info("this is a long message that should trigger rotation")
	}

	// Check that backup files were created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	backupFound := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "rotation.log.") {
			backupFound = true
			break
		}
	}
	assert.True(t, backupFound, "No backup log file found")
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := &utils.Logger{
		Out:   &buf,
		Level: utils.INFO,
	}

	// Debug should be filtered out
	logger.Debug("debug message")
	assert.Empty(t, buf.String())

	// Info and above should be logged
	logger.Info("info message")
	assert.Contains(t, buf.String(), "info message")

	buf.Reset()
	logger.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")

	buf.Reset()
	logger.Error("error message")
	assert.Contains(t, buf.String(), "error message")
}

func TestLoggerConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "concurrent.log")

	logger, err := utils.NewLogger(logFile, utils.DEBUG, 1024*1024)
	require.NoError(t, err)
	defer logger.Close()

	// Test concurrent logging
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				logger.Info("concurrent log message %d-%d", i, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify log file exists and has content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestLoggerClose(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "close.log")

	logger, err := utils.NewLogger(logFile, utils.DEBUG, 1024)
	require.NoError(t, err)

	// Write some logs
	logger.Info("test message")

	// Close logger
	err = logger.Close()
	require.NoError(t, err)

	// Verify logs were written
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test message")
}

func TestDefaultLogger(t *testing.T) {
	logger1 := utils.GetLogger()
	logger2 := utils.GetLogger()

	// Should return same instance
	assert.Same(t, logger1, logger2)

	// Test global convenience functions
	utils.Debug("debug message")
	utils.Info("info message")
	utils.Warn("warn message")
	utils.Error("error message")
}
