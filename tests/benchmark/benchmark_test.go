// Package benchmark provides performance benchmarking tests for NSM operations
package benchmark

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/cmd"
	"github.com/mdaashir/NSM/tests/testutils"
)

func BenchmarkInitCommand(b *testing.B) {
	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		os.RemoveAll(filepath.Join(tmpDir, "shell.nix"))
		os.RemoveAll(filepath.Join(tmpDir, "flake.nix"))
		b.StartTimer()

		cmd.RootCmd.SetArgs([]string{"init"})
		if err := cmd.RootCmd.Execute(); err != nil {
			b.Fatalf("init command failed: %v", err)
		}
	}
}

func BenchmarkAddPackage(b *testing.B) {
	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	// Setup initial shell.nix
	cmd.RootCmd.SetArgs([]string{"init"})
	if err := cmd.RootCmd.Execute(); err != nil {
		b.Fatalf("init setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.RootCmd.SetArgs([]string{"add", "gcc"})
		if err := cmd.RootCmd.Execute(); err != nil {
			b.Fatalf("add command failed: %v", err)
		}
	}
}

func BenchmarkPackageSearch(b *testing.B) {
	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.RootCmd.SetArgs([]string{"search", "python"})
		if err := cmd.RootCmd.Execute(); err != nil {
			b.Fatalf("search command failed: %v", err)
		}
	}
}

func BenchmarkConfigOperations(b *testing.B) {
	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	b.Run("ConfigSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"config", "set", "shell.format", "flake.nix"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("config set failed: %v", err)
			}
		}
	})

	b.Run("ConfigGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"config", "get", "shell.format"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("config get failed: %v", err)
			}
		}
	})
}

func BenchmarkFlakeOperations(b *testing.B) {
	if !CheckFlakeSupport() {
		b.Skip("Flakes not supported")
	}

	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	b.Run("FlakeInit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			os.RemoveAll(filepath.Join(tmpDir, "flake.nix"))
			b.StartTimer()

			cmd.RootCmd.SetArgs([]string{"init", "--flake"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("flake init failed: %v", err)
			}
		}
	})

	b.Run("FlakeUpdate", func(b *testing.B) {
		// Setup flake.nix first
		cmd.RootCmd.SetArgs([]string{"init", "--flake"})
		if err := cmd.RootCmd.Execute(); err != nil {
			b.Fatalf("flake setup failed: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"upgrade"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("flake update failed: %v", err)
			}
		}
	})
}

func BenchmarkParallelOperations(b *testing.B) {
	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	// Setup initial files
	cmd.RootCmd.SetArgs([]string{"init"})
	if err := cmd.RootCmd.Execute(); err != nil {
		b.Fatalf("init setup failed: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cmd.RootCmd.SetArgs([]string{"add", "gcc"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("parallel operation failed: %v", err)
			}
		}
	})
}

func BenchmarkLargeOperations(b *testing.B) {
	tmpDir, cleanup := testutils.TempDir(b)
	defer cleanup()

	// Create large shell.nix with many packages
	var packages []string
	for i := 0; i < 100; i++ {
		packages = append(packages, "gcc", "python3", "nodejs", "go")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Reset shell.nix
		cmd.RootCmd.SetArgs([]string{"init"})
		if err := cmd.RootCmd.Execute(); err != nil {
			b.Fatalf("init failed: %v", err)
		}

		b.StartTimer()
		for _, pkg := range packages {
			cmd.RootCmd.SetArgs([]string{"add", pkg})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("large operation failed: %v", err)
			}
		}
	}
}
