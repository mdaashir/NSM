// Package testutils provides testing utilities for NSM
package testutils

import (
	"fmt"
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

	// Create mock shell.nix with secure permissions
	shellContent := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    gcc
    python3
  ];
}`
	if err := os.WriteFile(config.ShellNixPath, []byte(shellContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create mock flake.nix with secure permissions
	flakeContent := `{
  description = "Test environment";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.mkShell {
      buildInputs = [ gcc python3 ];
    };
  };
}`
	if err := os.WriteFile(config.FlakeNixPath, []byte(flakeContent), 0600); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			return
		}
	}

	return config, cleanup
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
	if err := os.WriteFile(mockPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return mockPath
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
	if err := os.WriteFile(config.ShellNixPath, []byte(shellContent), 0600); err != nil {
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
	if err := os.WriteFile(config.FlakeNixPath, []byte(flakeContent), 0600); err != nil {
		b.Fatal(err)
	}

	cleanup := func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			return
		}
	}

	return config, cleanup
}
