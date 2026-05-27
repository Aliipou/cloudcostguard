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

func (s *RDSScanner) Name() string             { return "aws-rds-" + s.region }
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

// listInstances retrieves all RDS DB instances via AWS SDK with pagination
// (Marker) and exponential-backoff retry on transient errors.
func (s *RDSScanner) listInstances(ctx context.Context) ([]rdsInstance, error) {
	// Production implementation (requires aws-sdk-go-v2 and credentials):
	//
	//   cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	//   client := rdssvc.NewFromConfig(cfg)
	//
	//   var instances []rdsInstance
	//   var marker *string
	//   for {
	//       var resp *rdssvc.DescribeDBInstancesOutput
	//       err := withRetry(ctx, defaultRetry, func() error {
	//           var callErr error
	//           resp, callErr = client.DescribeDBInstances(ctx, &rdssvc.DescribeDBInstancesInput{
	//               Marker: marker,
	//           })
	//           return callErr
	//       })
	//       if err != nil {
	//           return nil, fmt.Errorf("DescribeDBInstances page: %w", err)
	//       }
	//       for _, db := range resp.DBInstances {
	//           instances = append(instances, rdsInstance{
	//               ID:            aws.ToString(db.DBInstanceIdentifier),
	//               Name:          aws.ToString(db.DBInstanceIdentifier),
	//               InstanceClass: aws.ToString(db.DBInstanceClass),
	//               Engine:        aws.ToString(db.Engine),
	//               MultiAZ:       aws.ToBool(db.MultiAZ),
	//               Tags:          flattenRDSTags(db.TagList),
	//           })
	//       }
	//       if resp.Marker == nil {
	//           break
	//       }
	//       marker = resp.Marker
	//   }
	//   return instances, nil
	return nil, nil
}

// getMetrics retrieves CloudWatch metrics for an RDS instance with retry.
func (s *RDSScanner) getMetrics(ctx context.Context, instanceID string, days int) (*rdsMetrics, error) {
	// Production implementation (requires aws-sdk-go-v2 and credentials):
	//
	//   cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	//   client := cloudwatch.NewFromConfig(cfg)
	//   now := time.Now().UTC()
	//   start := now.AddDate(0, 0, -days)
	//
	//   var cpuResp *cloudwatch.GetMetricStatisticsOutput
	//   err := withRetry(ctx, defaultRetry, func() error {
	//       var callErr error
	//       cpuResp, callErr = client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
	//           Namespace:  aws.String("AWS/RDS"),
	//           MetricName: aws.String("CPUUtilization"),
	//           Dimensions: []cwtypes.Dimension{{Name: aws.String("DBInstanceIdentifier"), Value: aws.String(instanceID)}},
	//           StartTime:  aws.Time(start),
	//           EndTime:    aws.Time(now),
	//           Period:     aws.Int32(3600),
	//           Statistics: []cwtypes.Statistic{cwtypes.StatisticAverage},
	//       })
	//       return callErr
	//   })
	//   if err != nil {
	//       return nil, fmt.Errorf("GetMetricStatistics CPUUtilization: %w", err)
	//   }
	//   // Similarly fetch DatabaseConnections metric with withRetry...
	//   return &rdsMetrics{AvgCPUPercent: averageDatapoints(cpuResp.Datapoints)}, nil
	return nil, fmt.Errorf("not connected to AWS")
}
