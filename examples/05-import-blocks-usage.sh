#!/usr/bin/env bash
#
# Example: Using Terraform Import Blocks
#
# This demonstrates the import blocks feature that allows automatic
# import of existing DaVinci resources into Terraform state.
#
# Requires: Terraform 1.5.0 or higher
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
        exit 1
    fi
done

echo "=== Terraform Import Blocks Feature Demo ==="
echo ""
echo "This feature generates Terraform import blocks alongside resources"
echo "(enabled by default), enabling automatic import of existing infrastructure."
echo ""

# Step 1: Export with import blocks (default behavior)
echo "Step 1: Exporting with import blocks (default)..."
OUTPUT_FILE="/tmp/davinci-with-imports.tf"

"${BINARY}" export \
    --out "${OUTPUT_FILE}"

echo "✓ Export complete: ${OUTPUT_FILE}"
echo ""

# Show statistics
RESOURCE_COUNT=$(grep -c '^resource "' "${OUTPUT_FILE}" || echo "0")
IMPORT_COUNT=$(grep -c '^import {' "${OUTPUT_FILE}" || echo "0")

echo "Statistics:"
echo "  Resources: ${RESOURCE_COUNT}"
echo "  Import blocks: ${IMPORT_COUNT}"
echo ""

# Show example import block
echo "Example import block + resource:"
echo "---"
head -30 "${OUTPUT_FILE}" | tail -20
echo "---"
echo ""

# Step 2: Set up Terraform workspace
echo "Step 2: Setting up Terraform workspace..."
WORKSPACE="/tmp/terraform-import-test"
rm -rf "${WORKSPACE}"
mkdir -p "${WORKSPACE}"
cd "${WORKSPACE}"

# Copy exported HCL
cp "${OUTPUT_FILE}" ./resources.tf

# Create provider configuration
cat > provider.tf << 'EOF'
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = ">= 0.28.0"
    }
  }
}

provider "pingone" {
  client_id      = var.pingone_client_id
  client_secret  = var.pingone_client_secret
  environment_id = var.pingone_environment_id
  region_code    = var.pingone_region_code
}
EOF

# Create variables
cat > variables.tf << 'EOF'
variable "pingone_client_id" {
  description = "PingOne OAuth2 client ID"
  type        = string
  sensitive   = true
}

variable "pingone_client_secret" {
  description = "PingOne OAuth2 client secret"
  type        = string
  sensitive   = true
}

variable "pingone_environment_id" {
  description = "PingOne environment ID"
  type        = string
}

variable "pingone_region_code" {
  description = "PingOne region code"
  type        = string
  default     = "NA"
}

variable "environment_id" {
  description = "DaVinci environment ID"
  type        = string
}
EOF

# Create tfvars (you would fill this in with real values)
cat > terraform.tfvars.example << EOF
# Copy this file to terraform.tfvars and fill in your values
# DO NOT commit terraform.tfvars to version control

pingone_client_id      = "your-client-id"
pingone_client_secret  = "your-client-secret"
pingone_environment_id = "${PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID}"
pingone_region_code    = "${PINGCLI_PINGONE_REGION_CODE:-NA}"
environment_id         = "${PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID:-${PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID}}"
EOF

echo "✓ Terraform workspace created: ${WORKSPACE}"
echo ""
echo "Files created:"
ls -lh
echo ""

# Step 3: Show what would happen with terraform plan
echo "Step 3: What happens next..."
echo ""
echo "To complete the import process:"
echo ""
echo "  cd ${WORKSPACE}"
echo "  cp terraform.tfvars.example terraform.tfvars"
echo "  # Edit terraform.tfvars with your credentials"
echo ""
echo "  terraform init"
echo "  terraform plan"
echo ""
echo "Expected terraform plan output:"
echo "  Plan: ${RESOURCE_COUNT} to import, 0 to add, 0 to change, 0 to destroy"
echo ""
echo "  terraform apply"
echo ""
echo "This will import all ${RESOURCE_COUNT} resources into Terraform state!"
echo ""

# Show the benefit
echo "=== Benefits of Import Blocks ==="
echo ""
echo "Without import blocks (using --skip-imports):"
echo "  • You'd need to run ${RESOURCE_COUNT} manual 'terraform import' commands"
echo "  • Each command requires correct resource address and ID"
echo "  • Easy to make mistakes with IDs"
echo "  • Time-consuming for large environments"
echo ""
echo "With import blocks (default behavior):"
echo "  • Single 'terraform apply' imports everything"
echo "  • No manual ID mapping required"
echo "  • Correct import IDs guaranteed"
echo "  • Fast and automated"
echo ""

echo "=== Advanced Usage ==="
echo ""
echo "Combine with skip-dependencies for standalone resources:"
echo "  ${BINARY} export --skip-dependencies --out standalone.tf"
echo ""
echo "This creates fully standalone HCL with hardcoded IDs and import blocks"
echo "that can be imported and applied without any variable dependencies."
echo ""
echo "Disable import blocks for Terraform < 1.5:"
echo "  ${BINARY} export --skip-imports --out legacy.tf"
echo ""
