// Package utils provides utility functions for file operations, logging, configuration,
// and Nix-related functionality for the NSM (Nix Shell Manager) application.
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	fileLocks sync.Map
)

// SafeWrite writes data to a file atomically using a temporary file
func SafeWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	// Create temp file in same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}
	defer cleanup()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Set permissions before renaming
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to chmod temp file: %v", err)
	}

	// Rename temp file to target path
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to move temp file: %v", err)
	}

	return nil
}

// SafeRead reads a file with proper error handling
func SafeRead(path string) ([]byte, error) {
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
	if _, err := os.Stat(path); os.IsNotExist(err) {
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
	return lock
}

// Release releases the file lock
func (l *FileLock) Release() {
	l.mu.Unlock()
}

// CopyFile copies a file with proper error handling
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	dstFile, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpPath := dstFile.Name()

	cleanup := func() {
		dstFile.Close()
		os.Remove(tmpPath)
	}
	defer cleanup()

	// Copy the contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %v", err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %v", err)
	}

	// Copy source permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	if err := os.Chmod(tmpPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to chmod temp file: %v", err)
	}

	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("failed to move temp file: %v", err)
	}

	return nil
}

// RemovePath safely removes a file or directory
func RemovePath(path string) error {
	lock := AcquireLock(path)
	defer lock.Release()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove path: %v", err)
	}

	return nil
}

// ValidatePath checks if a path is safe to use
func ValidatePath(path string) error {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for directory traversal
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path contains directory traversal")
	}

	// Check if path is absolute
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	return nil
}

// GetFileSize returns the size of a file
func GetFileSize(path string) (int64, error) {
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
