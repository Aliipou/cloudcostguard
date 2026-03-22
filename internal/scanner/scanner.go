package scanner

import (
	"context"

	"github.com/Aliipou/cloudcostguard/internal/model"
)

// Scanner is the interface every cloud resource scanner must implement.
type Scanner interface {
	// Name returns a human-readable name for this scanner.
	Name() string

	// Category returns what type of resources this scanner checks.
	Category() model.Category

	// Scan runs the analysis and returns findings.
	Scan(ctx context.Context) ([]model.Finding, error)
}
