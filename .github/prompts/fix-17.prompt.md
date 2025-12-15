There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug.

## Enchancement 17

Complete the following enhancements to the converter to address recent schema changes and improve usability. Look through the code of the project and try to align to it's standards where possible. To ensure quality, use test driven development (TDD) practices by writing unit tests that cover the new functionality and edge cases. Include acceptance tests against a live PingOne environment to validate end-to-end behavior. Create a phased plan on a new markdown file in the `.github/prompts/` directory named `fix-17-PROGRESS.md` that outlines the steps to implement these enhancements and provides space to document progress.

This enhancement will include three parts:

1. The Terraform provider schema had some updates, and so the converter needs to be updated to match the new schema. For `pingone_davinci_variable` resource, the API is called and the type for the value field is assigned based on the `data_type` field. This needs to be updated to no longer assign the type based on `data_type`, but instead identify the type based on what the API returns in the `value` field. The `data_type` field should still be stored in the state, but not used to determine the type of the `value` field.

Example variable API response:

```json
{
  "_links": {
    "self": {
      "href": "https://api.pingone.com/v1/environments/1b1e3c7d-8dd0-4280-b244-482dcb33716d/variables/0aab23cf-20e2-410b-8f7e-114d37461147"
    },
    "environment": {
      "href": "https://api.pingone.com/v1/environments/1b1e3c7d-8dd0-4280-b244-482dcb33716d"
    }
  },
  "id": "0aab23cf-20e2-410b-8f7e-114d37461147",
  "environment": {
    "id": "1b1e3c7d-8dd0-4280-b244-482dcb33716d"
  },
  "name": "ciam_facebookEnabled",
  "dataType": "boolean",
  "displayName": "Facebook enabled",
  "context": "company",
  "value": "false",
  "mutable": true,
  "min": 0,
  "max": 2000,
  "createdAt": "2025-12-09T02:42:47.552Z",
  "updatedAt": "2025-12-09T02:42:47.552Z"
}
```

new expected HCL after import:

```hcl
resource "pingone_davinci_variable" "pingcli__ciam_facebookEnabled_company" {
  environment_id = var.pingone_environment_id

  name           = "ciam_facebookEnabled"
  context        = "company"
  data_type      = "boolean"

  value = {
    string = "false"
  }
  mutable        = true
  min            = 0
  max            = 2000
}
```

Note that the `value` field is now a map with a `string` key, instead of a boolean key and value.

2. When the `export` command for this tool is run with the `--include-imports` flag, the exported HCL includes an `imports.tf` file that contains import blocks for all resources. Update the logic so that a set of commented our `terraform import` commands are included in the `imports.tf` file along  actual import blocks. This will make it easier for users to see the import commands they need to run without having to extract them from import blocks.

3. There have been a number of sorting features added to the converter to ensure deterministic ordering of resources in the generated HCL. The generated `ping-export-terraform.auto.tfvars` file, keeps the variables grouped by resource type, but the variables within each resource type are not sorted. Update the logic to sort the variables alphabetically within each resource type group.