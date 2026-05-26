# Mastery Engineering Audit — CloudCostGuard

> "Simplicity under pressure = mastery"

## Project Summary
A CLI and web-based cloud cost optimization engine for AWS and Azure. Concurrently scans infrastructure (EC2, EBS, S3, RDS, ELB, Azure VMs, managed disks, public IPs), applies pricing lookups and region multipliers, and reports idle/oversized resources with exact monthly and annual savings estimates. Includes a Prometheus metrics endpoint and a web dashboard.

## Failure Mode Analysis

### What breaks when the main service goes down?
CloudCostGuard is primarily a CLI tool — scans are on-demand, not continuous. A process crash loses the current scan result but does not affect infrastructure. The `serve` mode exposes a web dashboard and metrics endpoint; if that process goes down, the dashboard is unavailable but no data is lost (scans are stateless reads against cloud APIs). There is no persistent state that could become corrupted.

### What breaks when the database/storage slows down?
CloudCostGuard has no database. All scan results are computed in-memory and either printed to stdout or returned via the HTTP API. The only external dependencies are the AWS/Azure cloud APIs. A slow cloud API response blocks the relevant scanner goroutine; the semaphore-based concurrency control prevents goroutine explosion, but a scan will take proportionally longer.

### What breaks when the network is partitioned?
If the network to AWS or Azure APIs is partitioned, the affected provider scanners return errors. The engine aggregates: if all scanners fail, the scan returns an error; if some fail and some succeed, partial findings are returned (the partial-failure path is explicit in engine.go). The web dashboard continues serving the last cached scan result. There is no retry logic for individual scanner failures — a transient API blip causes an entire scanner to be skipped.

### What breaks under duplicate execution?
Two simultaneous scans of the same infrastructure are safe — both are read-only against cloud APIs. The `activeScans` Prometheus gauge may be inconsistent if a scan panics without decrementing, but there is no data corruption risk. Running two `cloudcostguard serve` instances on the same port will fail cleanly on bind.

## Checklist Status

| Check | Status | Notes |
|-------|--------|-------|
| Invariants defined | ✅ | Scan is purely read-only; no infrastructure state is mutated — invariants are trivially maintained |
| Idempotency | ✅ | Every scan is fully idempotent — re-running produces the same output for the same cloud state |
| Race conditions handled | ✅ | Engine uses a semaphore channel for concurrency control; WaitGroup + results channel pattern is correct |
| State consistency | ✅ | No mutable shared state between scans; Prometheus metrics collector uses sync.Mutex correctly |
| Structured logging | ⚠️ | No visible structured logging (zap/slog) in engine or scanner code — errors are returned, not logged with context fields |
| Metrics (not just logs) | ✅ | Custom Prometheus-compatible metrics collector (scans_total, active_scans, findings_total by severity/category, scan_duration histogram) |
| Distributed tracing | ❌ | No OpenTelemetry tracing — for a CLI tool this is acceptable, but the serve mode has no request tracing |
| Rollback strategy | ✅ | No database, no state — rollback is simply running the previous binary version |
| Safe migrations | ✅ | No migrations needed — stateless tool |
| 10x traffic plan | ⚠️ | At 10x the number of AWS resources, scan time grows linearly; no pagination or incremental scanning implemented |
| Bottleneck identified | ⚠️ | Cloud API rate limits (AWS ~100 req/s for EC2 Describe) are the primary bottleneck; no rate-limit handling or backoff implemented |
| Simplicity test passed | ✅ | The core value (scan -> price -> report) works end-to-end without a database, queue, or external state |

## Critical Gaps (Must Fix)

1. **No retry/backoff on cloud API errors**: a single transient 429 or 503 from AWS/Azure silently skips that entire scanner and the finding is lost. Implement exponential backoff with jitter for cloud API calls.
2. **No structured logging**: scanner errors are returned as Go errors and logged generically (if at all). Add zap/slog with fields like `provider`, `scanner_type`, `region`, `resource_id` so failed scans are debuggable.
3. **Pricing data is hardcoded**: if AWS/Azure changes pricing, the tool produces incorrect savings estimates silently. Add a last-updated timestamp to pricing data and warn when it is older than 90 days.
4. **No pagination in resource listing**: large AWS accounts with thousands of EC2 instances or EBS volumes may hit API result limits and return incomplete findings without any indication of truncation.
5. **activeScans gauge leaks on panic**: if a scanner goroutine panics, `DecActiveScans` is never called. Wrap scanner goroutines in a recover and ensure the gauge is always decremented.

## What is Already Mastery-Level

- **Concurrent scanner orchestration**: the semaphore + WaitGroup + results channel pattern in `engine.go` is clean, correct, and prevents goroutine leaks even when individual scanners fail.
- **Partial failure handling**: if some scanners succeed and some fail, the engine returns the successful findings rather than failing the entire scan — operationally pragmatic.
- **Custom Prometheus metrics without a dependency**: the `metrics` package implements Prometheus text exposition format using only the standard library — zero external dependency for observability.
- **Histogram for scan duration**: `scan_duration_seconds` is exposed as a proper histogram (not just a gauge or counter), enabling SLO tracking on scan latency.
- **MinSavings filter**: the `--min-savings` flag prevents noise from sub-dollar findings, making the output actionable rather than overwhelming.
- **golangci-lint configuration**: the `.golangci.yml` file shows a mature code quality mindset baked into the development workflow.

## Recommended Next Steps

1. **Add retry with exponential backoff**: wrap all AWS/Azure API calls in a retry loop (3 attempts, jitter backoff) to handle transient rate limits and 5xx responses gracefully.
2. **Pricing freshness check**: embed a build timestamp in pricing tables and emit a warning (`WARN: pricing data is X days old`) when running with stale data.
3. **Paginated resource listing**: implement cursor/token-based pagination for all AWS Describe* and Azure List* calls to handle accounts with large resource counts.
4. **Structured logging**: replace fmt.Errorf chains with zap logger, attaching provider, region, and resource type to every error log entry.
5. **Scan result caching in serve mode**: cache the last scan result with a timestamp so the dashboard serves stale-but-valid data while a new scan is in progress, rather than showing an empty state.
