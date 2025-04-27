// Package integration provides integration tests for NSM functionality,
// testing the interaction between different components and external dependencies.
package integration

import (
	"os"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
)

// TestPackageManagement tests package management operations
func TestPackageManagement(t *testing.T) {
	config, cleanup := testutils.CreateTestConfig(t)
	defer cleanup()

	// Create test shell.nix with secure permissions
	testShellContent := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    gcc
  ];
}`
	if err := os.WriteFile(config.ShellNixPath, []byte(testShellContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Save the current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatal(err)
		}
	}(origDir)

	// Change to test directory
	if err := os.Chdir(config.TempDir); err != nil {
		t.Fatal(err)
	}

	t.Run("package operations", func(t *testing.T) {
		// Verify initial packages
		content, err := os.ReadFile("shell.nix")
		if err != nil {
			t.Fatal(err)
		}

		initialPkgs := utils.ExtractShellNixPackages(string(content))
		if len(initialPkgs) != 2 {
			t.Errorf("Expected 2 initial packages, got %d", len(initialPkgs))
		}

		// Pin a package version
		if err := utils.PinPackage(); err != nil {
			t.Errorf("Failed to pin package: %v", err)
		}

		// Verify the pin was saved
		cfg, err := utils.LoadConfig()
		if err != nil {
			t.Fatal(err)
		}

		if version := cfg.Pins["gcc"]; version != "12.3.0" {
			t.Errorf("Pin version = %q, want 12.3.0", version)
		}
	})
}
