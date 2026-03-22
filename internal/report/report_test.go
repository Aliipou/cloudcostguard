package report

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/model"
)

func sampleFindings() []model.Finding {
	return []model.Finding{
		{
			ID:             "test-1",
			Provider:       "aws",
			Region:         "us-east-1",
			Category:       model.CategoryCompute,
			ResourceType:   "ec2:instance",
			ResourceID:     "i-123",
			ResourceName:   "web-server-1",
			Severity:       model.SeverityHigh,
			Title:          "Idle EC2 instance: web-server-1",
			CurrentCost:    140.16,
			ProjectedCost:  0,
			MonthlySavings: 140.16,
			AnnualSavings:  1681.92,
			Recommendation: "Terminate instance",
			Effort:         "low",
			Risk:           "medium",
			DetectedAt:     time.Now(),
		},
		{
			ID:             "test-2",
			Provider:       "azure",
			Region:         "westeurope",
			Category:       model.CategoryStorage,
			ResourceType:   "microsoft.compute/disks",
			ResourceID:     "/sub/123/disk/test",
			ResourceName:   "orphaned-disk",
			Severity:       model.SeverityLow,
			Title:          "Unattached Azure disk: orphaned-disk",
			CurrentCost:    16.90,
			ProjectedCost:  0,
			MonthlySavings: 16.90,
			AnnualSavings:  202.80,
			Recommendation: "Delete disk",
			Effort:         "low",
			Risk:           "low",
			DetectedAt:     time.Now(),
		},
	}
}

func TestToJSON(t *testing.T) {
	findings := sampleFindings()
	report := ToJSON(findings)

	if report.Summary.TotalFindings != 2 {
		t.Errorf("TotalFindings = %d, want 2", report.Summary.TotalFindings)
	}

	expectedMonthly := 140.16 + 16.90
	if report.Summary.TotalMonthly != expectedMonthly {
		t.Errorf("TotalMonthly = %.2f, want %.2f", report.Summary.TotalMonthly, expectedMonthly)
	}

	if report.Summary.HighCount != 1 {
		t.Errorf("HighCount = %d, want 1", report.Summary.HighCount)
	}

	if report.Summary.LowCount != 1 {
		t.Errorf("LowCount = %d, want 1", report.Summary.LowCount)
	}

	computeSavings := report.Summary.ByCategory["compute"]
	if computeSavings != 140.16 {
		t.Errorf("Compute savings = %.2f, want 140.16", computeSavings)
	}
}

func TestWriteTable(t *testing.T) {
	var buf bytes.Buffer
	WriteTable(&buf, sampleFindings(), 0)

	output := buf.String()

	if !strings.Contains(output, "SEVERITY") {
		t.Error("Table should contain header")
	}
	if !strings.Contains(output, "TOTAL (2 findings)") {
		t.Error("Table should contain total line")
	}
	if !strings.Contains(output, "$140.16") {
		t.Error("Table should contain finding amount")
	}
}

func TestWriteTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	WriteTable(&buf, nil, 0)

	if !strings.Contains(buf.String(), "No optimization opportunities") {
		t.Error("Empty findings should show 'no opportunities' message")
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	err := WriteCSV(&buf, sampleFindings())
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	// header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV rows = %d, want 3 (header + 2 findings)", len(records))
	}

	if records[0][0] != "severity" {
		t.Errorf("First header = %s, want 'severity'", records[0][0])
	}
}
