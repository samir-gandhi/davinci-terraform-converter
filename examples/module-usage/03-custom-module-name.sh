#!/bin/bash
# Example: Custom module directory name
# Use Case: Organize modules with descriptive names

set -e

MODULE_NAME=${1:-my-auth-flows}
OUTPUT_DIR="./modules"

echo "=== Generating Custom Named Module: ${MODULE_NAME} ==="
echo "Output: ${OUTPUT_DIR}/${MODULE_NAME}"
echo

davinci-convert export \
  --pingone-worker-environment-id "${PINGONE_WORKER_ENVIRONMENT_ID}" \
  --pingone-export-environment-id "${PINGONE_EXPORT_ENVIRONMENT_ID}" \
  --pingone-worker-client-id "${PINGONE_WORKER_CLIENT_ID}" \
  --pingone-worker-client-secret "${PINGONE_WORKER_CLIENT_SECRET}" \
  --pingone-region-code "${PINGONE_REGION_CODE:-NA}" \
  --module-dir "${MODULE_NAME}" \
  --out "${OUTPUT_DIR}"

echo
echo "=== Module Structure ==="
tree "${OUTPUT_DIR}"

echo
echo "=== Using Custom Module Name ==="
cat <<EOF
# module.tf generated with custom source path:
module "davinci" {
  source = "./${MODULE_NAME}"
  
  pingone_environment_id = "your-env-id"
}
EOF

echo
echo "=== Use Cases for Custom Names ==="
echo "- my-auth-flows: Authentication-focused flows"
echo "- registration-flows: User registration flows"
echo "- mfa-policies: Multi-factor authentication policies"
echo "- partner-integrations: Third-party connector configurations"
