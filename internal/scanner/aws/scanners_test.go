package aws

import (
	"testing"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

func TestEBSScanner(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewEBSScanner(cfg, "us-east-1")
	if s.Name() != "aws-ebs-us-east-1" {
		t.Errorf("Name() = %s, want aws-ebs-us-east-1", s.Name())
	}
	if s.Category() != model.CategoryStorage {
		t.Errorf("Category() = %s, want storage", s.Category())
	}
}

func TestS3Scanner(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewS3Scanner(cfg, "us-east-1")
	if s.Name() != "aws-s3-us-east-1" {
		t.Errorf("Name() = %s, want aws-s3-us-east-1", s.Name())
	}
	if s.Category() != model.CategoryStorage {
		t.Errorf("Category() = %s, want storage", s.Category())
	}
}

func TestELBScanner(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewELBScanner(cfg, "us-east-1")
	if s.Name() != "aws-elb-us-east-1" {
		t.Errorf("Name() = %s, want aws-elb-us-east-1", s.Name())
	}
	if s.Category() != model.CategoryNetwork {
		t.Errorf("Category() = %s, want network", s.Category())
	}
}

func TestRDSScanner(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewRDSScanner(cfg, "us-east-1")
	if s.Name() != "aws-rds-us-east-1" {
		t.Errorf("Name() = %s, want aws-rds-us-east-1", s.Name())
	}
	if s.Category() != model.CategoryDatabase {
		t.Errorf("Category() = %s, want database", s.Category())
	}
}

func TestNewScanners_All(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AWS = &config.AWSConfig{Regions: []string{"us-east-1"}}
	scanners, err := NewScanners(cfg, "all")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 5 {
		t.Errorf("expected 5 scanners for 'all', got %d", len(scanners))
	}
}

func TestNewScanners_Compute(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AWS = &config.AWSConfig{Regions: []string{"us-east-1"}}
	scanners, err := NewScanners(cfg, "compute")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 1 {
		t.Errorf("expected 1 scanner for 'compute', got %d", len(scanners))
	}
}

func TestNewScanners_DefaultRegion(t *testing.T) {
	cfg := config.DefaultConfig()
	scanners, err := NewScanners(cfg, "compute")
	if err != nil {
		t.Fatalf("NewScanners error: %v", err)
	}
	if len(scanners) != 1 {
		t.Errorf("expected 1 scanner with default region, got %d", len(scanners))
	}
}
