# Fix 20

reference:

- (./ARCHITECTURE.md)

Look at the reference documents listed above to understand this repository.
This repo works to convert API responses to Terraform HCL for `pingone_davinci_*` resources that are being developed.

This fix is a simple change to update the naming of some generated files. Right now, when the tool generates a child module, there is a module name flag:

```text
--module-name string                     Terraform module name used in module.tf and import blocks (default "ping-export")
```

This edits the module name in the module.tf file. Your job is to:

1. Find the logic where the module flag is accepted and update it to also adjust the generated root file names.
2. Update the `module-name` flag description to: `Used to define Terraform module and prefix generated content (default "ping-export")`
3. Ensure there are no regressions, use a phased approach if needed.

Current example output now:

```text  
.
├── imports.tf
├── module.tf
├── ping-export-module
│   ├── outputs.tf
│   ├── pingone_davinci_application_flow_policy.tf
│   ├── pingone_davinci_application.tf
│   ├── pingone_davinci_connector_instance.tf
│   ├── pingone_davinci_flow.tf
│   ├── pingone_davinci_variable.tf
│   ├── variables.tf
│   └── versions.tf
├── ping-export-terraform.auto.tfvars
└── variables.tf
```

Desired output:

```text
.
├── ping-export-imports.tf
├── ping-export-module.tf
├── ping-export-module
│   ├── outputs.tf
│   ├── pingone_davinci_application_flow_policy.tf
│   ├── pingone_davinci_application.tf
│   ├── pingone_davinci_connector_instance.tf
│   ├── pingone_davinci_flow.tf
│   ├── pingone_davinci_variable.tf
│   ├── variables.tf
│   └── versions.tf
├── ping-export-terraform.auto.tfvars
└── ping-export-variables.tf
```