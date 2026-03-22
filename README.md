# CloudCostGuard

Cloud cost optimization engine that scans AWS and Azure infrastructure to find wasted resources and recommends actions with exact dollar savings.

## What it detects

| Category | AWS | Azure |
|----------|-----|-------|
| **Compute** | Idle/oversized EC2 instances | Idle/oversized VMs |
| **Storage** | Unattached EBS volumes, S3 lifecycle gaps, old versions | Unattached managed disks |
| **Network** | Idle load balancers | Unassociated public IPs |
| **Database** | Idle RDS instances, unnecessary Multi-AZ | — |

## Quick start

```bash
# Install
go install github.com/Aliipou/cloudcostguard@latest

# Or build from source
make build

# Configure
cp cloudcostguard.example.yaml ~/.cloudcostguard.yaml
# Edit with your cloud credentials

# Scan
cloudcostguard scan --provider aws
cloudcostguard scan --provider azure --type compute
cloudcostguard scan --provider aws --min-savings 50 --output json
```

## Output formats

**Table** (default) — human-readable terminal output:
```
SEVERITY   PROVIDER   CATEGORY   TITLE                                      MONTHLY       ANNUAL   EFFORT
--------------------------------------------------------------------------------------------------------------
CRITICAL   aws        compute    Idle EC2 instance: api-server-prod           $560.64    $6727.68   low
HIGH       aws        storage    Unattached EBS volume: old-data (500GB)      $50.00      $600.00   low
MEDIUM     azure      compute    Idle Azure VM: staging-web                   $70.08      $840.96   low
--------------------------------------------------------------------------------------------------------------
TOTAL (3 findings)                                                            $680.72    $8168.64
```

**JSON** — structured output for automation:
```bash
cloudcostguard scan --provider aws --output json | jq '.summary.total_annual_savings'
```

**CSV** — for spreadsheets and reporting:
```bash
cloudcostguard scan --provider aws --output csv > report.csv
```

## Configuration

```yaml
# ~/.cloudcostguard.yaml
provider: aws

aws:
  profile: default
  regions: [us-east-1, us-west-2, eu-west-1]

rules:
  idle_cpu_threshold: 5.0      # % CPU below = idle
  idle_days: 14                # observation window
  unattached_disk_days: 7      # days before flagging
  oversized_cpu_threshold: 20.0
```

See [cloudcostguard.example.yaml](cloudcostguard.example.yaml) for all options.

## Architecture

```
cmd/                    CLI commands (cobra)
internal/
  config/               YAML config loading with defaults
  engine/               Scanner interface + concurrent orchestration
  model/                Finding, Severity, Category types
  pricing/              Cloud pricing lookups (EC2, RDS, EBS, Azure VM/Disk)
  report/               Table, JSON, CSV output formatters
  scanner/
    aws/                EC2, EBS, S3, ELB, RDS scanners
    azure/              VM, Disk, Network scanners
```

## Required permissions

**AWS** — read-only access:
```json
{
  "Effect": "Allow",
  "Action": [
    "ec2:DescribeInstances", "ec2:DescribeVolumes",
    "s3:ListAllMyBuckets", "s3:GetBucketLocation", "s3:GetLifecycleConfiguration",
    "elasticloadbalancing:DescribeLoadBalancers", "elasticloadbalancing:DescribeTargetHealth",
    "rds:DescribeDBInstances",
    "cloudwatch:GetMetricStatistics"
  ],
  "Resource": "*"
}
```

**Azure** — Reader role on the subscription.

## Development

```bash
make test       # Run tests with coverage
make lint       # Run golangci-lint
make build      # Build binary
make docker     # Build Docker image
```

## Docker

```bash
docker run --rm \
  -v ~/.aws:/home/appuser/.aws:ro \
  -v ~/.cloudcostguard.yaml:/home/appuser/.cloudcostguard.yaml:ro \
  cloudcostguard:latest scan --provider aws
```

## License

MIT
