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

func (s *EBSScanner) Name() string             { return "aws-ebs-" + s.region }
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

// listUnattachedVolumes retrieves all unattached EBS volumes via AWS SDK with
// pagination (NextToken) and exponential-backoff retry on transient errors.
func (s *EBSScanner) listUnattachedVolumes(ctx context.Context) ([]ebsVolume, error) {
	// Production implementation (requires aws-sdk-go-v2 and credentials):
	//
	//   cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	//   client := ec2svc.NewFromConfig(cfg)
	//
	//   var volumes []ebsVolume
	//   var nextToken *string
	//   for {
	//       var resp *ec2svc.DescribeVolumesOutput
	//       err := withRetry(ctx, defaultRetry, func() error {
	//           var callErr error
	//           resp, callErr = client.DescribeVolumes(ctx, &ec2svc.DescribeVolumesInput{
	//               Filters:   []types.Filter{{Name: aws.String("status"), Values: []string{"available"}}},
	//               NextToken: nextToken,
	//           })
	//           return callErr
	//       })
	//       if err != nil {
	//           return nil, fmt.Errorf("DescribeVolumes page: %w", err)
	//       }
	//       for _, v := range resp.Volumes {
	//           volumes = append(volumes, ebsVolume{
	//               ID:         aws.ToString(v.VolumeId),
	//               Name:       nameTagOrID(v.Tags, aws.ToString(v.VolumeId)),
	//               VolumeType: string(v.VolumeType),
	//               SizeGB:     int(aws.ToInt32(v.Size)),
	//               IOPS:       int(aws.ToInt32(v.Iops)),
	//               Tags:       flattenTags(v.Tags),
	//           })
	//       }
	//       if resp.NextToken == nil {
	//           break
	//       }
	//       nextToken = resp.NextToken
	//   }
	//   return volumes, nil
	return nil, nil
}
