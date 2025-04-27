/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>

*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// freezeCmd represents the freeze command
var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "Freeze the nixpkgs version in use",
	Run: func(cmd *cobra.Command, args []string) {
		// Run `nix-instantiate --eval -E "with import <nixpkgs> {}; builtins.currentSystem"`
		c := exec.Command("nix-instantiate", "--eval", "-E", "with import <nixpkgs> {}; builtins.currentSystem")
		output, err := c.Output()
		if err != nil {
			fmt.Println("❌ Error freezing nixpkgs:", err)
			return
		}

		system := strings.TrimSpace(string(output))
		lockFile := "nixpkgs.json"

		// Write lock file
		err = os.WriteFile(lockFile, []byte(fmt.Sprintf(`{"nixpkgs_commit": "%s"}`, system)), 0644)
		if err != nil {
			fmt.Println("❌ Error writing nixpkgs.json:", err)
			return
		}

		fmt.Printf("✅ Frozen nixpkgs version in %s\n", lockFile)
	},
}

func init() {
	rootCmd.AddCommand(freezeCmd)
}
