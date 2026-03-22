package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/engine"
	"github.com/Aliipou/cloudcostguard/internal/report"
	"github.com/spf13/cobra"
)

var (
	provider    string
	scanType    string
	minSavings  float64
	dryRun      bool
	concurrency int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan cloud infrastructure for cost optimization opportunities",
	Long: `Scan connects to your cloud provider and analyzes resources
for waste: idle VMs, unattached disks, oversized instances,
unused load balancers, and more.

Examples:
  cloudcostguard scan --provider aws
  cloudcostguard scan --provider azure --type compute
  cloudcostguard scan --provider aws --min-savings 50 --output json`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVarP(&provider, "provider", "p", "", "cloud provider: aws, azure (required)")
	scanCmd.Flags().StringVarP(&scanType, "type", "t", "all", "resource type: compute, storage, network, database, all")
	scanCmd.Flags().Float64Var(&minSavings, "min-savings", 0, "minimum monthly savings threshold in USD")
	scanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would be scanned without connecting")
	scanCmd.Flags().IntVar(&concurrency, "concurrency", 5, "max concurrent API calls")
	scanCmd.MarkFlagRequired("provider")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if provider != "" {
		cfg.Provider = provider
	}

	if dryRun {
		fmt.Println("Dry run mode — would scan the following:")
		fmt.Printf("  Provider:    %s\n", cfg.Provider)
		fmt.Printf("  Type:        %s\n", scanType)
		fmt.Printf("  Min savings: $%.2f/month\n", minSavings)
		return nil
	}

	eng, err := engine.New(cfg, engine.Options{
		ScanType:    scanType,
		MinSavings:  minSavings,
		Concurrency: concurrency,
		Verbose:     verbose,
	})
	if err != nil {
		return fmt.Errorf("initializing scan engine: %w", err)
	}

	fmt.Printf("Scanning %s resources...\n", cfg.Provider)
	start := time.Now()

	findings, err := eng.Scan(ctx)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	duration := time.Since(start)
	fmt.Printf("Scan complete in %s — found %d optimization opportunities\n\n", duration.Round(time.Millisecond), len(findings))

	switch outputFmt {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report.ToJSON(findings))
	case "csv":
		return report.WriteCSV(os.Stdout, findings)
	default:
		report.WriteTable(os.Stdout, findings, minSavings)
	}

	return nil
}
