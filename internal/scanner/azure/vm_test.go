package azure

import (
	"testing"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

func TestVMScanner_Name(t *testing.T) {
	s := NewVMScanner(config.DefaultConfig(), "sub-123")
	want := "azure-vm"
	if got := s.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestVMScanner_Category(t *testing.T) {
	s := NewVMScanner(config.DefaultConfig(), "sub-123")
	if got := s.Category(); got != model.CategoryCompute {
		t.Errorf("Category() = %q, want %q", got, model.CategoryCompute)
	}
}
