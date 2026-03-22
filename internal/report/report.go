package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Aliipou/cloudcostguard/internal/model"
)

// JSONReport is the structured output for JSON format.
type JSONReport struct {
	Summary  ReportSummary   `json:"summary"`
	Findings []model.Finding `json:"findings"`
}

// ReportSummary holds aggregate numbers.
type ReportSummary struct {
	TotalFindings   int     `json:"total_findings"`
	TotalMonthly    float64 `json:"total_monthly_savings"`
	TotalAnnual     float64 `json:"total_annual_savings"`
	CriticalCount   int     `json:"critical_count"`
	HighCount       int     `json:"high_count"`
	MediumCount     int     `json:"medium_count"`
	LowCount        int     `json:"low_count"`
	ByCategory      map[string]float64 `json:"savings_by_category"`
}

// ToJSON creates a structured report from findings.
func ToJSON(findings []model.Finding) JSONReport {
	summary := ReportSummary{
		TotalFindings: len(findings),
		ByCategory:    make(map[string]float64),
	}

	for _, f := range findings {
		summary.TotalMonthly += f.MonthlySavings
		summary.TotalAnnual += f.AnnualSavings
		summary.ByCategory[string(f.Category)] += f.MonthlySavings

		switch f.Severity {
		case model.SeverityCritical:
			summary.CriticalCount++
		case model.SeverityHigh:
			summary.HighCount++
		case model.SeverityMedium:
			summary.MediumCount++
		case model.SeverityLow:
			summary.LowCount++
		}
	}

	return JSONReport{Summary: summary, Findings: findings}
}

// WriteTable writes a human-readable table to the writer.
func WriteTable(w io.Writer, findings []model.Finding, minSavings float64) {
	if len(findings) == 0 {
		fmt.Fprintln(w, "No optimization opportunities found.")
		return
	}

	// Sort by monthly savings descending
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].MonthlySavings > findings[j].MonthlySavings
	})

	// Print header
	fmt.Fprintf(w, "%-10s %-10s %-10s %-40s %12s %12s %-8s\n",
		"SEVERITY", "PROVIDER", "CATEGORY", "TITLE", "MONTHLY", "ANNUAL", "EFFORT")
	fmt.Fprintln(w, strings.Repeat("-", 110))

	totalMonthly := 0.0
	totalAnnual := 0.0

	for _, f := range findings {
		title := f.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		fmt.Fprintf(w, "%-10s %-10s %-10s %-40s %11s %11s %-8s\n",
			strings.ToUpper(string(f.Severity)),
			f.Provider,
			f.Category,
			title,
			fmt.Sprintf("$%.2f", f.MonthlySavings),
			fmt.Sprintf("$%.2f", f.AnnualSavings),
			f.Effort,
		)

		totalMonthly += f.MonthlySavings
		totalAnnual += f.AnnualSavings
	}

	fmt.Fprintln(w, strings.Repeat("-", 110))
	fmt.Fprintf(w, "%-70s %11s %11s\n",
		fmt.Sprintf("TOTAL (%d findings)", len(findings)),
		fmt.Sprintf("$%.2f", totalMonthly),
		fmt.Sprintf("$%.2f", totalAnnual),
	)
}

// WriteCSV writes findings in CSV format.
func WriteCSV(w io.Writer, findings []model.Finding) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"severity", "provider", "region", "category", "resource_type",
		"resource_id", "resource_name", "title", "current_monthly_cost",
		"projected_monthly_cost", "monthly_savings", "annual_savings",
		"recommendation", "effort", "risk",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, f := range findings {
		row := []string{
			string(f.Severity),
			f.Provider,
			f.Region,
			string(f.Category),
			f.ResourceType,
			f.ResourceID,
			f.ResourceName,
			f.Title,
			fmt.Sprintf("%.2f", f.CurrentCost),
			fmt.Sprintf("%.2f", f.ProjectedCost),
			fmt.Sprintf("%.2f", f.MonthlySavings),
			fmt.Sprintf("%.2f", f.AnnualSavings),
			f.Recommendation,
			f.Effort,
			f.Risk,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
