There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 02

When the export command is run, we have definitions within the project for which resource attributes will become variables. These terraform variable placeholders are being generated correctly within the generated HCL. Additionally the terraform variable inputs are also being created withtin the module.tf file. There are two improvements that need to be made:

1. When there is no `--include-values` flag provided to the export command, the terraform variable inputs created on the terraform module should also have terraform variables as the values on module.tf. Then a template `ping-export-terraform.tfvars` file should be created so a consumer can easily fill in variable values safely. 

2. When the `--include-values` is provided flag should extract the terraform variable values from the collected API response and populate the `ping-export-terraform.tfvars` variable values with actual responses. Note, this still will not be possible for values that return asterisks from the API response. Those values must be known by the user. 