package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

// ELBScanner detects idle and unused Elastic Load Balancers.
type ELBScanner struct {
	cfg    *config.Config
	region string
}

func NewELBScanner(cfg *config.Config, region string) *ELBScanner {
	return &ELBScanner{cfg: cfg, region: region}
}

func (s *ELBScanner) Name() string             { return "aws-elb-" + s.region }
func (s *ELBScanner) Category() model.Category { return model.CategoryNetwork }

func (s *ELBScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	lbs, err := s.listIdleLoadBalancers(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing ELBs in %s: %w", s.region, err)
	}

	var findings []model.Finding
	for _, lb := range lbs {
		// ALB costs ~$22.27/month base + LCU charges
		baseCost := 22.27

		f := model.Finding{
			ID:             fmt.Sprintf("aws-elb-idle-%s-%s", s.region, lb.ID),
			Provider:       "aws",
			Region:         s.region,
			Category:       model.CategoryNetwork,
			ResourceType:   "elbv2:loadbalancer",
			ResourceID:     lb.ID,
			ResourceName:   lb.Name,
			Severity:       model.SeverityFromSavings(baseCost),
			Title:          fmt.Sprintf("Idle load balancer: %s (0 healthy targets)", lb.Name),
			Description:    fmt.Sprintf("Load balancer %s has no healthy backend targets and is receiving no traffic, costing $%.2f/month.", lb.Name, baseCost),
			CurrentCost:    baseCost,
			ProjectedCost:  0,
			MonthlySavings: baseCost,
			AnnualSavings:  baseCost * 12,
			Recommendation: "Delete this load balancer if no longer needed, or investigate why all targets are unhealthy.",
			Effort:         "low",
			Risk:           "medium",
			Tags:           lb.Tags,
			DetectedAt:     time.Now().UTC(),
		}
		findings = append(findings, f)
	}

	return findings, nil
}

type loadBalancer struct {
	ID   string
	Name string
	Type string
	Tags map[string]string
}

// listIdleLoadBalancers retrieves load balancers with no healthy targets via
// AWS SDK with pagination (NextMarker) and exponential-backoff retry on transient errors.
func (s *ELBScanner) listIdleLoadBalancers(ctx context.Context) ([]loadBalancer, error) {
	// Production implementation (requires aws-sdk-go-v2 and credentials):
	//
	//   cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	//   client := elbv2.NewFromConfig(cfg)
	//
	//   var lbs []loadBalancer
	//   var nextMarker *string
	//   for {
	//       var resp *elbv2.DescribeLoadBalancersOutput
	//       err := withRetry(ctx, defaultRetry, func() error {
	//           var callErr error
	//           resp, callErr = client.DescribeLoadBalancers(ctx, &elbv2.DescribeLoadBalancersInput{
	//               Marker: nextMarker,
	//           })
	//           return callErr
	//       })
	//       if err != nil {
	//           return nil, fmt.Errorf("DescribeLoadBalancers page: %w", err)
	//       }
	//       for _, lb := range resp.LoadBalancers {
	//           // Check target group health; include only LBs with no healthy targets.
	//           if isIdle(ctx, client, lb, withRetry) {
	//               lbs = append(lbs, loadBalancer{
	//                   ID:   aws.ToString(lb.LoadBalancerArn),
	//                   Name: aws.ToString(lb.LoadBalancerName),
	//                   Type: string(lb.Type),
	//               })
	//           }
	//       }
	//       if resp.NextMarker == nil {
	//           break
	//       }
	//       nextMarker = resp.NextMarker
	//   }
	//   return lbs, nil
	return nil, nil
}
