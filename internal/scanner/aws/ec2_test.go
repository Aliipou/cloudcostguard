package aws

import (
	"testing"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/model"
)

func TestEC2Scanner_Name(t *testing.T) {
	s := NewEC2Scanner(config.DefaultConfig(), "eu-west-1")
	want := "aws-ec2-eu-west-1"
	if got := s.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestEC2Scanner_NameDefaultRegion(t *testing.T) {
	s := NewEC2Scanner(config.DefaultConfig(), "us-east-1")
	want := "aws-ec2-us-east-1"
	if got := s.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestEC2Scanner_Category(t *testing.T) {
	s := NewEC2Scanner(config.DefaultConfig(), "us-east-1")
	if got := s.Category(); got != model.CategoryCompute {
		t.Errorf("Category() = %q, want %q", got, model.CategoryCompute)
	}
}
