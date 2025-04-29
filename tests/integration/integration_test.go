// Package integration provides integration tests for NSM functionality
package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/mdaashir/NSM/cmd"
	"github.com/mdaashir/NSM/tests/testutils"
)

func TestWorkflowInitToRun(t *testing.T) {
	testutils.SkipIfNotNix(t)

	tmpDir, cleanup := testutils.TempDir(t)
	defer cleanup()

	testutils.WithWorkDir(t, tmpDir, func() {
		// Test init command
		stdout, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"init"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("init command failed: %v", err)
			}
		})

		testutils.AssertFileExists(t, "shell.nix")
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
		if !strings.Contains(stdout, "Created shell.nix") {
			t.Errorf("Expected success message, got: %s", stdout)
		}

		// Test add command
		output, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"add", "gcc"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("add command failed: %v", err)
			}
		})

		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
		if !strings.Contains(output, "Added package") {
			t.Errorf("Expected success message, got: %s", output)
		}

		// Verify package was added to shell.nix
		content, err := os.ReadFile("shell.nix")
		if err != nil {
			t.Fatalf("Failed to read shell.nix: %v", err)
		}
		if !strings.Contains(string(content), "gcc") {
			t.Error("Package not found in shell.nix")
		}
	})
}

func TestFlakeWorkflow(t *testing.T) {
	testutils.SkipIfNotNix(t)

	tmpDir, cleanup := testutils.TempDir(t)
	defer cleanup()

	testutils.WithWorkDir(t, tmpDir, func() {
		// Test init with flake
		stdout, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"init", "--flake"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("init command failed: %v", err)
			}
		})

		testutils.AssertFileExists(t, "flake.nix")
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
		if !strings.Contains(stdout, "Created flake.nix") {
			t.Errorf("Expected success message, got: %s", stdout)
		}

		// Test flake package operations
		stdout, stderr = testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"add", "python3", "--flake"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("add command failed: %v", err)
			}
		})

		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
		if !strings.Contains(stdout, "Added package") {
			t.Errorf("Expected success message in stdout, got: %s", stdout)
		}

		// Verify package was added to flake.nix
		content, err := os.ReadFile("flake.nix")
		if err != nil {
			t.Fatalf("Failed to read flake.nix: %v", err)
		}
		if !strings.Contains(string(content), "python3") {
			t.Error("Package not found in flake.nix")
		}
	})
}

func TestErrorHandling(t *testing.T) {
	testutils.SkipIfNotNix(t)

	tmpDir, cleanup := testutils.TempDir(t)
	defer cleanup()

	testutils.WithWorkDir(t, tmpDir, func() {
		// Test invalid package name
		_, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"add", "invalid-package-name-that-does-not-exist"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Logf("Expected error executing command: %v", err)
			}
		})

		if !strings.Contains(stderr, "package not found") {
			t.Error("Expected error message for invalid package")
		}

		// Test invalid config value
		_, stderr = testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"config", "set", "shell.format", "invalid"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Logf("Expected error executing command: %v", err)
			}
		})

		if !strings.Contains(stderr, "must be either 'shell.nix' or 'flake.nix'") {
			t.Error("Expected error message for invalid config value")
		}
	})
}

func TestConfigOperations(t *testing.T) {
	tmpDir, cleanup := testutils.TempDir(t)
	defer cleanup()

	testutils.WithWorkDir(t, tmpDir, func() {
		// Test config set
		_, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"config", "set", "shell.format", "flake.nix"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("config set command failed: %v", err)
			}
		})

		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}

		// Test config get
		output, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"config", "show"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("config show command failed: %v", err)
			}
		})

		if !strings.Contains(output, "shell.format: flake.nix") {
			t.Error("Config value not set correctly")
		}
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})
}

func TestCleanup(t *testing.T) {
	testutils.SkipIfNotNix(t)

	tmpDir, cleanup := testutils.TempDir(t)
	defer cleanup()

	testutils.WithWorkDir(t, tmpDir, func() {
		// Create test files
		testutils.CreateTestFile(t, tmpDir, "shell.nix", "# Test content")
		testutils.CreateTestFile(t, tmpDir, "flake.nix", "# Test content")

		// Test cleanup command
		_, stderr := testutils.CaptureOutput(t, func() {
			cmd.RootCmd.SetArgs([]string{"clean", "--force"})
			if err := cmd.RootCmd.Execute(); err != nil {
				t.Fatalf("clean command failed: %v", err)
			}
		})

		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}

		// Verify files were cleaned up
		testutils.AssertFileNotExists(t, "shell.nix")
		testutils.AssertFileNotExists(t, "flake.nix")
	})
}
