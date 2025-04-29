// Package utils provides utility functions for file operations
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	fileLocks sync.Map
)

// FileExists checks if a file exists and is not a directory
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		Debug("Error checking if file exists: %v", err)
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		Debug("Error checking if directory exists: %v", err)
		return false
	}
	return info.IsDir()
}

// BackupFile creates a backup of a file with timestamp
func BackupFile(path string) error {
	if !FileExists(path) {
		return fmt.Errorf("file %s does not exist", path)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.backup", path, timestamp)

	Debug("Creating backup of %s to %s", path, backupPath)
	return CopyFile(path, backupPath)
}

// GetProjectConfigType returns the type of project configuration (shell.nix or flake.nix)
func GetProjectConfigType() string {
	if FileExists("shell.nix") {
		return "shell.nix"
	}
	if FileExists("flake.nix") {
		return "flake.nix"
	}
	return ""
}

// SafeWrite writes data to a file atomically using a temporary file
func SafeWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	// Ensure directory exists
	if err := EnsureDir(dir); err != nil {
		return err
	}

	// Create temp file in same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp-nsm-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("failed to sync temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Set permissions before renaming
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to chmod temp file: %v", err)
	}

	// Take a backup if file exists
	if FileExists(path) {
		if err := BackupFile(path); err != nil {
			Debug("Failed to backup file before overwriting: %v", err)
		}
	}

	// Acquire lock for the target path
	lock := AcquireLock(path)
	defer lock.Release()

	// Rename temp file to target path
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move temp file: %v", err)
	}

	Debug("Successfully wrote file: %s (%d bytes)", path, len(data))
	return nil
}

// SafeRead reads a file with proper error handling
func SafeRead(path string) ([]byte, error) {
	if !FileExists(path) {
		return nil, fmt.Errorf("file %s does not exist", path)
	}

	// Acquire lock for the file path
	lock := AcquireLock(path)
	defer lock.Release()

	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return data, nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("empty directory path")
	}

	if !DirExists(path) {
		Debug("Creating directory: %s", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}
	return nil
}

// FileLock provides a simple file locking mechanism
type FileLock struct {
	path string
	mu   sync.Mutex
}

// AcquireLock acquires a lock for a file path
func AcquireLock(path string) *FileLock {
	normalizedPath := filepath.Clean(path)
	val, _ := fileLocks.LoadOrStore(normalizedPath, &FileLock{path: normalizedPath})
	lock := val.(*FileLock)
	lock.mu.Lock()
	Debug("Acquired lock for: %s", normalizedPath)
	return lock
}

// Release releases the file lock
func (l *FileLock) Release() {
	Debug("Released lock for: %s", l.path)
	l.mu.Unlock()
}

// CopyFile copies a file with proper error handling
func CopyFile(src, dst string) error {
	if src == dst {
		return fmt.Errorf("source and destination are the same")
	}

	if !FileExists(src) {
		return fmt.Errorf("source file %s does not exist", src)
	}

	// Acquire locks for both source and destination
	srcLock := AcquireLock(src)
	defer srcLock.Release()

	dstLock := AcquireLock(dst)
	defer dstLock.Release()

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	dstFile, err := os.CreateTemp(filepath.Dir(dst), ".tmp-nsm-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpPath := dstFile.Name()

	cleanup := func() {
		dstFile.Close()
		os.Remove(tmpPath)
	}

	// Copy the contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		cleanup()
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	if err := dstFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("failed to sync destination file: %v", err)
	}

	if err := dstFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close destination file: %v", err)
	}

	// Copy source permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	if err := os.Chmod(tmpPath, srcInfo.Mode()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to chmod temp file: %v", err)
	}

	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move temp file: %v", err)
	}

	Debug("Successfully copied file: %s to %s", src, dst)
	return nil
}

// RemovePath safely removes a file or directory
func RemovePath(path string) error {
	if path == "" || path == "/" || path == "." || path == ".." {
		return fmt.Errorf("invalid path for removal: %s", path)
	}

	lock := AcquireLock(path)
	defer lock.Release()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Take backup if it's a file
	if FileExists(path) {
		if err := BackupFile(path); err != nil {
			Debug("Failed to backup file before removal: %v", err)
		}
	}

	Debug("Removing path: %s", path)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove path: %v", err)
	}

	return nil
}

// ValidatePath checks if a path is safe to use
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for directory traversal
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path contains directory traversal")
	}

	// Check if path is absolute when it shouldn't be
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	return nil
}

// GetFileSize returns the size of a file
func GetFileSize(path string) (int64, error) {
	if !FileExists(path) {
		return 0, fmt.Errorf("file %s does not exist", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %v", err)
	}

	if info.IsDir() {
		return 0, fmt.Errorf("path is a directory")
	}

	return info.Size(), nil
}

// IsEmptyDir checks if a directory is empty
func IsEmptyDir(path string) (bool, error) {
	if !DirExists(path) {
		return false, fmt.Errorf("directory %s does not exist", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open directory: %v", err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}

// ReadFile reads a file with proper error handling
func ReadFile(path string) (string, error) {
	if !FileExists(path) {
		return "", fmt.Errorf("file %s does not exist", path)
	}

	// Use SafeRead for consistent handling
	content, err := SafeRead(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile writes data to a file with proper error handling
func WriteFile(path string, content string) error {
	return SafeWrite(path, []byte(content), 0600)
}

// EnsureConfigDir ensures the configuration directory exists
func EnsureConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	configDir := filepath.Join(home, ".config", "nsm")
	if err := EnsureDir(configDir); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	// Ensure logs directory exists
	logsDir := filepath.Join(configDir, "logs")
	if err := EnsureDir(logsDir); err != nil {
		Debug("Failed to create logs directory: %v", err)
	}

	Debug("Config directory: %s", configDir)
	return configDir, nil
}
