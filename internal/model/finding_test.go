package model

import "testing"

func TestSeverityFromSavings(t *testing.T) {
	tests := []struct {
		name     string
		savings  float64
		expected Severity
	}{
		{"critical threshold", 500.0, SeverityCritical},
		{"above critical", 1000.0, SeverityCritical},
		{"high threshold", 100.0, SeverityHigh},
		{"high range", 250.0, SeverityHigh},
		{"medium threshold", 25.0, SeverityMedium},
		{"medium range", 50.0, SeverityMedium},
		{"low range", 24.99, SeverityLow},
		{"zero savings", 0.0, SeverityLow},
		{"negative savings", -5.0, SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SeverityFromSavings(tt.savings)
			if got != tt.expected {
				t.Errorf("SeverityFromSavings(%.2f) = %s, want %s", tt.savings, got, tt.expected)
			}
		})
	}
}
