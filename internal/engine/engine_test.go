package engine

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/scanner"
)

// mockScanner implements scanner.Scanner for testing.
type mockScanner struct {
	name     string
	category model.Category
	findings []model.Finding
	err      error
	called   atomic.Int32
	delay    time.Duration
}

func (m *mockScanner) Name() string            { return m.name }
func (m *mockScanner) Category() model.Category { return m.category }

func (m *mockScanner) Scan(ctx context.Context) ([]model.Finding, error) {
	m.called.Add(1)
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.findings, m.err
}

// Compile-time check that mockScanner satisfies scanner.Scanner.
var _ scanner.Scanner = (*mockScanner)(nil)

func TestNew_UnsupportedProvider(t *testing.T) {
	cfg := &config.Config{Provider: "gcp"}
	_, err := New(cfg, Options{ScanType: "all", Concurrency: 1})
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
	want := "unsupported provider: gcp (supported: aws, azure)"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestNew_EmptyScanTypeNoMatch(t *testing.T) {
	cfg := &config.Config{
		Provider: "aws",
		AWS:      &config.AWSConfig{Regions: []string{"us-east-1"}},
	}
	// "bogus" doesn't match any switch case, so NewScanners returns empty slice
	_, err := New(cfg, Options{ScanType: "bogus", Concurrency: 1})
	if err == nil {
		t.Fatal("expected error for empty scan type match, got nil")
	}
	expected := `no scanners matched type "bogus" for provider aws`
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}

func TestScan_FiltersByMinSavings(t *testing.T) {
	cheap := model.Finding{
		ID:             "cheap",
		MonthlySavings: 5.0,
	}
	expensive := model.Finding{
		ID:             "expensive",
		MonthlySavings: 100.0,
	}

	e := &Engine{
		scanners: []scanner.Scanner{&mockScanner{
			name:     "test",
			category: model.CategoryCompute,
			findings: []model.Finding{cheap, expensive},
		}},
		opts: Options{MinSavings: 50.0, Concurrency: 1},
	}

	results, err := e.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(results))
	}
	if results[0].ID != "expensive" {
		t.Errorf("expected finding ID 'expensive', got %q", results[0].ID)
	}
}

func TestScan_HandlesScannerErrors(t *testing.T) {
	good := &mockScanner{
		name:     "good",
		category: model.CategoryCompute,
		findings: []model.Finding{{ID: "f1", MonthlySavings: 10}},
	}
	bad := &mockScanner{
		name: "bad",
		err:  fmt.Errorf("connection refused"),
	}

	e := &Engine{
		scanners: []scanner.Scanner{good, bad},
		opts:     Options{MinSavings: 0, Concurrency: 2},
	}

	results, err := e.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan() should not fail when some scanners succeed, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(results))
	}
}

func TestScan_AllScannersFailReturnsError(t *testing.T) {
	bad1 := &mockScanner{name: "bad1", err: fmt.Errorf("timeout")}
	bad2 := &mockScanner{name: "bad2", err: fmt.Errorf("denied")}

	e := &Engine{
		scanners: []scanner.Scanner{bad1, bad2},
		opts:     Options{MinSavings: 0, Concurrency: 2},
	}

	_, err := e.Scan(context.Background())
	if err == nil {
		t.Fatal("expected error when all scanners fail, got nil")
	}
}

func TestScan_ConcurrentExecution(t *testing.T) {
	const numScanners = 5
	var scanners []scanner.Scanner
	for i := 0; i < numScanners; i++ {
		scanners = append(scanners, &mockScanner{
			name:     fmt.Sprintf("scanner-%d", i),
			category: model.CategoryCompute,
			findings: []model.Finding{{
				ID:             fmt.Sprintf("f-%d", i),
				MonthlySavings: float64(i + 1),
			}},
			delay: 10 * time.Millisecond,
		})
	}

	e := &Engine{
		scanners: scanners,
		opts:     Options{MinSavings: 0, Concurrency: numScanners},
	}

	start := time.Now()
	results, err := e.Scan(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(results) != numScanners {
		t.Fatalf("expected %d findings, got %d", numScanners, len(results))
	}

	// With full concurrency, all scanners run in parallel.
	// Should complete in roughly one delay period, not numScanners * delay.
	maxExpected := time.Duration(numScanners) * 10 * time.Millisecond
	if elapsed >= maxExpected {
		t.Errorf("expected concurrent execution to finish faster than %v, took %v", maxExpected, elapsed)
	}

	// Verify every scanner was called exactly once
	for _, s := range scanners {
		ms := s.(*mockScanner)
		if ms.called.Load() != 1 {
			t.Errorf("scanner %s called %d times, want 1", ms.name, ms.called.Load())
		}
	}
}
