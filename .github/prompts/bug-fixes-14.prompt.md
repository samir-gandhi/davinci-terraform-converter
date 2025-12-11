There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug.

## Bug 14

Similar to the fixes that were made in bug 09 regarding connector property extraction, there are additional cases where masked secret attributes are not being extracted on variable resources. There is a sample API response at ()./api_response_sample_variable_company_context_secret_value.json) to use for creating data in tests. 

The logic should be like:

1. if the value is empty, regardless of `data_type`, leave a todo comment for the user to fill in later. There are instances where values are intentionally left blank and thus the value block should not be created. (this is already implemented)
2. if the variable's `data_type` is `secret` and it has a populated `value` attribute, replace the value with a variable reference and add that variable to the `ping-export-terraform.auto.tfvars` file with an empty string as the value for the user to fill in later. the `variable.value` in the API response can be expected to be `"******"`. (this is the missing part that needs to be implemented)
2. if the variable's `data_type` is not `secret`, and the `value` attribute is populated,  replace the value with a variable reference and add that variable to the `ping-export-terraform.auto.tfvars` file with the actual value. (this is already implemented)

currently the generated output of a variable with `data_type` of `secret` looks like this:

```hcl
resource "pingone_davinci_variable" "pingcli__samplesecretvar_company" {
  environment_id = var.pingone_environment_id

  name           = "samplesecretvar"
  context        = "company"
  data_type      = "secret"
  mutable        = true
  display_name   = "sample secret value"
  min            = 0
  max            = 2000

  # TODO: Add secret value manually
  # value = {
  #   secret_string = "your-secret-value"
  # }
}
```

This todo should be replaced with a variable reference like so:

```hcl
resource "pingone_davinci_variable" "pingcli__samplesecretvar_company" {
  environment_id = var.pingone_environment_id

  name           = "samplesecretvar"
  context        = "company"
  data_type      = "secret"
  mutable        = true
  display_name   = "sample secret value"
  min            = 0
  max            = 2000
  value = {
    secret_string = var.davinci_variable_samplesecretvar_company
  }
}

And the `ping-export-terraform.auto.tfvars` file should have an entry for this variable with a test value:

```hcl
davinci_variable_samplesecretvar_company = ""  # Secret value - provide manually
```

If the variable value in the API response is empty, the todo comment should remain as is:

```hcl
"pingcli__samplesecretvar_company" {
  environment_id = var.pingone_environment_id

  name           = "samplesecretvar"
  context        = "company"
  data_type      = "secret"
  mutable        = true
  display_name   = "sample secret value"
  min            = 0
  max            = 2000

  # TODO: Add secret value manually
  # value = {
  #   secret_string = "your-secret-value"
  # }
}
```
