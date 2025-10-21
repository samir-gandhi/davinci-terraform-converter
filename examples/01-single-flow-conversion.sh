#!/usr/bin/env bash
#
# Example: Convert a single DaVinci flow JSON file to HCL
#
# This demonstrates file-based conversion using the davinci-to-hcl command.
# Use this when you have exported flow JSON files from DaVinci UI.
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${DAVINCI_CONVERT_BINARY:-davinci-convert}"

# Check if binary exists
if ! command -v "${BINARY}" &> /dev/null; then
    echo "Error: ${BINARY} not found in PATH"
    echo "Build with: make install"
    echo "Or set DAVINCI_CONVERT_BINARY environment variable"
    exit 1
fi

# Example 1: Convert to stdout
echo "=== Example 1: Convert flow to stdout ==="
"${BINARY}" davinci-to-hcl \
    --flow-json "${SCRIPT_DIR}/../.github/prompts/simple-demo-flow.json"

echo ""
echo "=== Example 2: Convert flow to file ==="
"${BINARY}" davinci-to-hcl \
    --flow-json "${SCRIPT_DIR}/../.github/prompts/simple-demo-flow.json" \
    --out /tmp/output-flow.tf

echo "Output written to: /tmp/output-flow.tf"
cat /tmp/output-flow.tf

echo ""
echo "=== Example 3: Convert with skip-dependencies ==="
"${BINARY}" davinci-to-hcl \
    --flow-json "${SCRIPT_DIR}/../.github/prompts/simple-demo-flow.json" \
    --skip-dependencies \
    --out /tmp/output-flow-skip-deps.tf

echo "Output written to: /tmp/output-flow-skip-deps.tf"
echo "Note: environment_id will be hardcoded if available in JSON"
