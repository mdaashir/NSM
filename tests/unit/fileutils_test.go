package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
)

func TestFileExists(t *testing.T) {
	dir := testutils.CreateTempDir(t)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(dir)

	testCases := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "existing file",
			setup: func() string {
				path := filepath.Join(dir, "test.txt")
				err := os.WriteFile(path, []byte("test"), 0600)
				if err != nil {
					return ""
				}
				return path
			},
			expected: true,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(dir, "nonexistent.txt")
			},
			expected: false,
		},
		{
			name: "directory",
			setup: func() string {
				path := filepath.Join(dir, "testdir")
				err := os.Mkdir(path, 0755)
				if err != nil {
					return ""
				}
				return path
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup()
			if got := utils.FileExists(path); got != tc.expected {
				t.Errorf("FileExists(%q) = %v, want %v", path, got, tc.expected)
			}
		})
	}
}

func TestBackupFile(t *testing.T) {
	dir := testutils.CreateTempDir(t)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(dir)

	t.Run("successful backup", func(t *testing.T) {
		// Create an original file
		originalPath := filepath.Join(dir, "original.txt")
		originalContent := "test content"
		if err := os.WriteFile(originalPath, []byte(originalContent), 0600); err != nil {
			t.Fatal(err)
		}

		// Create backup
		if err := utils.BackupFile(originalPath); err != nil {
			t.Fatalf("BackupFile() error = %v", err)
		}

		// Verify backup file
		backupPath := originalPath + ".backup"
		if !utils.FileExists(backupPath) {
			t.Error("Backup file was not created")
		}

		content, err := os.ReadFile(backupPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(content) != originalContent {
			t.Errorf("Backup content = %q, want %q", string(content), originalContent)
		}
	})

	t.Run("backup nonexistent file", func(t *testing.T) {
		nonexistentPath := filepath.Join(dir, "nonexistent.txt")
		if err := utils.BackupFile(nonexistentPath); err == nil {
			t.Error("Expected error when backing up nonexistent file")
		}
	})
}

func TestGetProjectConfigType(t *testing.T) {
	dir := testutils.CreateTempDir(t)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(dir)

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
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name           string
		files          map[string]bool // filename -> should exist
		expectedConfig string
	}{
		{
			name:           "no config files",
			files:          map[string]bool{},
			expectedConfig: "",
		},
		{
			name: "only shell.nix",
			files: map[string]bool{
				"shell.nix": true,
			},
			expectedConfig: "shell.nix",
		},
		{
			name: "only flake.nix",
			files: map[string]bool{
				"flake.nix": true,
			},
			expectedConfig: "flake.nix",
		},
		{
			name: "both config files",
			files: map[string]bool{
				"shell.nix": true,
				"flake.nix": true,
			},
			expectedConfig: "shell.nix", // shell.nix takes precedence
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up previous files
			err := os.Remove("shell.nix")
			if err != nil {
				return
			}
			err = os.Remove("flake.nix")
			if err != nil {
				return
			}

			// Create test files
			for file, shouldExist := range tc.files {
				if shouldExist {
					if err := os.WriteFile(file, []byte("test"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}

			if got := utils.GetProjectConfigType(); got != tc.expectedConfig {
				t.Errorf("GetProjectConfigType() = %q, want %q", got, tc.expectedConfig)
			}
		})
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Save original XDG_CONFIG_HOME
	origXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			return
		}
	}("XDG_CONFIG_HOME", origXdgConfig)

	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		dir := testutils.CreateTempDir(t)
		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {
				t.Fatal(err)
			}
		}(dir)

		err := os.Setenv("XDG_CONFIG_HOME", dir)
		if err != nil {
			return
		}

		configDir, err := utils.EnsureConfigDir()
		if err != nil {
			t.Fatalf("EnsureConfigDir() error = %v", err)
		}

		expectedPath := filepath.Join(dir, "NSM")
		if configDir != expectedPath {
			t.Errorf("EnsureConfigDir() = %q, want %q", configDir, expectedPath)
		}

		testutils.AssertDirExists(t, configDir)
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		err := os.Setenv("XDG_CONFIG_HOME", "")
		if err != nil {
			return
		}

		configDir, err := utils.EnsureConfigDir()
		if err != nil {
			t.Fatalf("EnsureConfigDir() error = %v", err)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}

		expectedPath := filepath.Join(home, ".config", "NSM")
		if configDir != expectedPath {
			t.Errorf("EnsureConfigDir() = %q, want %q", configDir, expectedPath)
		}

		testutils.AssertDirExists(t, configDir)
	})
}
