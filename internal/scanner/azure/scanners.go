package azure

import (
	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/scanner"
)

// NewScanners creates Azure scanners based on the requested scan type.
func NewScanners(cfg *config.Config, scanType string) ([]scanner.Scanner, error) {
	subID := ""
	if cfg.Azure != nil {
		subID = cfg.Azure.SubscriptionID
	}

	var scanners []scanner.Scanner

	switch scanType {
	case "compute":
		scanners = append(scanners, NewVMScanner(cfg, subID))
	case "storage":
		scanners = append(scanners, NewDiskScanner(cfg, subID))
	case "network":
		scanners = append(scanners, NewNetworkScanner(cfg, subID))
	case "all", "":
		scanners = append(scanners,
			NewVMScanner(cfg, subID),
			NewDiskScanner(cfg, subID),
			NewNetworkScanner(cfg, subID),
		)
	}

	return scanners, nil
}
