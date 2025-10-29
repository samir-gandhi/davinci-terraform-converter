#!/bin/bash
# Example: Generate a shareable Terraform module for DaVinci resources
# Use Case: Create reusable module for multiple environments or teams

set -e

echo "=== Generating Shareable DaVinci Module ==="
echo "Output: ./shareable-module/"
echo

davinci-convert export \
  --pingone-worker-environment-id "${PINGONE_WORKER_ENVIRONMENT_ID}" \
  --pingone-export-environment-id "${PINGONE_EXPORT_ENVIRONMENT_ID}" \
  --pingone-worker-client-id "${PINGONE_WORKER_CLIENT_ID}" \
  --pingone-worker-client-secret "${PINGONE_WORKER_CLIENT_SECRET}" \
  --pingone-region-code "${PINGONE_REGION_CODE:-NA}" \
  --out ./shareable-module

echo
echo "=== Module Structure Generated ==="
tree ./shareable-module

echo
echo "=== Module Usage ==="
cat <<'EOF'
# In your Terraform configuration:
module "davinci" {
  source = "./shareable-module/davinci-module"

  pingone_environment_id = "your-env-id"
  
  # Variables are empty by default - provide values for your environment
  davinci_variable_company_name_value = "Your Company"
  davinci_connection_http_base_url    = "https://your-api.example.com"
}

# Access module outputs:
output "flow_ids" {
  value = module.davinci.flow_ids
}
EOF

echo
echo "=== Next Steps ==="
echo "1. Review generated module in ./shareable-module/davinci-module/"
echo "2. Customize variables in module.tf for your environment"
echo "3. Run 'terraform init && terraform plan' to preview changes"
echo "4. Optional: Publish module to Terraform Registry or version control"
