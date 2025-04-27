// Package testutils provides testing utilities for NSM
package testutils

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TempDir creates a temporary directory and returns its path along with a cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "nsm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Failed to cleanup temp dir: %v", err)
		}
	}
}

// CreateTestFile creates a file with given content in the specified directory
func CreateTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return path
}

// CaptureOutput captures stdout/stderr during test execution
func CaptureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	done := make(chan struct{})
	outC := make(chan string)
	errC := make(chan string)

	go func() {
		outBytes, err := io.ReadAll(rOut)
		if err != nil {
			t.Errorf("Failed to read stdout: %v", err)
		}
		outC <- string(outBytes)
	}()

	go func() {
		errBytes, err := io.ReadAll(rErr)
		if err != nil {
			t.Errorf("Failed to read stderr: %v", err)
		}
		errC <- string(errBytes)
	}()

	fn()

	wOut.Close()
	wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout = <-outC
	stderr = <-errC

	close(done)
	return stdout, stderr
}

// SkipIfNotNix skips tests that require Nix if it's not installed
func SkipIfNotNix(t *testing.T) {
	t.Helper()

	if _, err := os.Stat("/nix"); os.IsNotExist(err) {
		t.Skip("Nix is not installed")
	}
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// NormalizePath normalizes a file path for the current OS
func NormalizePath(path string) string {
	if IsWindows() {
		return filepath.FromSlash(path)
	}
	return path
}

// AssertFileContent checks if a file contains expected content
func AssertFileContent(t *testing.T, path string, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	if string(content) != expected {
		t.Errorf("File content mismatch.\nExpected:\n%s\nGot:\n%s", expected, string(content))
	}
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", path)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected file %s to not exist", path)
	}
}

// AssertDirExists checks if a directory exists
func AssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected directory %s to exist", path)
		return
	}

	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory", path)
	}
}

// WithWorkDir changes the working directory for test execution
func WithWorkDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	fn()
}
