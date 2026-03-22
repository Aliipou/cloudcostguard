package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

// NetworkScanner detects unused Azure networking resources (public IPs, load balancers).
type NetworkScanner struct {
	cfg            *config.Config
	subscriptionID string
}

func NewNetworkScanner(cfg *config.Config, subscriptionID string) *NetworkScanner {
	return &NetworkScanner{cfg: cfg, subscriptionID: subscriptionID}
}

func (s *NetworkScanner) Name() string            { return "azure-network" }
func (s *NetworkScanner) Category() model.Category { return model.CategoryNetwork }

func (s *NetworkScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	ips, err := s.listUnassociatedPublicIPs(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing unassociated public IPs: %w", err)
	}

	var findings []model.Finding

	for _, ip := range ips {
		// Static public IP costs ~$3.65/month
		monthlyCost := 3.65
		if ip.SKU == "Standard" {
			monthlyCost = 3.65
		}

		f := model.Finding{
			ID:             fmt.Sprintf("azure-pip-unused-%s", ip.ResourceID),
			Provider:       "azure",
			AccountID:      s.subscriptionID,
			Region:         ip.Location,
			Category:       model.CategoryNetwork,
			ResourceType:   "microsoft.network/publicipaddresses",
			ResourceID:     ip.ResourceID,
			ResourceName:   ip.Name,
			Severity:       model.SeverityFromSavings(monthlyCost),
			Title:          fmt.Sprintf("Unassociated public IP: %s (%s)", ip.Name, ip.Address),
			Description:    fmt.Sprintf("Public IP %s (%s) in %s is not associated with any resource, costing $%.2f/month.", ip.Name, ip.Address, ip.Location, monthlyCost),
			CurrentCost:    monthlyCost,
			ProjectedCost:  0,
			MonthlySavings: monthlyCost,
			AnnualSavings:  monthlyCost * 12,
			Recommendation: "Delete this public IP if no longer needed. Unassociated static IPs incur charges.",
			Effort:         "low",
			Risk:           "low",
			Tags:           ip.Tags,
			DetectedAt:     time.Now().UTC(),
		}
		findings = append(findings, f)
	}

	return findings, nil
}

type publicIP struct {
	ResourceID string
	Name       string
	Address    string
	SKU        string
	Location   string
	Tags       map[string]string
}

func (s *NetworkScanner) listUnassociatedPublicIPs(ctx context.Context) ([]publicIP, error) {
	return nil, nil
}
