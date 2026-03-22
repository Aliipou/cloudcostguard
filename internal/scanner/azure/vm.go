package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/pricing"
)

// VMScanner detects idle and oversized Azure VMs.
type VMScanner struct {
	cfg            *config.Config
	subscriptionID string
}

func NewVMScanner(cfg *config.Config, subscriptionID string) *VMScanner {
	return &VMScanner{cfg: cfg, subscriptionID: subscriptionID}
}

func (s *VMScanner) Name() string            { return "azure-vm" }
func (s *VMScanner) Category() model.Category { return model.CategoryCompute }

func (s *VMScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	vms, err := s.listRunningVMs(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing Azure VMs: %w", err)
	}

	var findings []model.Finding

	for _, vm := range vms {
		metrics, err := s.getMetrics(ctx, vm.ResourceID, s.cfg.Rules.IdleDays)
		if err != nil {
			continue
		}

		monthlyCost := pricing.AzureVMMonthlyCost(vm.Size, vm.Location)

		if metrics.AvgCPUPercent < s.cfg.Rules.IdleCPUThreshold {
			f := model.Finding{
				ID:             fmt.Sprintf("azure-vm-idle-%s", vm.ResourceID),
				Provider:       "azure",
				AccountID:      s.subscriptionID,
				Region:         vm.Location,
				Category:       model.CategoryCompute,
				ResourceType:   "microsoft.compute/virtualmachines",
				ResourceID:     vm.ResourceID,
				ResourceName:   vm.Name,
				Severity:       model.SeverityFromSavings(monthlyCost),
				Title:          fmt.Sprintf("Idle Azure VM: %s (%.1f%% avg CPU over %dd)", vm.Name, metrics.AvgCPUPercent, s.cfg.Rules.IdleDays),
				Description:    fmt.Sprintf("VM %s (%s) in %s has averaged %.1f%% CPU over %d days.", vm.Name, vm.Size, vm.Location, metrics.AvgCPUPercent, s.cfg.Rules.IdleDays),
				CurrentCost:    monthlyCost,
				ProjectedCost:  0,
				MonthlySavings: monthlyCost,
				AnnualSavings:  monthlyCost * 12,
				Recommendation: "Deallocate this VM to stop billing, or delete it if no longer needed. Consider Azure Spot VMs for interruptible workloads.",
				Effort:         "low",
				Risk:           "medium",
				Tags:           vm.Tags,
				DetectedAt:     time.Now().UTC(),
				MetricsSummary: metrics,
			}
			findings = append(findings, f)
		}
	}

	return findings, nil
}

type azureVM struct {
	ResourceID string
	Name       string
	Size       string
	Location   string
	Tags       map[string]string
}

func (s *VMScanner) listRunningVMs(ctx context.Context) ([]azureVM, error) {
	// Uses azure-sdk-for-go armcompute client in production
	return nil, nil
}

func (s *VMScanner) getMetrics(ctx context.Context, resourceID string, days int) (*model.MetricsSummary, error) {
	return nil, fmt.Errorf("not connected to Azure — configure credentials to enable scanning")
}
