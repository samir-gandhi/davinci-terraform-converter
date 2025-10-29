#!/bin/bash
# Example: Legacy single-file export (no module structure)
# Use Case: Backwards compatibility, simple exports, custom workflows

set -e

OUTPUT_FILE="davinci-resources.tf"

echo "=== Generating Legacy Single-File Export ==="
echo "Output: ${OUTPUT_FILE}"
echo

davinci-convert export \
  --pingone-worker-environment-id "${PINGONE_WORKER_ENVIRONMENT_ID}" \
  --pingone-export-environment-id "${PINGONE_EXPORT_ENVIRONMENT_ID}" \
  --pingone-worker-client-id "${PINGONE_WORKER_CLIENT_ID}" \
  --pingone-worker-client-secret "${PINGONE_WORKER_CLIENT_SECRET}" \
  --pingone-region-code "${PINGONE_REGION_CODE:-NA}" \
  --module=false \
  --out "${OUTPUT_FILE}"

echo
echo "=== Single File Generated ==="
ls -lh "${OUTPUT_FILE}"

echo
echo "=== File Structure ==="
head -n 50 "${OUTPUT_FILE}"
echo "..."
echo "(truncated for brevity)"

echo
echo "=== When to Use Legacy Mode ==="
echo "- Converting to existing Terraform configuration (no modules)"
echo "- Quick exports for analysis or documentation"
echo "- Custom module structure requirements"
echo "- Integration with legacy tooling expecting single file"

echo
echo "=== Migration Path ==="
echo "To migrate from legacy to module mode:"
echo "1. Export with module mode: davinci-convert export --out ./modules"
echo "2. Review generated module structure in davinci-module/"
echo "3. Update references from direct resources to module.davinci.resource_name"
echo "4. Test with 'terraform plan' to verify no changes"
