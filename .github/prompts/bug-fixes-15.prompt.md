There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug.

## Bug 15

I am now at a place with a created, live, PingOne environment. I have then exported that environment with this tool, and imported all of the resources to create a terraform state file. The goal is that this process leads to the ability to run `terraform plan` against the export output and the plan return "No changes. Your infrastructure matches the configuration." This would be a final proof that the export process of this tool is a true, complete representation of a live environment. So far this almost there. The last item left is flows. 

So, your job is to:

1. Identify and document which items are missing from the export, or are being mapped incorrectly.
2. Create a plan with a phased approach for how to update the structs in the api folder, converter and, if needed, exporter tools.
3. Implement the phases with test-driven development. 

To help with completeness I am providing all the resource examples and documentation that should be needed. 

Naturally, there are items in the terraform plan output that are not shown because the attribute is marked as sensitive. You should still be able to use the provided terraform configuration and state file to identify what is missing or mapped incorrectly.

Examples and Documentation:

- (./prompt-resources/davinci_flow.md) - Terraform documentation for the flow resource in order to see the expected HCL output schema.
- (./prompt-resources/agreement-subflow-api-response.json) - the JSON from an actual API response for this target davinci flow. retrieved from `environments/1b1e3c7d-8dd0-4280-b244-482dcb33716d/flows/0be33533d3896031bb0a4d569884d027`
- (./prompt-resources/agreement-subflow-tf-plan.txt) - A snippet of `terraform plan` output that shows what changes terraform is currently identifying.
- (./prompt-resources/agreement-subflow-export.json) - The historical version of the exported flow, this export is a download that happens through the davinci UI.
- (./prompt-resources/agreement-subflow-state.tfstate) - a snippet of the terraform state file after importing the resource into terraform.
- (./prompt-resources/agreement-subflow-hcl.tf) - The HCL that was generated from the export process of this tool for the target flow.
