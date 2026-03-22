package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/pricing"
)

// DiskScanner detects unattached Azure managed disks.
type DiskScanner struct {
	cfg            *config.Config
	subscriptionID string
}

func NewDiskScanner(cfg *config.Config, subscriptionID string) *DiskScanner {
	return &DiskScanner{cfg: cfg, subscriptionID: subscriptionID}
}

func (s *DiskScanner) Name() string            { return "azure-disk" }
func (s *DiskScanner) Category() model.Category { return model.CategoryStorage }

func (s *DiskScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	disks, err := s.listUnattachedDisks(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing unattached Azure disks: %w", err)
	}

	var findings []model.Finding

	for _, d := range disks {
		monthlyCost := pricing.AzureDiskMonthlyCost(d.SKU, d.SizeGB)

		f := model.Finding{
			ID:             fmt.Sprintf("azure-disk-unattached-%s", d.ResourceID),
			Provider:       "azure",
			AccountID:      s.subscriptionID,
			Region:         d.Location,
			Category:       model.CategoryStorage,
			ResourceType:   "microsoft.compute/disks",
			ResourceID:     d.ResourceID,
			ResourceName:   d.Name,
			Severity:       model.SeverityFromSavings(monthlyCost),
			Title:          fmt.Sprintf("Unattached Azure disk: %s (%dGB %s)", d.Name, d.SizeGB, d.SKU),
			Description:    fmt.Sprintf("Managed disk %s (%dGB, %s) in %s is not attached to any VM, costing $%.2f/month.", d.Name, d.SizeGB, d.SKU, d.Location, monthlyCost),
			CurrentCost:    monthlyCost,
			ProjectedCost:  0,
			MonthlySavings: monthlyCost,
			AnnualSavings:  monthlyCost * 12,
			Recommendation: "Create a snapshot for backup, then delete this disk. Snapshots cost significantly less than managed disks.",
			Effort:         "low",
			Risk:           "low",
			Tags:           d.Tags,
			DetectedAt:     time.Now().UTC(),
		}
		findings = append(findings, f)
	}

	return findings, nil
}

type azureDisk struct {
	ResourceID string
	Name       string
	SKU        string
	SizeGB     int
	Location   string
	Tags       map[string]string
}

func (s *DiskScanner) listUnattachedDisks(ctx context.Context) ([]azureDisk, error) {
	return nil, nil
}
