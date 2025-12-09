There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 10

The are default values created in the variables.tf file in the child module that gets creating in this project. These default values may have actual secret values in them. Update the generator to not include default values. The actual value should be set in the `ping-export-terraform.auto.tfvars` file - This is working as expected.