#!/usr/bin/env bash
#
# Example: Export complete DaVinci environment from PingOne API
#
# This demonstrates API-based export using the export command.
# Exports all resources (flows, connectors, variables, applications, policies)
# with automatic dependency resolution.
#

set -e

BINARY="${DAVINCI_CONVERT_BINARY:-davinci-convert}"

# Check if binary exists
if ! command -v "${BINARY}" &> /dev/null; then
    echo "Error: ${BINARY} not found in PATH"
    echo "Build with: make install"
    exit 1
fi

# Check required environment variables
required_vars=(
    "PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID"
    "PINGCLI_PINGONE_WORKER_CLIENT_ID"
    "PINGCLI_PINGONE_WORKER_CLIENT_SECRET"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: Required environment variable ${var} not set"
        echo ""
        echo "Set up environment variables:"
        echo "  export PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID=\"your-worker-env-id\""
        echo "  export PINGCLI_PINGONE_WORKER_CLIENT_ID=\"your-client-id\""
        echo "  export PINGCLI_PINGONE_WORKER_CLIENT_SECRET=\"your-client-secret\""
        echo "  export PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID=\"your-target-env-id\"  # Optional"
        echo "  export PINGCLI_PINGONE_REGION_CODE=\"NA\"  # Optional, defaults to NA"
        exit 1
    fi
done

# Example 1: Export to stdout
echo "=== Example 1: Export environment to stdout ==="
"${BINARY}" export

echo ""
echo "=== Example 2: Export environment to file ==="
OUTPUT_FILE="/tmp/environment-export.tf"
"${BINARY}" export --out "${OUTPUT_FILE}"

echo "Export complete!"
echo "Output file: ${OUTPUT_FILE}"
echo ""
echo "File statistics:"
wc -l "${OUTPUT_FILE}"
echo ""
echo "Resource counts:"
grep -c '^resource "' "${OUTPUT_FILE}" || echo "0 resources"
grep -c 'pingone_davinci_variable' "${OUTPUT_FILE}" || echo "0 variables"
grep -c 'pingone_davinci_connector_instance' "${OUTPUT_FILE}" || echo "0 connector instances"
grep -c 'pingone_davinci_flow' "${OUTPUT_FILE}" || echo "0 flows"
grep -c 'pingone_davinci_application' "${OUTPUT_FILE}" || echo "0 applications"
grep -c 'pingone_davinci_application_flow_policy' "${OUTPUT_FILE}" || echo "0 flow policies"

echo ""
echo "=== Example 3: Export with skip-dependencies ==="
OUTPUT_FILE_SKIP_DEPS="/tmp/environment-export-skip-deps.tf"
"${BINARY}" export --skip-dependencies --out "${OUTPUT_FILE_SKIP_DEPS}"

echo "Export with skip-dependencies complete!"
echo "Output file: ${OUTPUT_FILE_SKIP_DEPS}"
echo ""
echo "Comparison:"
echo "With dependencies:"
grep -m1 'environment_id' "${OUTPUT_FILE}" || echo "(none found)"
echo ""
echo "Without dependencies (hardcoded):"
grep -m1 'environment_id' "${OUTPUT_FILE_SKIP_DEPS}" || echo "(none found)"

echo ""
echo "Next steps:"
echo "1. Review generated HCL: cat ${OUTPUT_FILE}"
echo "2. Create variables.tf with required variables"
echo "3. Create provider.tf with PingOne provider configuration"
echo "4. Run: terraform init"
echo "5. Run: terraform plan"
