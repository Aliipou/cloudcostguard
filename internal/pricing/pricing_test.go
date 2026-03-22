package pricing

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestEC2MonthlyCost(t *testing.T) {
	tests := []struct {
		name         string
		instanceType string
		region       string
		wantMin      float64
	}{
		{"t3.micro us-east-1", "t3.micro", "us-east-1", 7.59},
		{"m5.large us-east-1", "m5.large", "us-east-1", 70.08},
		{"unknown type", "x1.megabig", "us-east-1", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EC2MonthlyCost(tt.instanceType, tt.region)
			if !almostEqual(got, tt.wantMin) {
				t.Errorf("EC2MonthlyCost(%s, %s) = %.2f, want %.2f", tt.instanceType, tt.region, got, tt.wantMin)
			}
		})
	}
}

func TestEC2MonthlyCostRegionMultiplier(t *testing.T) {
	usEast := EC2MonthlyCost("t3.micro", "us-east-1")
	euWest := EC2MonthlyCost("t3.micro", "eu-west-1")

	if euWest <= usEast {
		t.Errorf("EU pricing (%.2f) should be higher than US pricing (%.2f)", euWest, usEast)
	}
}

func TestRecommendSmaller(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"t3.2xlarge", "t3.xlarge"},
		{"t3.xlarge", "t3.large"},
		{"t3.micro", "t3.micro"}, // no smaller size
		{"unknown.type", "unknown.type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RecommendSmaller(tt.input)
			if got != tt.expected {
				t.Errorf("RecommendSmaller(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEBSMonthlyCost(t *testing.T) {
	// 100GB gp3 should cost $8/month in us-east-1
	cost := EBSMonthlyCost("gp3", 100, 3000, "us-east-1")
	if !almostEqual(cost, 8.0) {
		t.Errorf("EBSMonthlyCost(gp3, 100GB) = %.2f, want 8.00", cost)
	}

	// io1 with IOPS should include IOPS charge
	costIO1 := EBSMonthlyCost("io1", 100, 1000, "us-east-1")
	baseIO1 := 0.125 * 100
	iopsCharge := 1000 * 0.065
	expected := baseIO1 + iopsCharge
	if !almostEqual(costIO1, expected) {
		t.Errorf("EBSMonthlyCost(io1, 100GB, 1000 IOPS) = %.2f, want %.2f", costIO1, expected)
	}
}

func TestRDSMonthlyCost(t *testing.T) {
	singleAZ := RDSMonthlyCost("db.m5.large", "mysql", false, "us-east-1")
	multiAZ := RDSMonthlyCost("db.m5.large", "mysql", true, "us-east-1")

	if !almostEqual(multiAZ, singleAZ*2) {
		t.Errorf("Multi-AZ (%.2f) should be 2x Single-AZ (%.2f)", multiAZ, singleAZ)
	}
}

func TestAzureVMMonthlyCost(t *testing.T) {
	cost := AzureVMMonthlyCost("Standard_B1s", "eastus")
	if !almostEqual(cost, 7.59) {
		t.Errorf("AzureVMMonthlyCost(Standard_B1s) = %.2f, want 7.59", cost)
	}
}

func TestAzureDiskMonthlyCost(t *testing.T) {
	// 128GB Premium_LRS disk
	cost := AzureDiskMonthlyCost("Premium_LRS", 128)
	expected := 0.132 * 128
	if !almostEqual(cost, expected) {
		t.Errorf("AzureDiskMonthlyCost(Premium_LRS, 128) = %.2f, want %.2f", cost, expected)
	}
}
