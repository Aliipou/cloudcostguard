package cmd

import (
	"fmt"

	"github.com/Aliipou/cloudcostguard/internal/api"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web dashboard and API server",
	Long: `Start an HTTP server that serves the CloudCostGuard web dashboard
and exposes a REST API for running scans and viewing results.

Examples:
  cloudcostguard serve
  cloudcostguard serve --port 3000
  cloudcostguard serve --config myconfig.yaml`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "port to listen on")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	srv := api.NewServer(configFile, servePort)
	fmt.Printf("Starting CloudCostGuard server on port %d...\n", servePort)
	return srv.Start()
}
