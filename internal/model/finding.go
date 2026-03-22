package model

import "time"

// Severity indicates how urgent the cost optimization is.
type Severity string

const (
	SeverityCritical Severity = "critical" // >$500/month waste
	SeverityHigh     Severity = "high"     // $100-500/month
	SeverityMedium   Severity = "medium"   // $25-100/month
	SeverityLow      Severity = "low"      // <$25/month
)

// Category groups findings by resource type.
type Category string

const (
	CategoryCompute  Category = "compute"
	CategoryStorage  Category = "storage"
	CategoryNetwork  Category = "network"
	CategoryDatabase Category = "database"
)

// Finding represents a single cost optimization opportunity.
type Finding struct {
	ID              string            `json:"id"`
	Provider        string            `json:"provider"`
	AccountID       string            `json:"account_id"`
	Region          string            `json:"region"`
	Category        Category          `json:"category"`
	ResourceType    string            `json:"resource_type"`
	ResourceID      string            `json:"resource_id"`
	ResourceName    string            `json:"resource_name"`
	Severity        Severity          `json:"severity"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	CurrentCost     float64           `json:"current_monthly_cost"`
	ProjectedCost   float64           `json:"projected_monthly_cost"`
	MonthlySavings  float64           `json:"monthly_savings"`
	AnnualSavings   float64           `json:"annual_savings"`
	Recommendation  string            `json:"recommendation"`
	Effort          string            `json:"effort"` // low, medium, high
	Risk            string            `json:"risk"`   // low, medium, high
	Tags            map[string]string `json:"tags,omitempty"`
	DetectedAt      time.Time         `json:"detected_at"`
	MetricsSummary  *MetricsSummary   `json:"metrics_summary,omitempty"`
}

// MetricsSummary holds utilization data that supports the finding.
type MetricsSummary struct {
	AvgCPUPercent     float64 `json:"avg_cpu_percent,omitempty"`
	MaxCPUPercent     float64 `json:"max_cpu_percent,omitempty"`
	AvgMemoryPercent  float64 `json:"avg_memory_percent,omitempty"`
	AvgNetworkIn      float64 `json:"avg_network_in_bytes,omitempty"`
	AvgNetworkOut     float64 `json:"avg_network_out_bytes,omitempty"`
	AvgDiskIOPS       float64 `json:"avg_disk_iops,omitempty"`
	StorageUsedBytes  int64   `json:"storage_used_bytes,omitempty"`
	StorageTotalBytes int64   `json:"storage_total_bytes,omitempty"`
	ObservationDays   int     `json:"observation_days"`
}

// SeverityFromSavings determines severity based on monthly savings amount.
func SeverityFromSavings(monthlySavings float64) Severity {
	switch {
	case monthlySavings >= 500:
		return SeverityCritical
	case monthlySavings >= 100:
		return SeverityHigh
	case monthlySavings >= 25:
		return SeverityMedium
	default:
		return SeverityLow
	}
}
