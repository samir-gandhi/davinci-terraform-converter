There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 09

This is an extension to the fix that was made in bug-08. Right now we have handling to replace secret values with a placeholder text "TODO: Replace with actual %s". This value should a variable reference like the other properties, and this test value should be the value for the variable on the `ping-export-terraform.auto.tfvars` file.

Current:

```hcl
resource "pingone_davinci_connector_instance" "PingOne" {
  connector = {
    id = "pingOneSSOConnector"
  }
  environment_id = pingone_environment.master_flow_environment.id
  name           = "PingOne"
  properties     = jsonencode({
      "clientId": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_clientId}"
      },
      "clientSecret": {
          "type": "string",
          "value": "${TODO: Replace with actual client secret}"
      },
      "envId": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_envId}"
      },
      "region": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_region}"
      }
  })
}
```
Expected:

```hcl
resource "pingone_davinci_connector_instance" "PingOne" {
  connector = {
    id = "pingOneSSOConnector"
  }
  environment_id = pingone_environment.master_flow_environment.id
  name           = "PingOne"
  properties     = jsonencode({
      "clientId": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_clientId}"
      },
      "clientSecret": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_clientSecret}"
      },
      "envId": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_envId}"
      },
      "region": {
          "type": "string",
          "value": "${var.davinci_connection_PingOne_region}"
      }
  })
}