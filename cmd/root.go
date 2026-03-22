package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	outputFmt  string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "cloudcostguard",
	Short: "Cloud cost optimization engine",
	Long: `CloudCostGuard scans your AWS and Azure infrastructure to find
wasted resources and recommends actions that save real money.

It detects idle VMs, unattached disks, oversized instances, unused
load balancers, and more — then generates a report with exact
dollar savings per resource.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default: ~/.cloudcostguard.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "output format: table, json, csv")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
