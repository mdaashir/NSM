/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new nix environment",
	Run: func(cmd *cobra.Command, args []string) {
		content := `{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    # Add your packages here
  ];
}`

		err := os.WriteFile("shell.nix", []byte(content), 0644)
		if err != nil {
			fmt.Println("Error creating shell.nix:", err)
			return
		}

		fmt.Println("✅ Initialized new nix environment (shell.nix created)!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
