# Security Policy

## Required Permissions

CloudCostGuard requires **read-only** access to cloud resources. It never creates, modifies, or deletes infrastructure.

**AWS** -- Minimum IAM permissions:
- `ec2:DescribeInstances`, `ec2:DescribeVolumes`
- `elasticloadbalancing:DescribeLoadBalancers`
- `rds:DescribeDBInstances`
- `s3:ListBuckets`, `s3:GetBucketMetricsConfiguration`
- `cloudwatch:GetMetricStatistics`

**Azure** -- Minimum RBAC role:
- `Reader` on the target subscription (built-in role)
- `Monitoring Reader` for metrics access

## Credential Handling

- CloudCostGuard **does not store credentials**. It relies on the standard provider SDKs (`aws-sdk-go-v2`, `azure-sdk-for-go`) which read credentials from environment variables, config files, or instance metadata.
- Never pass credentials via CLI flags or config files checked into version control.
- Use short-lived credentials (IAM roles, Azure Managed Identity) whenever possible.
- Scan results may contain resource IDs and tags. Treat output files as sensitive.

## Reporting a Vulnerability

If you discover a security issue, please email **security@cloudcostguard.dev** with:
1. A description of the vulnerability
2. Steps to reproduce
3. Any relevant logs or screenshots

We will acknowledge receipt within 48 hours and aim to provide a fix or mitigation within 7 days. Please do not open a public issue for security vulnerabilities.
