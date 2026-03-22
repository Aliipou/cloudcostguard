package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cloudcostguard %s (commit: %s, built: %s)\n", Version, CommitSHA, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
