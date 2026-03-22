package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScansTotal(t *testing.T) {
	c := New()
	c.IncScansTotal()
	c.IncScansTotal()
	c.IncScansTotal()

	body := getMetrics(t, c)

	if !strings.Contains(body, "cloudcostguard_scans_total 3") {
		t.Errorf("expected scans_total 3, got body:\n%s", body)
	}
}

func TestActiveScans(t *testing.T) {
	c := New()
	c.IncActiveScans()
	c.IncActiveScans()
	c.DecActiveScans()

	body := getMetrics(t, c)

	if !strings.Contains(body, "cloudcostguard_active_scans 1") {
		t.Errorf("expected active_scans 1, got body:\n%s", body)
	}
}

func TestActiveScansDoesNotGoBelowZero(t *testing.T) {
	c := New()
	c.DecActiveScans()

	body := getMetrics(t, c)

	if !strings.Contains(body, "cloudcostguard_active_scans 0") {
		t.Errorf("expected active_scans 0, got body:\n%s", body)
	}
}

func TestSetActiveScans(t *testing.T) {
	c := New()
	c.SetActiveScans(5)

	body := getMetrics(t, c)

	if !strings.Contains(body, "cloudcostguard_active_scans 5") {
		t.Errorf("expected active_scans 5, got body:\n%s", body)
	}
}

func TestFindingsTotal(t *testing.T) {
	c := New()
	c.IncFindingsTotal("critical", "compute")
	c.IncFindingsTotal("critical", "compute")
	c.IncFindingsTotal("high", "storage")
	c.IncFindingsTotal("low", "network")

	body := getMetrics(t, c)

	expected := []string{
		`cloudcostguard_findings_total{severity="critical",category="compute"} 2`,
		`cloudcostguard_findings_total{severity="high",category="storage"} 1`,
		`cloudcostguard_findings_total{severity="low",category="network"} 1`,
	}
	for _, exp := range expected {
		if !strings.Contains(body, exp) {
			t.Errorf("expected %q in body:\n%s", exp, body)
		}
	}
}

func TestScanDurationHistogram(t *testing.T) {
	c := New()
	c.ObserveScanDuration(0.3)  // fits in 0.5 bucket
	c.ObserveScanDuration(1.5)  // fits in 2.5 bucket
	c.ObserveScanDuration(7.0)  // fits in 10 bucket
	c.ObserveScanDuration(500)  // exceeds all buckets

	body := getMetrics(t, c)

	// Check that histogram lines are present.
	if !strings.Contains(body, `cloudcostguard_scan_duration_seconds_bucket{le="0.5"} 1`) {
		t.Errorf("expected le=0.5 bucket count 1, got body:\n%s", body)
	}
	if !strings.Contains(body, `cloudcostguard_scan_duration_seconds_bucket{le="2.5"} 2`) {
		t.Errorf("expected le=2.5 bucket count 2, got body:\n%s", body)
	}
	if !strings.Contains(body, `cloudcostguard_scan_duration_seconds_bucket{le="10"} 3`) {
		t.Errorf("expected le=10 bucket count 3, got body:\n%s", body)
	}
	if !strings.Contains(body, `cloudcostguard_scan_duration_seconds_bucket{le="+Inf"} 4`) {
		t.Errorf("expected +Inf bucket count 4, got body:\n%s", body)
	}
	if !strings.Contains(body, "cloudcostguard_scan_duration_seconds_count 4") {
		t.Errorf("expected duration count 4, got body:\n%s", body)
	}
	if !strings.Contains(body, "cloudcostguard_scan_duration_seconds_sum 508.8") {
		t.Errorf("expected duration sum 508.8, got body:\n%s", body)
	}
}

func TestMetricsContentType(t *testing.T) {
	c := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	c.Handler()(rec, req)

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %q", ct)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMetricsHasTypeAndHelpLines(t *testing.T) {
	c := New()
	body := getMetrics(t, c)

	requiredHeaders := []string{
		"# HELP cloudcostguard_scans_total",
		"# TYPE cloudcostguard_scans_total counter",
		"# HELP cloudcostguard_active_scans",
		"# TYPE cloudcostguard_active_scans gauge",
		"# HELP cloudcostguard_findings_total",
		"# TYPE cloudcostguard_findings_total counter",
		"# HELP cloudcostguard_scan_duration_seconds",
		"# TYPE cloudcostguard_scan_duration_seconds histogram",
	}
	for _, h := range requiredHeaders {
		if !strings.Contains(body, h) {
			t.Errorf("expected %q in metrics output", h)
		}
	}
}

func TestMetricsEmptyState(t *testing.T) {
	c := New()
	body := getMetrics(t, c)

	if !strings.Contains(body, "cloudcostguard_scans_total 0") {
		t.Errorf("expected scans_total 0 on fresh collector, got body:\n%s", body)
	}
	if !strings.Contains(body, "cloudcostguard_active_scans 0") {
		t.Errorf("expected active_scans 0 on fresh collector, got body:\n%s", body)
	}
}

// getMetrics is a test helper that invokes the metrics handler and returns the
// response body as a string.
func getMetrics(t *testing.T, c *Collector) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	c.Handler()(rec, req)
	return rec.Body.String()
}
