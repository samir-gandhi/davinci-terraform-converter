#!/usr/bin/env bash
#
# Example: Export from one environment using worker app in another
#
# This demonstrates the two-environment authentication model:
# - Worker environment: Contains OAuth2 worker app for authentication
# - Export environment: Contains target resources to export
#
# Use case: Export production environment using shared worker app
#

set -e

BINARY="${DAVINCI_CONVERT_BINARY:-davinci-convert}"

# Check if binary exists
if ! command -v "${BINARY}" &> /dev/null; then
    echo "Error: ${BINARY} not found in PATH"
    exit 1
fi

# Two-environment configuration
# These can be set as environment variables or passed as flags

echo "=== Two-Environment Export Example ==="
echo ""
echo "Scenario: Export production environment using shared worker app"
echo "  - Worker Environment: ${PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID:-<not set>}"
echo "  - Export Environment: ${PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID:-<not set>}"
echo ""

# Check worker environment variables
if [ -z "${PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID}" ]; then
    echo "Error: PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID not set"
    exit 1
fi

if [ -z "${PINGCLI_PINGONE_WORKER_CLIENT_ID}" ]; then
    echo "Error: PINGCLI_PINGONE_WORKER_CLIENT_ID not set"
    exit 1
fi

if [ -z "${PINGCLI_PINGONE_WORKER_CLIENT_SECRET}" ]; then
    echo "Error: PINGCLI_PINGONE_WORKER_CLIENT_SECRET not set"
    exit 1
fi

# Example 1: Export using environment variables
if [ -n "${PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID}" ]; then
    echo "=== Export using environment variables ==="
    OUTPUT_FILE="/tmp/two-env-export.tf"
    "${BINARY}" export --out "${OUTPUT_FILE}"
    
    echo "Export complete: ${OUTPUT_FILE}"
    wc -l "${OUTPUT_FILE}"
fi

echo ""
echo "=== Export using explicit flags ==="

# Prompt for export environment if not set
if [ -z "${EXPORT_ENV_ID}" ]; then
    read -p "Enter export environment ID: " EXPORT_ENV_ID
fi

OUTPUT_FILE="/tmp/two-env-export-explicit.tf"

"${BINARY}" export \
    --pingone-worker-environment-id "${PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID}" \
    --pingone-export-environment-id "${EXPORT_ENV_ID}" \
    --pingone-worker-client-id "${PINGCLI_PINGONE_WORKER_CLIENT_ID}" \
    --pingone-worker-client-secret "${PINGCLI_PINGONE_WORKER_CLIENT_SECRET}" \
    --pingone-region-code "${PINGCLI_PINGONE_REGION_CODE:-NA}" \
    --out "${OUTPUT_FILE}"

echo "Export complete: ${OUTPUT_FILE}"
wc -l "${OUTPUT_FILE}"

echo ""
echo "Benefits of two-environment model:"
echo "  ✓ Worker app credentials isolated from exported resources"
echo "  ✓ Single worker app can export from multiple environments"
echo "  ✓ Supports dev/staging/prod workflows"
echo "  ✓ Easier credential rotation (only update worker app)"
