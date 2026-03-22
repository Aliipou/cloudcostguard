package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Rules.IdleCPUThreshold != 5.0 {
		t.Errorf("IdleCPUThreshold = %.1f, want 5.0", cfg.Rules.IdleCPUThreshold)
	}
	if cfg.Rules.IdleDays != 14 {
		t.Errorf("IdleDays = %d, want 14", cfg.Rules.IdleDays)
	}
	if cfg.Rules.UnattachedDiskDays != 7 {
		t.Errorf("UnattachedDiskDays = %d, want 7", cfg.Rules.UnattachedDiskDays)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load should not error on missing file, got: %v", err)
	}
	if cfg.Rules.IdleCPUThreshold != 5.0 {
		t.Error("Should return defaults for missing file")
	}
}

func TestLoadValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
provider: aws
aws:
  profile: production
  regions:
    - us-east-1
    - eu-west-1
rules:
  idle_cpu_threshold: 10.0
  idle_days: 7
  unattached_disk_days: 3
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Provider != "aws" {
		t.Errorf("Provider = %s, want aws", cfg.Provider)
	}
	if cfg.AWS == nil {
		t.Fatal("AWS config should not be nil")
	}
	if cfg.AWS.Profile != "production" {
		t.Errorf("AWS.Profile = %s, want production", cfg.AWS.Profile)
	}
	if len(cfg.AWS.Regions) != 2 {
		t.Errorf("AWS.Regions length = %d, want 2", len(cfg.AWS.Regions))
	}
	if cfg.Rules.IdleCPUThreshold != 10.0 {
		t.Errorf("IdleCPUThreshold = %.1f, want 10.0", cfg.Rules.IdleCPUThreshold)
	}
	if cfg.Rules.IdleDays != 7 {
		t.Errorf("IdleDays = %d, want 7", cfg.Rules.IdleDays)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load should error on invalid YAML")
	}
}
