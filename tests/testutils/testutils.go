package testutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	TempDir      string
	ShellNixPath string
	FlakeNixPath string
}

// CreateTempDir creates a temporary directory and returns its path
func CreateTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "nsm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return dir
}

// CreateTestConfig creates a test configuration with mock files
func CreateTestConfig(t *testing.T) (*TestConfig, func()) {
	t.Helper()
	tempDir := CreateTempDir(t)

	config := &TestConfig{
		TempDir:      tempDir,
		ShellNixPath: filepath.Join(tempDir, "shell.nix"),
		FlakeNixPath: filepath.Join(tempDir, "flake.nix"),
	}

	// Create mock shell.nix
	shellContent := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    gcc
    python3
  ];
}`
	if err := os.WriteFile(config.ShellNixPath, []byte(shellContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create mock flake.nix
	flakeContent := `{
  description = "Test environment";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.mkShell {
      buildInputs = [ gcc python3 ];
    };
  };
}`
	if err := os.WriteFile(config.FlakeNixPath, []byte(flakeContent), 0644); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return config, cleanup
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", path)
	}
}

// AssertFileContains checks if a file contains expected content
func AssertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	if string(content) != expected {
		t.Errorf("File content mismatch.\nGot:\n%s\nWant:\n%s", string(content), expected)
	}
}

// AssertDirExists checks if a directory exists
func AssertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected directory %s to exist", path)
	}
	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory", path)
	}
}

// CreateMockCmd creates a mock command for testing
func CreateMockCmd(t *testing.T, name, output string, exitCode int) string {
	t.Helper()
	content := fmt.Sprintf(`#!/bin/sh
echo "%s"
exit %d
`, output, exitCode)

	mockPath := filepath.Join(os.TempDir(), name)
	if err := os.WriteFile(mockPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	return mockPath
}

// CaptureOutput captures stdout/stderr output during test execution
func CaptureOutput(f func()) (string, string) {
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

// CreateBenchTempDir creates a temporary directory for benchmarks
func CreateBenchTempDir(b *testing.B) string {
	dir, err := os.MkdirTemp("", "nsm-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	return dir
}

// CreateBenchConfig creates a test configuration for benchmarks
func CreateBenchConfig(b *testing.B) (*TestConfig, func()) {
	tempDir := CreateBenchTempDir(b)

	config := &TestConfig{
		TempDir:      tempDir,
		ShellNixPath: filepath.Join(tempDir, "shell.nix"),
		FlakeNixPath: filepath.Join(tempDir, "flake.nix"),
	}

	// Create mock shell.nix
	shellContent := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    gcc
    python3
  ];
}`
	if err := os.WriteFile(config.ShellNixPath, []byte(shellContent), 0644); err != nil {
		b.Fatal(err)
	}

	// Create mock flake.nix
	flakeContent := `{
  description = "Test environment";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.mkShell {
      buildInputs = [ gcc python3 ];
    };
  };
}`
	if err := os.WriteFile(config.FlakeNixPath, []byte(flakeContent), 0644); err != nil {
		b.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return config, cleanup
}
