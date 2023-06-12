package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version string = "?.?.?"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the CLI app",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Bulut version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
