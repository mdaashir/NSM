// Package integration provides integration tests for NSM functionality,
// testing the interaction between different components and external dependencies.
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
)

// TestPackageManagement tests package management operations
func TestPackageManagement(t *testing.T) {
	config, cleanup := testutils.CreateTestConfig(t)
	defer cleanup()

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

		// Mock the nix-env command to return package info
		mockPath := testutils.CreateMockCmd(t, "nix-env", `{
			"nixpkgs.gcc": {
				"name": "gcc-12.3.0",
				"version": "12.3.0",
				"system": "x86_64-linux",
				"outPath": "/nix/store/...-gcc-12.3.0"
			}
		}`, 0)
		defer os.Remove(mockPath)

		// Update PATH to include mock binary
		oldPath := os.Getenv("PATH")
		mockDir := filepath.Dir(mockPath)
		if err := os.Chmod(mockPath, 0755); err != nil {
			t.Fatal(err)
		}
		newPath := mockDir + string(os.PathListSeparator) + oldPath
		if err := os.Setenv("PATH", newPath); err != nil {
			t.Fatal(err)
		}
		defer os.Setenv("PATH", oldPath)

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
