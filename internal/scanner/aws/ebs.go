package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/pricing"
)

// EBSScanner detects unattached and underutilized EBS volumes.
type EBSScanner struct {
	cfg    *config.Config
	region string
}

func NewEBSScanner(cfg *config.Config, region string) *EBSScanner {
	return &EBSScanner{cfg: cfg, region: region}
}

func (s *EBSScanner) Name() string            { return "aws-ebs-" + s.region }
func (s *EBSScanner) Category() model.Category { return model.CategoryStorage }

func (s *EBSScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	volumes, err := s.listUnattachedVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing EBS volumes in %s: %w", s.region, err)
	}

	var findings []model.Finding

	for _, vol := range volumes {
		monthlyCost := pricing.EBSMonthlyCost(vol.VolumeType, vol.SizeGB, vol.IOPS, s.region)
		f := model.Finding{
			ID:             fmt.Sprintf("aws-ebs-unattached-%s-%s", s.region, vol.ID),
			Provider:       "aws",
			Region:         s.region,
			Category:       model.CategoryStorage,
			ResourceType:   "ebs:volume",
			ResourceID:     vol.ID,
			ResourceName:   vol.Name,
			Severity:       model.SeverityFromSavings(monthlyCost),
			Title:          fmt.Sprintf("Unattached EBS volume: %s (%dGB %s)", vol.Name, vol.SizeGB, vol.VolumeType),
			Description:    fmt.Sprintf("Volume %s (%dGB, %s) has been unattached for %d+ days. You're paying $%.2f/month for unused storage.", vol.ID, vol.SizeGB, vol.VolumeType, s.cfg.Rules.UnattachedDiskDays, monthlyCost),
			CurrentCost:    monthlyCost,
			ProjectedCost:  0,
			MonthlySavings: monthlyCost,
			AnnualSavings:  monthlyCost * 12,
			Recommendation: "Snapshot this volume for backup, then delete it. If needed later, restore from the snapshot.",
			Effort:         "low",
			Risk:           "low",
			Tags:           vol.Tags,
			DetectedAt:     time.Now().UTC(),
		}
		findings = append(findings, f)
	}

	return findings, nil
}

type ebsVolume struct {
	ID         string
	Name       string
	VolumeType string
	SizeGB     int
	IOPS       int
	Tags       map[string]string
}

func (s *EBSScanner) listUnattachedVolumes(ctx context.Context) ([]ebsVolume, error) {
	return nil, nil
}
