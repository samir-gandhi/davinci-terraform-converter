There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 04

There is an issue with duplicate resource names when generating child module resource names. Specifically this duplication is noted on the variables resource. Variables in the backend service pingone davinci have names and types. It is possible to have multiple variables with the same name but different types. For example a variable named "pingcli__enableFeatureX" could exist as both a boolean and a string type variable. When generating the terraform resource names, the current logic only uses the variable name, leading to duplicate resource names like:

```hcl
resource "pingone_davinci_variable" "pingcli__origin" {
  environment_id = var.environment_id

  name           = "origin"
  context        = "company"
  data_type      = "string"

  value = {
    string = var.davinci_variable_origin_value
  }
  mutable        = true
  min            = 0
  max            = 2000
}


resource "pingone_davinci_variable" "pingcli__origin" {
  environment_id = var.environment_id

  name           = "origin"
  context        = "flowInstance"
  data_type      = "string"

  value = {
    string = var.davinci_variable_origin_2_value
  }
  mutable        = true
  min            = 0
  max            = 2000
}
```

In this example the variable value references a terraform value that does have a de-duplicated name by appending _2 to the second variable. However the resource name itself is duplicated which will lead to terraform errors. This needs to be fixed by updating the resource naming logic to include both the variable name and context when generating the resource name. like `pingcli__origin_company` and `pingcli__origin_flowInstance`.