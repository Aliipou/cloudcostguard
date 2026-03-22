package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/pricing"
)

// RDSScanner detects idle and oversized RDS instances.
type RDSScanner struct {
	cfg    *config.Config
	region string
}

func NewRDSScanner(cfg *config.Config, region string) *RDSScanner {
	return &RDSScanner{cfg: cfg, region: region}
}

func (s *RDSScanner) Name() string            { return "aws-rds-" + s.region }
func (s *RDSScanner) Category() model.Category { return model.CategoryDatabase }

func (s *RDSScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	instances, err := s.listInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing RDS instances in %s: %w", s.region, err)
	}

	var findings []model.Finding

	for _, db := range instances {
		metrics, err := s.getMetrics(ctx, db.ID, s.cfg.Rules.IdleDays)
		if err != nil {
			continue
		}

		currentCost := pricing.RDSMonthlyCost(db.InstanceClass, db.Engine, db.MultiAZ, s.region)

		// Idle RDS instance
		if metrics.AvgCPUPercent < s.cfg.Rules.IdleCPUThreshold && metrics.AvgConnections < 1 {
			f := model.Finding{
				ID:             fmt.Sprintf("aws-rds-idle-%s-%s", s.region, db.ID),
				Provider:       "aws",
				Region:         s.region,
				Category:       model.CategoryDatabase,
				ResourceType:   "rds:instance",
				ResourceID:     db.ID,
				ResourceName:   db.Name,
				Severity:       model.SeverityFromSavings(currentCost),
				Title:          fmt.Sprintf("Idle RDS instance: %s (%.1f%% CPU, %.0f avg connections)", db.Name, metrics.AvgCPUPercent, metrics.AvgConnections),
				Description:    fmt.Sprintf("RDS instance %s (%s, %s) has near-zero CPU and connections over %d days, costing $%.2f/month.", db.ID, db.InstanceClass, db.Engine, s.cfg.Rules.IdleDays, currentCost),
				CurrentCost:    currentCost,
				ProjectedCost:  0,
				MonthlySavings: currentCost,
				AnnualSavings:  currentCost * 12,
				Recommendation: "Take a final snapshot and delete this instance, or stop it if needed intermittently.",
				Effort:         "low",
				Risk:           "high",
				Tags:           db.Tags,
				DetectedAt:     time.Now().UTC(),
				MetricsSummary: &model.MetricsSummary{
					AvgCPUPercent:   metrics.AvgCPUPercent,
					ObservationDays: s.cfg.Rules.IdleDays,
				},
			}
			findings = append(findings, f)
			continue
		}

		// Oversized RDS — check if Multi-AZ is needed
		if db.MultiAZ && metrics.AvgCPUPercent < 30 {
			singleAZCost := pricing.RDSMonthlyCost(db.InstanceClass, db.Engine, false, s.region)
			savings := currentCost - singleAZCost

			if savings > 10 {
				f := model.Finding{
					ID:             fmt.Sprintf("aws-rds-multiaz-%s-%s", s.region, db.ID),
					Provider:       "aws",
					Region:         s.region,
					Category:       model.CategoryDatabase,
					ResourceType:   "rds:instance",
					ResourceID:     db.ID,
					ResourceName:   db.Name,
					Severity:       model.SeverityFromSavings(savings),
					Title:          fmt.Sprintf("RDS Multi-AZ may be unnecessary: %s", db.Name),
					Description:    fmt.Sprintf("RDS instance %s runs Multi-AZ at low utilization (%.1f%% CPU). If this is a dev/staging database, switching to Single-AZ saves $%.2f/month.", db.ID, metrics.AvgCPUPercent, savings),
					CurrentCost:    currentCost,
					ProjectedCost:  singleAZCost,
					MonthlySavings: savings,
					AnnualSavings:  savings * 12,
					Recommendation: "Evaluate if Multi-AZ is required. For non-production databases, Single-AZ reduces cost by ~50%.",
					Effort:         "medium",
					Risk:           "medium",
					Tags:           db.Tags,
					DetectedAt:     time.Now().UTC(),
				}
				findings = append(findings, f)
			}
		}
	}

	return findings, nil
}

type rdsInstance struct {
	ID            string
	Name          string
	InstanceClass string
	Engine        string
	MultiAZ       bool
	Tags          map[string]string
}

type rdsMetrics struct {
	AvgCPUPercent  float64
	AvgConnections float64
}

func (s *RDSScanner) listInstances(ctx context.Context) ([]rdsInstance, error) {
	return nil, nil
}

func (s *RDSScanner) getMetrics(ctx context.Context, instanceID string, days int) (*rdsMetrics, error) {
	return nil, fmt.Errorf("not connected to AWS")
}
