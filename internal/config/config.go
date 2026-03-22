package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Provider string       `yaml:"provider"`
	AWS      *AWSConfig   `yaml:"aws,omitempty"`
	Azure    *AzureConfig `yaml:"azure,omitempty"`
	Rules    RulesConfig  `yaml:"rules"`
}

// AWSConfig holds AWS-specific settings.
type AWSConfig struct {
	Profile string   `yaml:"profile"`
	Regions []string `yaml:"regions"`
	RoleARN string   `yaml:"role_arn,omitempty"`
}

// AzureConfig holds Azure-specific settings.
type AzureConfig struct {
	SubscriptionID string `yaml:"subscription_id"`
	TenantID       string `yaml:"tenant_id,omitempty"`
}

// RulesConfig defines thresholds for detecting waste.
type RulesConfig struct {
	IdleCPUThreshold       float64 `yaml:"idle_cpu_threshold"`
	IdleDays               int     `yaml:"idle_days"`
	UnattachedDiskDays     int     `yaml:"unattached_disk_days"`
	OversizedCPUThreshold  float64 `yaml:"oversized_cpu_threshold"`
	OversizedMemThreshold  float64 `yaml:"oversized_mem_threshold"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Rules: RulesConfig{
			IdleCPUThreshold:       5.0,
			IdleDays:               14,
			UnattachedDiskDays:     7,
			OversizedCPUThreshold:  20.0,
			OversizedMemThreshold:  20.0,
		},
	}
}

// Load reads configuration from file, falling back to defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, nil
		}
		path = filepath.Join(home, ".cloudcostguard.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}
