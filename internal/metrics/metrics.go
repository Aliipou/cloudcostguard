// Package metrics provides Prometheus-compatible metrics using only the standard
// library. It exposes an HTTP handler that serves metrics in the Prometheus text
// exposition format at /metrics.
package metrics

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// Default histogram buckets for scan duration (in seconds).
var defaultDurationBuckets = []float64{0.5, 1, 2.5, 5, 10, 30, 60, 120, 300}

// Collector holds all application metrics.
type Collector struct {
	mu sync.Mutex

	scansTotal  float64
	activeScans float64

	// findings_total keyed by "severity,category"
	findingsTotal map[string]float64

	// histogram for scan_duration_seconds
	durationBuckets []float64
	durationCounts  []uint64 // one per bucket boundary
	durationSum     float64
	durationCount   uint64
}

// New creates a new metrics Collector.
func New() *Collector {
	buckets := make([]float64, len(defaultDurationBuckets))
	copy(buckets, defaultDurationBuckets)
	sort.Float64s(buckets)

	return &Collector{
		findingsTotal:   make(map[string]float64),
		durationBuckets: buckets,
		durationCounts:  make([]uint64, len(buckets)),
	}
}

// IncScansTotal increments the total scans counter by 1.
func (c *Collector) IncScansTotal() {
	c.mu.Lock()
	c.scansTotal++
	c.mu.Unlock()
}

// SetActiveScans sets the active scans gauge to the given value.
func (c *Collector) SetActiveScans(v float64) {
	c.mu.Lock()
	c.activeScans = v
	c.mu.Unlock()
}

// IncActiveScans increments the active scans gauge by 1.
func (c *Collector) IncActiveScans() {
	c.mu.Lock()
	c.activeScans++
	c.mu.Unlock()
}

// DecActiveScans decrements the active scans gauge by 1.
func (c *Collector) DecActiveScans() {
	c.mu.Lock()
	if c.activeScans > 0 {
		c.activeScans--
	}
	c.mu.Unlock()
}

// IncFindingsTotal increments the findings counter for the given severity and
// category.
func (c *Collector) IncFindingsTotal(severity, category string) {
	key := severity + "," + category
	c.mu.Lock()
	c.findingsTotal[key]++
	c.mu.Unlock()
}

// ObserveScanDuration records a scan duration observation in the histogram.
func (c *Collector) ObserveScanDuration(seconds float64) {
	c.mu.Lock()
	// Store in the first bucket whose boundary is >= the observation.
	for i, bound := range c.durationBuckets {
		if seconds <= bound {
			c.durationCounts[i]++
			break
		}
	}
	c.durationSum += seconds
	c.durationCount++
	c.mu.Unlock()
}

// Handler returns an http.HandlerFunc that writes all metrics in Prometheus
// text exposition format.
func (c *Collector) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		defer c.mu.Unlock()

		var b strings.Builder

		// scans_total
		b.WriteString("# HELP cloudcostguard_scans_total Total number of scans completed.\n")
		b.WriteString("# TYPE cloudcostguard_scans_total counter\n")
		fmt.Fprintf(&b, "cloudcostguard_scans_total %s\n", formatFloat(c.scansTotal))

		// active_scans
		b.WriteString("# HELP cloudcostguard_active_scans Number of scans currently in progress.\n")
		b.WriteString("# TYPE cloudcostguard_active_scans gauge\n")
		fmt.Fprintf(&b, "cloudcostguard_active_scans %s\n", formatFloat(c.activeScans))

		// findings_total
		b.WriteString("# HELP cloudcostguard_findings_total Total number of findings by severity and category.\n")
		b.WriteString("# TYPE cloudcostguard_findings_total counter\n")
		// Sort keys for deterministic output.
		keys := make([]string, 0, len(c.findingsTotal))
		for k := range c.findingsTotal {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			parts := strings.SplitN(key, ",", 2)
			severity := parts[0]
			category := ""
			if len(parts) > 1 {
				category = parts[1]
			}
			fmt.Fprintf(&b, "cloudcostguard_findings_total{severity=%q,category=%q} %s\n",
				severity, category, formatFloat(c.findingsTotal[key]))
		}

		// scan_duration_seconds histogram
		b.WriteString("# HELP cloudcostguard_scan_duration_seconds Duration of scans in seconds.\n")
		b.WriteString("# TYPE cloudcostguard_scan_duration_seconds histogram\n")
		var cumulative uint64
		for i, bound := range c.durationBuckets {
			cumulative += c.durationCounts[i]
			fmt.Fprintf(&b, "cloudcostguard_scan_duration_seconds_bucket{le=%q} %d\n",
				formatFloat(bound), cumulative)
		}
		fmt.Fprintf(&b, "cloudcostguard_scan_duration_seconds_bucket{le=\"+Inf\"} %d\n", c.durationCount)
		fmt.Fprintf(&b, "cloudcostguard_scan_duration_seconds_sum %s\n", formatFloat(c.durationSum))
		fmt.Fprintf(&b, "cloudcostguard_scan_duration_seconds_count %d\n", c.durationCount)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, b.String())
	}
}

// formatFloat formats a float64 for Prometheus exposition. Integers are
// printed without a decimal point.
func formatFloat(v float64) string {
	if v == math.Trunc(v) && !math.IsInf(v, 0) && !math.IsNaN(v) {
		return fmt.Sprintf("%g", v)
	}
	return fmt.Sprintf("%g", v)
}
