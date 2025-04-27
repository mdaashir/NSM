package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
)

func TestNixEnvironment(t *testing.T) {
	testDir, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Test nix shell environment detection
	inNixShell := utils.IsInNixShell()
	if os.Getenv("IN_NIX_SHELL") != "" && !inNixShell {
		t.Error("Failed to detect active nix-shell environment")
	}

	// Test nix installation check
	nixInstalled := utils.IsNixInstalled()
	if nixInstalled {
		path, err := utils.GetNixPath()
		testutils.AssertNoError(t, err)
		if path == "" {
			t.Error("Nix path is empty despite Nix being installed")
		}
	}
}

func TestFlakeManagement(t *testing.T) {
	testDir, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Create test flake
	flakePath := filepath.Join(testDir, "flake.nix")
	testutils.CreateTestFlakeNix(t, testDir, []string{"git", "go"})

	// Test flake initialization
	err := utils.InitFlake(testDir)
	testutils.AssertNoError(t, err)

	// Test flake lock existence
	lockPath := filepath.Join(testDir, "flake.lock")
	if !utils.FileExists(lockPath) {
		t.Error("Flake lock file was not created")
	}

	// Test flake update
	err = utils.UpdateFlake(testDir)
	testutils.AssertNoError(t, err)

	// Test invalid flake handling
	invalidDir := filepath.Join(testDir, "invalid")
	err = os.MkdirAll(invalidDir, 0755)
	testutils.AssertNoError(t, err)

	err = utils.InitFlake(invalidDir)
	testutils.AssertError(t, err)
}

func TestPackageManagement(t *testing.T) {
	testDir, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Test package list parsing
	packages := []string{"git", "go", "nodejs"}
	shellNixPath := filepath.Join(testDir, "shell.nix")
	testutils.CreateTestShellNix(t, testDir, packages)

	parsed, err := utils.ParsePackageList(shellNixPath)
	testutils.AssertNoError(t, err)

	if len(parsed) != len(packages) {
		t.Errorf("Expected %d packages, got %d", len(packages), len(parsed))
	}

	for _, pkg := range packages {
		found := false
		for _, p := range parsed {
			if p == pkg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Package %s not found in parsed list", pkg)
		}
	}

	// Test invalid package list
	invalidPath := filepath.Join(testDir, "invalid.nix")
	err = os.WriteFile(invalidPath, []byte("invalid nix content"), 0644)
	testutils.AssertNoError(t, err)

	_, err = utils.ParsePackageList(invalidPath)
	testutils.AssertError(t, err)
}

func TestShellEnvironment(t *testing.T) {
	testDir, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Test shell.nix generation
	packages := []string{"git", "go"}
	err := utils.GenerateShellNix(testDir, packages)
	testutils.AssertNoError(t, err)

	shellNixPath := filepath.Join(testDir, "shell.nix")
	if !utils.FileExists(shellNixPath) {
		t.Error("shell.nix was not generated")
	}

	// Verify shell.nix content
	content, err := os.ReadFile(shellNixPath)
	testutils.AssertNoError(t, err)

	for _, pkg := range packages {
		if !strings.Contains(string(content), pkg) {
			t.Errorf("Generated shell.nix does not contain package %s", pkg)
		}
	}

	// Test shell environment activation
	env, err := utils.GetNixShellEnv(testDir)
	testutils.AssertNoError(t, err)

	requiredVars := []string{"PATH", "NIX_PATH", "NIX_PROFILES"}
	for _, v := range requiredVars {
		if _, exists := env[v]; !exists {
			t.Errorf("Required environment variable %s not found", v)
		}
	}
}

func TestNixCache(t *testing.T) {
	testDir, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Test cache directory management
	cacheDir, err := utils.GetNixCacheDir()
	testutils.AssertNoError(t, err)

	err = utils.CleanNixCache()
	testutils.AssertNoError(t, err)

	// Verify cache directory is empty
	entries, err := os.ReadDir(cacheDir)
	testutils.AssertNoError(t, err)
	if len(entries) > 0 {
		t.Error("Cache directory not empty after cleaning")
	}

	// Test cache invalidation
	err = utils.InvalidateNixCache()
	testutils.AssertNoError(t, err)
}

func TestNixProfile(t *testing.T) {
	testDir, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Test profile management
	profile, err := utils.GetCurrentProfile()
	testutils.AssertNoError(t, err)

	if profile == "" {
		t.Error("Current profile path is empty")
	}

	// Test profile generation list
	gens, err := utils.ListProfileGenerations()
	testutils.AssertNoError(t, err)

	if len(gens) == 0 {
		t.Error("No profile generations found")
	}

	// Test profile rollback
	err = utils.RollbackProfile()
	if err != nil {
		// Rollback might fail if there's only one generation
		if !strings.Contains(err.Error(), "no generations to roll back to") {
			t.Error("Unexpected rollback error:", err)
		}
	}
}
