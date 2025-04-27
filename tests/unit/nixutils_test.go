package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
)

func TestValidatePackage(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		expected bool
	}{
		{"empty package", "", false},
		{"valid package", "gcc", true},
		{"valid package with version", "python3.9", true},
		{"valid package with hyphen", "node-red", true},
		{"invalid package with space", "gcc python", false},
		{"invalid package with special chars", "gcc$python", false},
		{"valid package with dot", "go.mod", true},
		{"valid package with underscore", "test_package", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.ValidatePackage(tt.pkg); got != tt.expected {
				t.Errorf("ValidatePackage(%q) = %v, want %v", tt.pkg, got, tt.expected)
			}
		})
	}
}

func TestCheckFlakeSupport(t *testing.T) {
	t.Run("flake supported version", func(t *testing.T) {
		// Create a mock nix command that returns version 2.4
		mockPath := testutils.CreateMockCmd(t, "nix", "nix (Nix) 2.4.0", 0)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				t.Fatal(err)
			}
		}(mockPath)

		oldPath := os.Getenv("PATH")
		err := os.Setenv("PATH", filepath.Dir(mockPath))
		if err != nil {
			return
		}
		defer func(key, value string) {
			err := os.Setenv(key, value)
			if err != nil {
				return
			}
		}("PATH", oldPath)

		if !utils.CheckFlakeSupport() {
			t.Error("Expected flake support for Nix 2.4.0")
		}
	})

	t.Run("flake unsupported version", func(t *testing.T) {
		mockPath := testutils.CreateMockCmd(t, "nix", "nix (Nix) 2.3.0", 0)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				t.Fatal(err)
			}
		}(mockPath)

		oldPath := os.Getenv("PATH")
		err := os.Setenv("PATH", filepath.Dir(mockPath))
		if err != nil {
			return
		}
		defer func(key, value string) {
			err := os.Setenv(key, value)
			if err != nil {
				return
			}
		}("PATH", oldPath)

		if utils.CheckFlakeSupport() {
			t.Error("Expected no flake support for Nix 2.3.0")
		}
	})
}

func TestExtractPackages(t *testing.T) {
	config, cleanup := testutils.CreateTestConfig(t)
	defer cleanup()

	t.Run("extract from shell.nix", func(t *testing.T) {
		content, err := os.ReadFile(config.ShellNixPath)
		if err != nil {
			t.Fatal(err)
		}

		packages := utils.ExtractShellNixPackages(string(content))
		expected := []string{"gcc", "python3"}

		if len(packages) != len(expected) {
			t.Errorf("got %d packages, want %d", len(packages), len(expected))
		}

		for i, pkg := range packages {
			if pkg != expected[i] {
				t.Errorf("package[%d] = %q, want %q", i, pkg, expected[i])
			}
		}
	})

	t.Run("extract from flake.nix", func(t *testing.T) {
		content, err := os.ReadFile(config.FlakeNixPath)
		if err != nil {
			t.Fatal(err)
		}

		packages := utils.ExtractFlakePackages(string(content))
		expected := []string{"gcc", "python3"}

		if len(packages) != len(expected) {
			t.Errorf("got %d packages, want %d", len(packages), len(expected))
		}

		for i, pkg := range packages {
			if pkg != expected[i] {
				t.Errorf("package[%d] = %q, want %q", i, pkg, expected[i])
			}
		}
	})
}

func TestGetNixVersion(t *testing.T) {
	expectedVersion := "nix (Nix) 2.4.0"
	mockPath := testutils.CreateMockCmd(t, "nix", expectedVersion, 0)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(mockPath)

	oldPath := os.Getenv("PATH")
	err := os.Setenv("PATH", filepath.Dir(mockPath))
	if err != nil {
		return
	}
	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			return
		}
	}("PATH", oldPath)

	version, err := utils.GetNixVersion()
	if err != nil {
		t.Fatalf("GetNixVersion() error = %v", err)
	}

	if version != expectedVersion {
		t.Errorf("GetNixVersion() = %q, want %q", version, expectedVersion)
	}
}

func TestGetPackageVersion(t *testing.T) {
	// Create a mock nix-env command that returns package info in JSON format
	mockOutput := `{
		"nixpkgs.gcc": {
			"name": "gcc-12.3.0",
			"version": "12.3.0",
			"system": "x86_64-linux",
			"outPath": "/nix/store/...-gcc-12.3.0"
		}
	}`
	mockPath := testutils.CreateMockCmd(t, "nix-env", mockOutput, 0)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(mockPath)

	oldPath := os.Getenv("PATH")
	err := os.Setenv("PATH", filepath.Dir(mockPath))
	if err != nil {
		return
	}
	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			return
		}
	}("PATH", oldPath)

	version, err := utils.GetPackageVersion("gcc")
	if err != nil {
		t.Fatalf("GetPackageVersion() error = %v", err)
	}

	expectedVersion := "12.3.0"
	if version != expectedVersion {
		t.Errorf("GetPackageVersion() = %q, want %q", version, expectedVersion)
	}
}
