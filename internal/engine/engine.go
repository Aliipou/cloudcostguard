package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
	"github.com/Aliipou/cloudcostguard/internal/scanner"
	"github.com/Aliipou/cloudcostguard/internal/scanner/aws"
	"github.com/Aliipou/cloudcostguard/internal/scanner/azure"
)

// Options configures a scan run.
type Options struct {
	ScanType    string
	MinSavings  float64
	Concurrency int
	Verbose     bool
}

// Engine orchestrates scanners and collects findings.
type Engine struct {
	scanners []scanner.Scanner
	opts     Options
}

// New creates an Engine with the appropriate scanners for the given provider.
func New(cfg *config.Config, opts Options) (*Engine, error) {
	var scanners []scanner.Scanner

	switch cfg.Provider {
	case "aws":
		s, err := aws.NewScanners(cfg, opts.ScanType)
		if err != nil {
			return nil, fmt.Errorf("initializing AWS scanners: %w", err)
		}
		scanners = s
	case "azure":
		s, err := azure.NewScanners(cfg, opts.ScanType)
		if err != nil {
			return nil, fmt.Errorf("initializing Azure scanners: %w", err)
		}
		scanners = s
	default:
		return nil, fmt.Errorf("unsupported provider: %s (supported: aws, azure)", cfg.Provider)
	}

	if len(scanners) == 0 {
		return nil, fmt.Errorf("no scanners matched type %q for provider %s", opts.ScanType, cfg.Provider)
	}

	return &Engine{scanners: scanners, opts: opts}, nil
}

// Scan runs all scanners concurrently and returns aggregated findings.
func (e *Engine) Scan(ctx context.Context) ([]model.Finding, error) {
	type result struct {
		findings []model.Finding
		err      error
	}

	sem := make(chan struct{}, e.opts.Concurrency)
	results := make(chan result, len(e.scanners))

	var wg sync.WaitGroup
	for _, s := range e.scanners {
		wg.Add(1)
		go func(sc scanner.Scanner) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			findings, err := sc.Scan(ctx)
			results <- result{findings: findings, err: err}
		}(s)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allFindings []model.Finding
	var errors []error

	for r := range results {
		if r.err != nil {
			errors = append(errors, r.err)
			continue
		}
		for _, f := range r.findings {
			if f.MonthlySavings >= e.opts.MinSavings {
				allFindings = append(allFindings, f)
			}
		}
	}

	if len(errors) > 0 && len(allFindings) == 0 {
		return nil, fmt.Errorf("all scanners failed, first error: %w", errors[0])
	}

	return allFindings, nil
}
