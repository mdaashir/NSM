// Package benchmark provides performance benchmarking tests for NSM operations
package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/cmd"
	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
)

func BenchmarkInitCommand(b *testing.B) {
	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"init"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("init command failed: %v", err)
			}
		}
	})
}

func BenchmarkAddPackage(b *testing.B) {
	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
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
	})
}

func BenchmarkPackageSearch(b *testing.B) {
	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"search", "python"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("search command failed: %v", err)
			}
		}
	})
}

func BenchmarkConfigOperations(b *testing.B) {
	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"config", "set", "channel.url", "nixos-unstable"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("config set failed: %v", err)
			}
		}
	})
}

func BenchmarkFlakeOperations(b *testing.B) {
	if !utils.CheckFlakeSupport() {
		b.Skip("Flakes not supported")
	}

	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
		// Initialize flake first
		cmd.RootCmd.SetArgs([]string{"init", "--flake"})
		if err := cmd.RootCmd.Execute(); err != nil {
			b.Fatalf("flake init failed: %v", err)
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
	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				cmd.RootCmd.SetArgs([]string{"init"})
				if err := cmd.RootCmd.Execute(); err != nil {
					b.Fatalf("parallel operation failed: %v", err)
				}
			}
		})
	})
}

func BenchmarkLargeOperations(b *testing.B) {
	tmpDir, cleanup := testutils.BenchTempDir(b)
	defer cleanup()

	testutils.WithWorkDir(b, tmpDir, func() {
		// Create a large number of test files
		for i := 0; i < 100; i++ {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.nix", i))
			if err := os.WriteFile(testFile, []byte("# Test content"), 0600); err != nil {
				b.Fatalf("Failed to create test file: %v", err)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.RootCmd.SetArgs([]string{"clean", "--all"})
			if err := cmd.RootCmd.Execute(); err != nil {
				b.Fatalf("large operation failed: %v", err)
			}
		}
	})
}
