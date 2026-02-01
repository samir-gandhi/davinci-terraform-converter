# Migrating DaVinci Resources to PingOne Provider

The following guide documents the process of migrating DaVinci Terraform resources from the legacy DaVinci provider to the new DaVinci resources within the PingOne Terraform provider.

> **NOTE:** The legacy DaVinci provider relied on human user credentials and browser-based SSO. The new PingOne Terraform provider utilizes PingOne worker applications, which removes the dependency on human credentials. This is significantly more acceptable for automation scenarios such as CI/CD pipelines or GitHub Actions workflows.

The goal of this migration process is to move configuration managed by the legacy provider to the PingOne provider with a focus on minizing impact to live infrastructure. This involves avoiding the deletion or recreation of resources and ensuring that `terraform apply` results in no functional changes during the process.

## Prerequisites

* Existing configuration managed by the Legacy DaVinci Provider.
* The `davinci-terraform-converter` tool installed.

## 1. Create a Configuration Export

First, you'll use the DaVinci exporter tool to create a configuration export of your current environment. This tool helps convert the legacy JSON-based flow definitions into the HCL schema required by the new provider.

Run the following command, ensuring you include the flags to capture values and imports:

```bash
davinci-terraform-converter export --include-values --include-imports --pingone-export-environment-id='<env-id>'

```

> **NOTE:** The `--include-values` flag ensures that attributes typically converted to variables have their live environment values included in a `.tfvars` file. The `--include-imports` flag generates the necessary import blocks to bring the infrastructure under Terraform management.

This command creates a child module and root module files. Ensure you add the generated `.tfvars` file to your `.gitignore` as it may contain sensitive values.

## 2. Update Provider Version

Next, you must update the provider version in your Terraform configuration to access the new beta resources.

Update the `pingidentity` provider version in your `terraform` block to point to the beta release (e.g., `v1.15.0-beta`). Since this is a beta release, you must use a direct tag.

```hcl
terraform {
  required_version = ">= 1.3"

  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = "v1.15.0-beta"
    }
  }
}

```

## 3. Bring New Resources Under Management

You will now import the new resources into your Terraform state. This process duplicates managed resources temporarily during the migration, which is acceptable as no `terraform apply` should be run yet.

Use the `ping-export-imports.tf` file generated in Step 1. This file contains a list of commented-out `terraform import` commands corresponding to your exported resources.

1. Uncomment the commands in `ping-export-imports.tf`.
2. Run the commands in your terminal to import the resources.

> **NOTE:** Running `terraform import` commands explicitly is recommended over using `import` blocks here. This allows for an isolated process where state can be manually edited if needed before HCL configuration changes are finalized.

## 4. Edit Secret Values

Certain values considered secrets in DaVinci cannot be read from the API. Instead, the API returns these as six asterisks (`******`). Consequently, your Terraform state will currently hold this masked value rather than the actual secret.

To ensure your Terraform state matches the configuration of a newly created item:

* **Update State:** Manually replace the `******` values in your state file with the actual secret values.
* **Update tfvars:** If you used the export command, the generated `*-terraform.auto.tfvars` file will have empty values for these secrets. Update these items manually.

```hcl
davinci_connection_PingOne_clientSecret = "actual_secret_value" # Secret value - provide manually

```

## 5. Update Dependencies and Verify

At this point, you should be able to run `terraform plan` and receive a "no changes needed" result.

Verify this outcome. Once verified, update any existing Terraform configuration that references the legacy DaVinci resources:

* **Same Schema:** If the attribute schema matches, simply update the provider and resource name (the first two items in the dot-notated reference).
* **Different Schema:** If the attribute differs, identify the new path to the equivalent attribute.

## 6. Remove Legacy Resources

Finally, remove the legacy resources from management. Since the new resources are now managing the infrastructure, the legacy items can be removed from the state.

Use the `terraform state rm` command to remove the old resources.

```bash
terraform state rm <legacy_resource_address>

```

After removing the legacy resources, run `terraform plan` one last time. It should show no changes. You can now remove references to the legacy DaVinci provider from your Terraform version block and run `terraform validate` to complete the migration.