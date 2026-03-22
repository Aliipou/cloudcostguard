package azure

import (
	"testing"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

func TestDiskScanner(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewDiskScanner(cfg, "sub-123")
	if s.Name() != "azure-disk" {
		t.Errorf("Name() = %s, want azure-disk", s.Name())
	}
	if s.Category() != model.CategoryStorage {
		t.Errorf("Category() = %s, want storage", s.Category())
	}
}

func TestNetworkScanner(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewNetworkScanner(cfg, "sub-123")
	if s.Name() != "azure-network" {
		t.Errorf("Name() = %s, want azure-network", s.Name())
	}
	if s.Category() != model.CategoryNetwork {
		t.Errorf("Category() = %s, want network", s.Category())
	}
}

func TestNewScanners_All(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Azure = &config.AzureConfig{SubscriptionID: "sub-123"}
	scanners, err := NewScanners(cfg, "all")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 3 {
		t.Errorf("expected 3 scanners for 'all', got %d", len(scanners))
	}
}

func TestNewScanners_Compute(t *testing.T) {
	cfg := config.DefaultConfig()
	scanners, err := NewScanners(cfg, "compute")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 1 {
		t.Errorf("expected 1 scanner for 'compute', got %d", len(scanners))
	}
}

func TestNewScanners_Storage(t *testing.T) {
	cfg := config.DefaultConfig()
	scanners, err := NewScanners(cfg, "storage")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 1 {
		t.Errorf("expected 1 scanner for 'storage', got %d", len(scanners))
	}
}

func TestNewScanners_Network(t *testing.T) {
	cfg := config.DefaultConfig()
	scanners, err := NewScanners(cfg, "network")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 1 {
		t.Errorf("expected 1 scanner for 'network', got %d", len(scanners))
	}
}
