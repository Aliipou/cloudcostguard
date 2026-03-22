package pricing

// EC2 on-demand pricing (USD/month, us-east-1 baseline).
// In production, these would be fetched from the AWS Pricing API.
var ec2Pricing = map[string]float64{
	"t3.micro":    7.59,
	"t3.small":    15.18,
	"t3.medium":   30.37,
	"t3.large":    60.74,
	"t3.xlarge":   121.47,
	"t3.2xlarge":  242.94,
	"m5.large":    70.08,
	"m5.xlarge":   140.16,
	"m5.2xlarge":  280.32,
	"m5.4xlarge":  560.64,
	"m6i.large":   69.35,
	"m6i.xlarge":  138.70,
	"m6i.2xlarge": 277.40,
	"c5.large":    62.05,
	"c5.xlarge":   124.10,
	"c5.2xlarge":  248.20,
	"r5.large":    91.98,
	"r5.xlarge":   183.96,
	"r5.2xlarge":  367.92,
}

// downsizeMap maps instance types to the next smaller size.
var downsizeMap = map[string]string{
	"t3.2xlarge":  "t3.xlarge",
	"t3.xlarge":   "t3.large",
	"t3.large":    "t3.medium",
	"t3.medium":   "t3.small",
	"t3.small":    "t3.micro",
	"m5.4xlarge":  "m5.2xlarge",
	"m5.2xlarge":  "m5.xlarge",
	"m5.xlarge":   "m5.large",
	"m6i.2xlarge": "m6i.xlarge",
	"m6i.xlarge":  "m6i.large",
	"c5.2xlarge":  "c5.xlarge",
	"c5.xlarge":   "c5.large",
	"r5.2xlarge":  "r5.xlarge",
	"r5.xlarge":   "r5.large",
}

// EC2MonthlyCost returns estimated monthly cost for an EC2 instance type.
func EC2MonthlyCost(instanceType, region string) float64 {
	if cost, ok := ec2Pricing[instanceType]; ok {
		return cost * regionMultiplier(region)
	}
	return 100.0 // conservative default
}

// RecommendSmaller returns the next smaller instance type.
func RecommendSmaller(instanceType string) string {
	if smaller, ok := downsizeMap[instanceType]; ok {
		return smaller
	}
	return instanceType
}

// EBS pricing per GB/month by volume type.
var ebsPricing = map[string]float64{
	"gp3":      0.08,
	"gp2":      0.10,
	"io1":      0.125,
	"io2":      0.125,
	"st1":      0.045,
	"sc1":      0.015,
	"standard": 0.05,
}

// EBSMonthlyCost returns estimated monthly cost for an EBS volume.
func EBSMonthlyCost(volumeType string, sizeGB, iops int, region string) float64 {
	perGB, ok := ebsPricing[volumeType]
	if !ok {
		perGB = 0.10
	}
	cost := perGB * float64(sizeGB)

	// io1/io2 have per-IOPS charges
	if volumeType == "io1" || volumeType == "io2" {
		cost += float64(iops) * 0.065
	}

	return cost * regionMultiplier(region)
}

// RDS pricing (simplified, on-demand Single-AZ).
var rdsPricing = map[string]float64{
	"db.t3.micro":    12.41,
	"db.t3.small":    24.82,
	"db.t3.medium":   49.64,
	"db.m5.large":    124.10,
	"db.m5.xlarge":   248.20,
	"db.m5.2xlarge":  496.40,
	"db.r5.large":    175.20,
	"db.r5.xlarge":   350.40,
	"db.r5.2xlarge":  700.80,
}

// RDSMonthlyCost returns estimated monthly cost for an RDS instance.
func RDSMonthlyCost(instanceClass, engine string, multiAZ bool, region string) float64 {
	cost, ok := rdsPricing[instanceClass]
	if !ok {
		cost = 200.0
	}
	if multiAZ {
		cost *= 2
	}
	return cost * regionMultiplier(region)
}

// Azure VM pricing (simplified, pay-as-you-go).
var azureVMPricing = map[string]float64{
	"Standard_B1s":   7.59,
	"Standard_B2s":   30.37,
	"Standard_D2s_v3": 70.08,
	"Standard_D4s_v3": 140.16,
	"Standard_D8s_v3": 280.32,
	"Standard_E2s_v3": 91.98,
	"Standard_E4s_v3": 183.96,
	"Standard_F2s_v2": 61.32,
	"Standard_F4s_v2": 122.64,
}

// AzureVMMonthlyCost returns estimated monthly cost for an Azure VM.
func AzureVMMonthlyCost(vmSize, location string) float64 {
	if cost, ok := azureVMPricing[vmSize]; ok {
		return cost * regionMultiplier(location)
	}
	return 100.0
}

// Azure managed disk pricing.
var azureDiskPricing = map[string]float64{
	"Standard_LRS":    0.04,  // per GB/month
	"StandardSSD_LRS": 0.075,
	"Premium_LRS":     0.132,
	"UltraSSD_LRS":    0.12,
}

// AzureDiskMonthlyCost returns estimated monthly cost for an Azure managed disk.
func AzureDiskMonthlyCost(sku string, sizeGB int) float64 {
	perGB, ok := azureDiskPricing[sku]
	if !ok {
		perGB = 0.10
	}
	return perGB * float64(sizeGB)
}

// regionMultiplier adjusts pricing for different regions.
func regionMultiplier(region string) float64 {
	multipliers := map[string]float64{
		"us-east-1":      1.0,
		"us-west-2":      1.0,
		"eu-west-1":      1.08,
		"eu-central-1":   1.10,
		"ap-southeast-1": 1.12,
		"ap-northeast-1": 1.15,
		"eastus":         1.0,
		"westeurope":     1.08,
		"northeurope":    1.05,
	}
	if m, ok := multipliers[region]; ok {
		return m
	}
	return 1.0
}
