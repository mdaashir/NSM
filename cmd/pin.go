/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"github.com/mdaashir/NSM/utils"

	"github.com/spf13/cobra"
)

var pinCmd = &cobra.Command{
	Use:   "pin [package] [version]",
	Short: "Pin a package to a specific version",
	Long: `Pin a package to a specific version. This will update your NSM configuration
to ensure the specified package version is used in future installations.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			utils.Error("Please provide both package name and version")
			utils.Tip("Usage: nsm pin PACKAGE VERSION")
			return
		}

		packageName := args[0]
		version := args[1]

		if err := utils.PinPackage(packageName, version); err != nil {
			utils.Error("Failed to pin package: %v", err)
			return
		}

		utils.Success("Successfully pinned %s to version %s", packageName, version)
		utils.Tip("Run 'nsm list' to see all pinned packages")
	},
}

func init() {
	RootCmd.AddCommand(pinCmd)
}
