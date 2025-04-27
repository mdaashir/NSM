package benchmark

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/viper"
)

// BenchmarkPackageOperations benchmarks common package operations
func BenchmarkPackageOperations(b *testing.B) {
	config, cleanup := testutils.CreateBenchConfig(b)
	defer cleanup()

	// Large shell.nix content for testing
	largeShellContent := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    gcc
    python3
    nodejs
    go
    rust
    cargo
    git
    vim
    vscode
    docker
    kubernetes-helm
    terraform
    ansible
    nginx
    postgresql
    redis
    mongodb
    mysql
    php
    ruby
  ];
}`

	if err := os.WriteFile(filepath.Join(config.TempDir, "large-shell.nix"), []byte(largeShellContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.Run("ExtractPackages", func(b *testing.B) {
		content, err := os.ReadFile(filepath.Join(config.TempDir, "large-shell.nix"))
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			utils.ExtractShellNixPackages(string(content))
		}
	})

	b.Run("ValidatePackage", func(b *testing.B) {
		packages := []string{
			"gcc", "python3", "invalid package",
			"test-package", "package_with_underscore",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, pkg := range packages {
				utils.ValidatePackage(pkg)
			}
		}
	})
}

// BenchmarkFileOperations benchmarks file-related operations
func BenchmarkFileOperations(b *testing.B) {
	dir := testutils.CreateBenchTempDir(b)
	defer os.RemoveAll(dir)

	// Create test files
	files := []string{"test1.txt", "test2.txt", "test3.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("test content"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("FileExists", func(b *testing.B) {
		path := filepath.Join(dir, "test1.txt")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			utils.FileExists(path)
		}
	})

	b.Run("BackupFile", func(b *testing.B) {
		path := filepath.Join(dir, "test2.txt")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			utils.BackupFile(path)
			os.Remove(path + ".backup")
		}
	})
}

// BenchmarkConfigOperations benchmarks configuration operations
func BenchmarkConfigOperations(b *testing.B) {
	dir := testutils.CreateBenchTempDir(b)
	defer os.RemoveAll(dir)

	configPath := filepath.Join(dir, "config.yaml")
	viper.SetConfigFile(configPath)
	viper.Set("channel.url", "nixos-unstable")
	viper.Set("shell.format", "shell.nix")
	viper.Set("default.packages", []string{"gcc", "python3"})
	viper.Set("config_version", "1.0.0")

	if err := viper.WriteConfig(); err != nil {
		b.Fatal(err)
	}

	b.Run("LoadConfig", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := utils.LoadConfig()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ValidateConfig", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			utils.ValidateConfig()
		}
	})
}

// BenchmarkLogging benchmarks logging operations
func BenchmarkLogging(b *testing.B) {
	utils.ConfigureLogger(true, false)

	messages := []struct {
		level   string
		message string
		logFunc func(string, ...interface{})
	}{
		{"debug", "Debug message", utils.Debug},
		{"info", "Info message", utils.Info},
		{"success", "Success message", utils.Success},
		{"warning", "Warning message", utils.Warn},
		{"error", "Error message", utils.Error},
		{"tip", "Tip message", utils.Tip},
	}

	for _, m := range messages {
		b.Run("Log"+m.level, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m.logFunc(m.message)
			}
		})
	}
}
