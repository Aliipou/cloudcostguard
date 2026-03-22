package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

// S3Scanner detects S3 buckets with cost optimization opportunities.
type S3Scanner struct {
	cfg    *config.Config
	region string
}

func NewS3Scanner(cfg *config.Config, region string) *S3Scanner {
	return &S3Scanner{cfg: cfg, region: region}
}

func (s *S3Scanner) Name() string            { return "aws-s3-" + s.region }
func (s *S3Scanner) Category() model.Category { return model.CategoryStorage }

func (s *S3Scanner) Scan(ctx context.Context) ([]model.Finding, error) {
	buckets, err := s.listBucketsWithMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing S3 buckets: %w", err)
	}

	var findings []model.Finding

	for _, b := range buckets {
		// Check for buckets without lifecycle policies on large storage
		if !b.HasLifecyclePolicy && b.SizeBytes > 100*1024*1024*1024 { // >100GB
			storageCostPerGB := 0.023 // S3 Standard per GB/month
			currentCost := float64(b.SizeBytes) / (1024 * 1024 * 1024) * storageCostPerGB
			// Intelligent-Tiering saves ~40% on average
			savings := currentCost * 0.40

			f := model.Finding{
				ID:             fmt.Sprintf("aws-s3-lifecycle-%s-%s", s.region, b.Name),
				Provider:       "aws",
				Region:         s.region,
				Category:       model.CategoryStorage,
				ResourceType:   "s3:bucket",
				ResourceID:     b.Name,
				ResourceName:   b.Name,
				Severity:       model.SeverityFromSavings(savings),
				Title:          fmt.Sprintf("S3 bucket without lifecycle policy: %s (%.1f GB)", b.Name, float64(b.SizeBytes)/(1024*1024*1024)),
				Description:    fmt.Sprintf("Bucket %s stores %.1f GB without lifecycle rules. Adding Intelligent-Tiering or transitioning old objects to Glacier could save ~$%.2f/month.", b.Name, float64(b.SizeBytes)/(1024*1024*1024), savings),
				CurrentCost:    currentCost,
				ProjectedCost:  currentCost - savings,
				MonthlySavings: savings,
				AnnualSavings:  savings * 12,
				Recommendation: "Enable S3 Intelligent-Tiering for automatic cost optimization, or add lifecycle rules to transition objects older than 90 days to Glacier.",
				Effort:         "low",
				Risk:           "low",
				DetectedAt:     time.Now().UTC(),
			}
			findings = append(findings, f)
		}

		// Check for buckets with no versioning cleanup
		if b.HasVersioning && b.OldVersionBytes > 50*1024*1024*1024 { // >50GB old versions
			storageCostPerGB := 0.023
			oldVersionCost := float64(b.OldVersionBytes) / (1024 * 1024 * 1024) * storageCostPerGB

			f := model.Finding{
				ID:             fmt.Sprintf("aws-s3-versions-%s-%s", s.region, b.Name),
				Provider:       "aws",
				Region:         s.region,
				Category:       model.CategoryStorage,
				ResourceType:   "s3:bucket",
				ResourceID:     b.Name,
				ResourceName:   b.Name,
				Severity:       model.SeverityFromSavings(oldVersionCost),
				Title:          fmt.Sprintf("S3 bucket with excessive old versions: %s (%.1f GB)", b.Name, float64(b.OldVersionBytes)/(1024*1024*1024)),
				Description:    fmt.Sprintf("Bucket %s has %.1f GB of non-current object versions costing $%.2f/month.", b.Name, float64(b.OldVersionBytes)/(1024*1024*1024), oldVersionCost),
				CurrentCost:    oldVersionCost,
				ProjectedCost:  0,
				MonthlySavings: oldVersionCost,
				AnnualSavings:  oldVersionCost * 12,
				Recommendation: "Add a lifecycle rule to expire non-current versions after 30 days.",
				Effort:         "low",
				Risk:           "low",
				DetectedAt:     time.Now().UTC(),
			}
			findings = append(findings, f)
		}
	}

	return findings, nil
}

type s3Bucket struct {
	Name               string
	SizeBytes          int64
	HasLifecyclePolicy bool
	HasVersioning      bool
	OldVersionBytes    int64
}

func (s *S3Scanner) listBucketsWithMetrics(ctx context.Context) ([]s3Bucket, error) {
	return nil, nil
}
