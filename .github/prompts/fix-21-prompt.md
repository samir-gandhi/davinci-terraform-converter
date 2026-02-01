# Fix 20

reference:

- (./data/dv-new/davinci_flow.md)
- (./data/dv-new/davinci_flow_deployed.md)
- (./data/dv-new/davinci_flow_enabled.md)
- (./data/dv-new/resource_davinci_flow_gen.go)
- (./data/)
- (./ARCHITECTURE.md)

Look at the reference documents listed above to understand this repository.
This repo works to convert API responses to Terraform HCL for `pingone_davinci_*` resources that are being developed.
There are two new resources `davinci_flow_deployed` and `davinci_flow_enabled` that need to be added to this tool.
Each of these resources depend on a davinci flow. There should be existing logic for getting the information needed for these resources because everything needed comes from the payload of a flow (whether API or provided file). The flow model may need to be updated.

The value for `enabled` on the `pingone_davinci_flow_enabled` should be set to true based on field `flowStatus` OR `enabled` at the root of a flow. The field `flowStatus` will exist (with value `disabled` or `enabled`) if the flow payload is from a file and the field `enabled` will exist if it is an API response. Thus if it is a `davinci-to-hcl` command, we can expect a `flowStatus` field and if is an export command we can expect an `enabled` field. However, logic should watch for both in case this is not consistent. Throw an error if both fields are found and contradict each other.

Example `pingone_davinci_flow_enabled` resource:

```hcl
resource "pingone_davinci_flow_enabled" "example" {
  environment_id = var.environment_id
  flow_id        = pingone_davinci_flow.example.id
  enabled        = pingone_davinci_flow.example.enabled
}
```

For the `pingone_davinci_flow_deploy` resource, the `deploy_trigger_values` map can have one key: `deployed_version`. The value for this can map to `publishedVersion` from the root of a flow. This would be a terraform dependency reference or hard code based on the `--skip-dependencies` flag.

Example `pingone_davinci_flow_deploy` resource:

```hcl
resource "pingone_davinci_flow_deploy" "example" {
  environment_id = var.environment_id
  flow_id        = pingone_davinci_flow.example.id
  deploy_trigger_values = {
    "deployed_version" = pingone_davinci_flow.example.published_version 
  }
}
```

Take a phased approach that tracks progress on `fix-21-PROGRESS.md` and complete the following tasks:

1. Identify if existign models need to be updated
2. Identify how these will be made as new resources. If they can be new files using existing functions.
3. Make phases for how to implement these new features. The phases should include plans for the fixes as well as plans for test updates.
4. Implement.
