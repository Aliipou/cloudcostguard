<div align="center">

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![AWS](https://img.shields.io/badge/AWS-supported-FF9900?style=flat&logo=amazonaws)](https://aws.amazon.com)
[![Azure](https://img.shields.io/badge/Azure-supported-0078D4?style=flat&logo=microsoftazure)](https://azure.microsoft.com)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)
[![CI](https://github.com/Aliipou/cloudcostguard/actions/workflows/ci.yml/badge.svg)](https://github.com/Aliipou/cloudcostguard/actions)

**Cloud cost optimization engine for AWS and Azure.**

Scans your infrastructure, finds wasted resources, and reports exact monthly and annual savings per resource.

</div>

## What it detects

| Category | AWS | Azure |
|----------|-----|-------|
| **Compute** | Idle EC2 instances, oversized instance types | Idle VMs |
| **Storage** | Unattached EBS volumes, S3 lifecycle gaps, old versions | Unattached managed disks |
| **Network** | Idle load balancers with no healthy targets | Unassociated public IPs |
| **Database** | Idle RDS instances, unnecessary Multi-AZ deployments | — |

## Architecture

```
cloudcostguard scan --provider aws
         |
         v
    [Engine]               Concurrent scanner orchestration
     /      \
[AWS]      [Azure]         Provider-specific scanners
  |          |
  EC2  EBS   VM  Disk      Resource scanners (per type)
  S3   ELB   Network
  RDS
         |
         v
    [Pricing]              Per-resource cost lookup + region multipliers
         |
         v
  [Report]                 Table / JSON / CSV output with $ savings
```

## Quick start

```bash
# Install
go install github.com/Aliipou/cloudcostguard@latest

# Or build from source
git clone https://github.com/Aliipou/cloudcostguard.git
cd cloudcostguard && make build

# Configure
cp cloudcostguard.example.yaml ~/.cloudcostguard.yaml
# Edit with your subscription ID / AWS profile

# Scan
cloudcostguard scan --provider aws
cloudcostguard scan --provider azure --type compute
cloudcostguard scan --provider aws --min-savings 50 --output json

# Web dashboard + Prometheus metrics
cloudcostguard serve --port 8080
# Open http://localhost:8080
```

## Output formats

**Table** (default):
```
SEVERITY   PROVIDER   CATEGORY   TITLE                                      MONTHLY       ANNUAL   EFFORT
--------------------------------------------------------------------------------------------------------------
CRITICAL   aws        compute    Idle EC2 instance: api-server-prod           $560.64    $6727.68   low
HIGH       aws        storage    Unattached EBS volume: old-data (500GB)      $50.00      $600.00   low
MEDIUM     azure      compute    Idle Azure VM: staging-web                   $70.08      $840.96   low
--------------------------------------------------------------------------------------------------------------
TOTAL (3 findings)                                                            $680.72    $8168.64
```

**JSON** — for automation and pipelines:
```bash
cloudcostguard scan --provider aws --output json | jq '.summary.total_annual_savings'
```

**CSV** — for spreadsheets:
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
  idle_cpu_threshold: 5.0       # % CPU below = idle
  idle_days: 14                 # observation window in days
  unattached_disk_days: 7       # days before flagging unattached disks
  oversized_cpu_threshold: 20.0 # % CPU below = oversized
```

See [cloudcostguard.example.yaml](cloudcostguard.example.yaml) for all options.

## Monitoring

Start the full stack (app + Prometheus + Grafana):

```bash
docker compose up -d
```

| Service | URL |
|---------|-----|
| Dashboard | http://localhost:8080 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 |

Grafana is pre-provisioned with a CloudCostGuard dashboard showing findings by severity, savings by category, and scan duration histograms.

## Required permissions

**AWS** — attach this read-only policy:
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

**Azure** — Reader role on the subscription is sufficient.

## Development

```bash
make test    # run tests with coverage
make lint    # golangci-lint
make build   # build binary
make docker  # build Docker image
```

## Project structure

```
cmd/                    CLI entry points (scan, serve, version)
internal/
  api/                  HTTP server + web dashboard handler
  config/               YAML config loading with defaults
  engine/               Concurrent scanner orchestration
  metrics/              Prometheus-compatible metrics (stdlib only)
  model/                Finding, Severity, Category types
  pricing/              Cloud pricing tables + region multipliers
  report/               Table, JSON, CSV formatters
  scanner/
    aws/                EC2, EBS, S3, ELB, RDS scanners
    azure/              VM, Disk, Network scanners
web/                    Embedded single-file dashboard
monitoring/             Prometheus + Grafana configs
```

## License

MIT
