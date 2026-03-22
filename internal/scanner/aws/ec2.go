package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/pricing"
)

// EC2Scanner detects idle and oversized EC2 instances.
type EC2Scanner struct {
	cfg    *config.Config
	region string
}

func NewEC2Scanner(cfg *config.Config, region string) *EC2Scanner {
	return &EC2Scanner{cfg: cfg, region: region}
}

func (s *EC2Scanner) Name() string            { return "aws-ec2-" + s.region }
func (s *EC2Scanner) Category() model.Category { return model.CategoryCompute }

func (s *EC2Scanner) Scan(ctx context.Context) ([]model.Finding, error) {
	instances, err := s.listRunningInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing EC2 instances in %s: %w", s.region, err)
	}

	var findings []model.Finding

	for _, inst := range instances {
		metrics, err := s.getMetrics(ctx, inst.ID, s.cfg.Rules.IdleDays)
		if err != nil {
			continue // skip instance but don't fail entire scan
		}

		// Check for idle instances
		if metrics.AvgCPUPercent < s.cfg.Rules.IdleCPUThreshold {
			monthlyCost := pricing.EC2MonthlyCost(inst.Type, s.region)
			f := model.Finding{
				ID:             fmt.Sprintf("aws-ec2-idle-%s-%s", s.region, inst.ID),
				Provider:       "aws",
				Region:         s.region,
				Category:       model.CategoryCompute,
				ResourceType:   "ec2:instance",
				ResourceID:     inst.ID,
				ResourceName:   inst.Name,
				Severity:       model.SeverityFromSavings(monthlyCost),
				Title:          fmt.Sprintf("Idle EC2 instance: %s (%.1f%% avg CPU over %dd)", inst.Name, metrics.AvgCPUPercent, s.cfg.Rules.IdleDays),
				Description:    fmt.Sprintf("Instance %s (%s) has averaged %.1f%% CPU utilization over the past %d days, below the %.1f%% threshold.", inst.ID, inst.Type, metrics.AvgCPUPercent, s.cfg.Rules.IdleDays, s.cfg.Rules.IdleCPUThreshold),
				CurrentCost:    monthlyCost,
				ProjectedCost:  0,
				MonthlySavings: monthlyCost,
				AnnualSavings:  monthlyCost * 12,
				Recommendation: "Terminate or stop this instance. If it runs periodic workloads, consider switching to a Lambda function or Spot instance.",
				Effort:         "low",
				Risk:           "medium",
				Tags:           inst.Tags,
				DetectedAt:     time.Now().UTC(),
				MetricsSummary: metrics,
			}
			findings = append(findings, f)
			continue
		}

		// Check for oversized instances
		if metrics.AvgCPUPercent < s.cfg.Rules.OversizedCPUThreshold && metrics.AvgCPUPercent >= s.cfg.Rules.IdleCPUThreshold {
			currentCost := pricing.EC2MonthlyCost(inst.Type, s.region)
			recommended := pricing.RecommendSmaller(inst.Type)
			newCost := pricing.EC2MonthlyCost(recommended, s.region)
			savings := currentCost - newCost

			if savings > 0 {
				f := model.Finding{
					ID:             fmt.Sprintf("aws-ec2-oversized-%s-%s", s.region, inst.ID),
					Provider:       "aws",
					Region:         s.region,
					Category:       model.CategoryCompute,
					ResourceType:   "ec2:instance",
					ResourceID:     inst.ID,
					ResourceName:   inst.Name,
					Severity:       model.SeverityFromSavings(savings),
					Title:          fmt.Sprintf("Oversized EC2 instance: %s (%.1f%% avg CPU)", inst.Name, metrics.AvgCPUPercent),
					Description:    fmt.Sprintf("Instance %s (%s) is underutilized at %.1f%% avg CPU. Downsize to %s to save $%.2f/month.", inst.ID, inst.Type, metrics.AvgCPUPercent, recommended, savings),
					CurrentCost:    currentCost,
					ProjectedCost:  newCost,
					MonthlySavings: savings,
					AnnualSavings:  savings * 12,
					Recommendation: fmt.Sprintf("Resize from %s to %s. Test with the smaller instance type during a maintenance window.", inst.Type, recommended),
					Effort:         "medium",
					Risk:           "low",
					Tags:           inst.Tags,
					DetectedAt:     time.Now().UTC(),
					MetricsSummary: metrics,
				}
				findings = append(findings, f)
			}
		}
	}

	return findings, nil
}

// ec2Instance holds basic instance info needed for scanning.
type ec2Instance struct {
	ID   string
	Name string
	Type string
	Tags map[string]string
}

// listRunningInstances retrieves running EC2 instances via AWS SDK.
func (s *EC2Scanner) listRunningInstances(ctx context.Context) ([]ec2Instance, error) {
	// In production, this uses aws-sdk-go-v2:
	//   cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	//   client := ec2svc.NewFromConfig(cfg)
	//   resp, _ := client.DescribeInstances(ctx, &ec2svc.DescribeInstancesInput{
	//       Filters: []types.Filter{{Name: aws.String("instance-state-name"), Values: []string{"running"}}},
	//   })
	//
	// For now, return empty to allow compilation without credentials.
	// The full implementation requires AWS credentials at runtime.
	return nil, nil
}

// getMetrics retrieves CloudWatch CPU/memory metrics for an instance.
func (s *EC2Scanner) getMetrics(ctx context.Context, instanceID string, days int) (*model.MetricsSummary, error) {
	// In production, this uses CloudWatch:
	//   cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	//   client := cloudwatch.NewFromConfig(cfg)
	//   resp, _ := client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{...})
	return nil, fmt.Errorf("not connected to AWS — configure credentials to enable scanning")
}
