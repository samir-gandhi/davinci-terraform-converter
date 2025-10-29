#!/bin/bash
# Example: Generate environment-specific module with values
# Use Case: Deploy to dev/staging/prod with actual configurations

set -e

ENVIRONMENT=${1:-dev}
OUTPUT_DIR="./envs/${ENVIRONMENT}"

echo "=== Generating Environment-Specific Module: ${ENVIRONMENT} ==="
echo "Output: ${OUTPUT_DIR}"
echo

davinci-convert export \
  --pingone-worker-environment-id "${PINGONE_WORKER_ENVIRONMENT_ID}" \
  --pingone-export-environment-id "${PINGONE_EXPORT_ENVIRONMENT_ID}" \
  --pingone-worker-client-id "${PINGONE_WORKER_CLIENT_ID}" \
  --pingone-worker-client-secret "${PINGONE_WORKER_CLIENT_SECRET}" \
  --pingone-region-code "${PINGONE_REGION_CODE:-NA}" \
  --include-values \
  --include-imports \
  --out "${OUTPUT_DIR}"

echo
echo "=== Environment Configuration Generated ==="
tree "${OUTPUT_DIR}"

echo
echo "=== module.tf Contents ==="
cat "${OUTPUT_DIR}/module.tf"

echo
echo "=== Deployment Steps ==="
cat <<EOF
cd ${OUTPUT_DIR}

# 1. Initialize Terraform
terraform init

# 2. Preview changes (includes import of existing resources)
terraform plan

# 3. Apply configuration (imports + applies changes)
terraform apply

# 4. Verify outputs
terraform output
EOF

echo
echo "=== Notes ==="
echo "- Variables populated with actual values from API"
echo "- Import blocks included for existing resources"
echo "- Secrets marked as TODO - provide values before apply"
echo "- Review module.tf for environment-specific configuration"
