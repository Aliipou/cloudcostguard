package aws

import (
	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/scanner"
)

// NewScanners creates AWS scanners based on the requested scan type.
func NewScanners(cfg *config.Config, scanType string) ([]scanner.Scanner, error) {
	regions := []string{"us-east-1"}
	if cfg.AWS != nil && len(cfg.AWS.Regions) > 0 {
		regions = cfg.AWS.Regions
	}

	var scanners []scanner.Scanner

	for _, region := range regions {
		switch scanType {
		case "compute":
			scanners = append(scanners, NewEC2Scanner(cfg, region))
		case "storage":
			scanners = append(scanners, NewEBSScanner(cfg, region))
			scanners = append(scanners, NewS3Scanner(cfg, region))
		case "network":
			scanners = append(scanners, NewELBScanner(cfg, region))
		case "database":
			scanners = append(scanners, NewRDSScanner(cfg, region))
		case "all", "":
			scanners = append(scanners,
				NewEC2Scanner(cfg, region),
				NewEBSScanner(cfg, region),
				NewS3Scanner(cfg, region),
				NewELBScanner(cfg, region),
				NewRDSScanner(cfg, region),
			)
		}
	}

	return scanners, nil
}
