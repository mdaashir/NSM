/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/mdaashir/NSM/utils"

	"github.com/spf13/cobra"
)

var pinCmd = &cobra.Command{
	Use:   "pin [package] [version]",
	Short: "Pin a package to a specific version",
	Long: `Pin a package to a specific version. This will update your NSM configuration
to ensure the specified package version is used in future installations.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := strings.TrimSpace(args[0])
		version := strings.TrimSpace(args[1])

		if pkg == "" || version == "" {
			return fmt.Errorf("package name and version cannot be empty")
		}

		// Validate version format
		if !strings.HasPrefix(version, "v") && !strings.Contains(version, ".") {
			utils.Warn("Version format might be invalid. Consider using semantic versioning (e.g., v1.0.0 or 1.0.0)")
		}

		err := utils.PinPackage(pkg, version)
		if err != nil {
			return fmt.Errorf("failed to pin package: %v", err)
		}

		utils.Success("Successfully pinned %s to version %s", pkg, version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pinCmd)
}
