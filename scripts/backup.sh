#!/usr/bin/env bash
# backup.sh - Export CloudCostGuard scan results to timestamped JSON and rotate old backups.
# Usage: ./scripts/backup.sh [backup_dir] [results_dir]
# Cron example (daily at 2 AM): 0 2 * * * /path/to/cloudcostguard/scripts/backup.sh

set -euo pipefail

BACKUP_DIR="${1:-${HOME}/cloudcostguard-backups}"
RESULTS_DIR="${2:-./results}"
RETENTION_DAYS=30
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "${BACKUP_DIR}"

# Export scan results to a timestamped JSON file
if [ -d "${RESULTS_DIR}" ] && ls "${RESULTS_DIR}"/*.json >/dev/null 2>&1; then
    DEST="${BACKUP_DIR}/scan-${TIMESTAMP}.json"
    # Merge all result files into a single backup
    if command -v jq >/dev/null 2>&1; then
        jq -s 'add' "${RESULTS_DIR}"/*.json > "${DEST}"
    else
        # Fallback: concatenate as JSON array without jq
        echo "[" > "${DEST}"
        first=true
        for f in "${RESULTS_DIR}"/*.json; do
            if [ "${first}" = true ]; then
                first=false
            else
                echo "," >> "${DEST}"
            fi
            cat "${f}" >> "${DEST}"
        done
        echo "]" >> "${DEST}"
    fi
    echo "Backup created: ${DEST}"
else
    echo "No results found in ${RESULTS_DIR}, skipping export."
fi

# Rotate old backups - remove files older than RETENTION_DAYS
find "${BACKUP_DIR}" -name "scan-*.json" -type f -mtime +"${RETENTION_DAYS}" -delete 2>/dev/null || true
REMAINING="$(find "${BACKUP_DIR}" -name "scan-*.json" -type f | wc -l)"
echo "Rotation complete. ${REMAINING} backup(s) retained."
