#!/usr/bin/env bash
#
# Example: Using davinci-convert as a PingCLI plugin
#
# This demonstrates plugin mode integration with pingcli.
# Commands are namespaced under 'pingcli tf'.
#

set -e

PINGCLI="${PINGCLI_BINARY:-pingcli}"

# Check if pingcli exists
if ! command -v "${PINGCLI}" &> /dev/null; then
    echo "Error: ${PINGCLI} not found in PATH"
    echo "Install from: https://github.com/pingidentity/pingcli"
    exit 1
fi

# Check if plugin is installed
if ! "${PINGCLI}" tf --help &> /dev/null; then
    echo "Error: davinci-convert plugin not found by pingcli"
    echo ""
    echo "Build and install plugin:"
    echo "  cd <davinci-terraform-converter>"
    echo "  make install"
    echo ""
    echo "Ensure binary is in PATH and named 'pingcli-tf-davinci-convert'"
    exit 1
fi

echo "=== PingCLI Plugin Mode Examples ==="
echo ""

# Example 1: Convert flow file
echo "=== Example 1: Convert flow JSON to HCL ==="
FLOW_JSON="${1:-../. github/prompts/simple-demo-flow.json}"

if [ ! -f "${FLOW_JSON}" ]; then
    echo "Error: Flow JSON file not found: ${FLOW_JSON}"
    echo "Usage: $0 <path-to-flow.json>"
    exit 1
fi

"${PINGCLI}" tf davinci-to-hcl --flow-json "${FLOW_JSON}"

echo ""
echo "=== Example 2: Convert flow to file ==="
"${PINGCLI}" tf davinci-to-hcl \
    --flow-json "${FLOW_JSON}" \
    --out /tmp/pingcli-flow-output.tf

echo "Output written to: /tmp/pingcli-flow-output.tf"

echo ""
echo "=== Example 3: Export environment via API ==="

# Check for required environment variables
if [ -z "${PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID}" ]; then
    echo "Skipping API export example (credentials not set)"
    echo ""
    echo "To run API export with pingcli:"
    echo "  export PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID=\"...\""
    echo "  export PINGCLI_PINGONE_WORKER_CLIENT_ID=\"...\""
    echo "  export PINGCLI_PINGONE_WORKER_CLIENT_SECRET=\"...\""
    echo "  ${PINGCLI} tf export --out environment.tf"
else
    "${PINGCLI}" tf export --out /tmp/pingcli-environment-export.tf
    echo "Export complete: /tmp/pingcli-environment-export.tf"
    wc -l /tmp/pingcli-environment-export.tf
fi

echo ""
echo "Plugin Mode Benefits:"
echo "  ✓ Unified CLI experience with other Ping tools"
echo "  ✓ Consistent logging and error handling"
echo "  ✓ Integration with pingcli configuration"
echo "  ✓ Automatic plugin lifecycle management"
