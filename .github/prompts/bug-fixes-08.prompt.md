There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 08

Create a phased approach to fix this bug with test driven development.

Right now the generator creates `pingone_davinci_connector_instance.properties` with a `jsonencode` payload. this payload type is correct, but the fields within the payload are formatted incorrectly. The properties field should be pulled from the API response and formatted as terraform acceptable json strings.

The properties field in the API response can vary widely depending on the connector type. The API server also has very little validation on the payload structure sent on creation/update of the connector instance. 

First, understand why the current implementation is formatting the properties this way. I believe there is some intentional mapping configuring it this way, it seems to match the format of the legacy terraform provider. Check with me before making any changes.

Second, create tests that will validate the properties field is formatted correctly. Use the expected formats below as a guide. It should be directly extracting the properties section of the API response. 

Sample API response for PingOne SSO connector instance:

```json
{
  "connectorId": "pingOneSSOConnector",
  "createdDate": 1763659579562,
  "customerId": "8b0c151269a21d19b8b2b90918dedce1",
  "name": "PingOne",
  "properties": {
      "clientId": {
          "type": "string",
          "value": "3642f58b-b0c2-4a35-b1b1-e24d051de546"
      },
      "clientSecret": {
          "type": "string",
          "value": "******"
      },
      "envId": {
          "type": "string",
          "value": "4111cd46-25bf-4a5b-8c74-184a9d0c1826"
      },
      "region": {
          "type": "string",
          "value": "NA"
      }
  },
  "updatedDate": 1763678194685,
  "connectionId": "94141bf2f1b9b59a5f5365ff135e02bb",
  "companyId": "4111cd46-25bf-4a5b-8c74-184a9d0c1826"
}
```

Current output from generator:

```hcl
resource "pingone_davinci_connector_instance" "pingcli__PingOne" {
  environment_id = var.pingone_environment_id

  name           = "PingOne"

  connector = {
    id = "pingOneSSOConnector"
  }

  properties = jsonencode({
    "clientId"     : var.davinci_connection_PingOne_clientId,
    "clientSecret" : "TODO: Replace with actual client secret",
    "envId"        : "4111cd46-25bf-4a5b-8c74-184a9d0c1826",
    "region"       : var.davinci_connection_PingOne_region
  })
}
```

Corrected output from generator:

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

Third, compare our current mapping system of property attributes so that variables are injected correctly. I believe the variable injection is currently rather brittle as it relies on a hardcoded mapping of known property keys to variable names. The `properties.<name>.value` field to variable mapping should be two parts. Part one is an attempt at dynamic based on structure. For example if the property structure is:

```json
"properties": {
    "clientId": {
        "type": "string",
        "value": "3642f58b-b0c2-4a35-b1b1-e24d051de546"
    },
    ...
}
we can dynamically create a variable name like `davinci_connection_<connectorName>_<propertyName>` so in this case it would be `davinci_connection_PingOne_clientId`.
If the dynamic mapping fails (for example if there are duplicate property names across different connectors or if the property structure is difference), we can fall back to a hardcoded mapping. This hardcoded mapping should be in a separate configuration file so it can be easily updated in the future and should be well documented.

Sample of unstructured properties that may require hardcoded mapping:

```json
{
    "connectorId": "genericConnector",
    "createdDate": 1763750061732,
    "customerId": "bb5fef4347dc9293873ae12ee9e73f1c",
    "name": "OIDC & OAuth IdP",
    "properties": {
        "customAuth": {
            "properties": {
                "providerName": {
                    "displayName": "Provider Name",
                    "preferredControlType": "textField",
                    "required": true,
                    "value": "asdf"
                },
                "authTypeDropdown": {
                    "displayName": "Auth Type",
                    "preferredControlType": "dropDown",
                    "required": true,
                    "options": [
                        {
                            "name": "Oauth2",
                            "value": "oauth2"
                        },
                        {
                            "name": "OpenId",
                            "value": "openId"
                        }
                    ],
                    "enum": [
                        "oauth2",
                        "openId"
                    ],
                    "value": "oauth2"
                },
                "skRedirectUri": {
                    "displayName": "Redirect URL",
                    "preferredControlType": "textField",
                    "disabled": true,
                    "initializeValue": "SINGULARKEY_REDIRECT_URI",
                    "copyToClip": true
                },
                "issuerUrl": {
                    "preferredControlType": "textField",
                    "displayName": "Issuer URL",
                    "info": "Required if auth type is OpenID"
                },
                "authorizationEndpoint": {
                    "displayName": "Authorization Endpoint",
                    "preferredControlType": "textField",
                    "required": true,
                    "value": "asdf"
                },
                "tokenEndpoint": {
                    "displayName": "Token Endpoint",
                    "preferredControlType": "textField",
                    "required": true,
                    "value": "asdf"
                },
                "bearerToken": {
                    "preferredControlType": "textField",
                    "type": "boolean",
                    "displayName": "Token Attachment",
                    "info": "Optional field. Prepend token with this value. Example: Bearer or Token"
                },
                "userInfoEndpoint": {
                    "displayName": "User Info Endpoint",
                    "preferredControlType": "textFieldArrayView",
                    "required": true,
                    "value": [
                        "asdf"
                    ]
                },
                "clientId": {
                    "displayName": "App ID",
                    "preferredControlType": "textField",
                    "required": true,
                    "value": "asdf"
                },
                "clientSecret": {
                    "displayName": "Client Secret",
                    "preferredControlType": "textField",
                    "secure": true,
                    "required": true,
                    "value": "******"
                },
                "scope": {
                    "displayName": "Scope",
                    "preferredControlType": "textField",
                    "required": true,
                    "value": "asdf"
                },
                "state": {
                    "displayName": "Use State",
                    "preferredControlType": "toggleSwitch",
                    "value": false
                },
                "userConnectorAttributeMapping": {
                    "type": "object",
                    "displayName": null,
                    "preferredControlType": "userConnectorAttributeMapping",
                    "newMappingAllowed": true,
                    "title1": null,
                    "title2": null,
                    "sections": [
                        "attributeMapping"
                    ]
                },
                "customAttributes": {
                    "type": "array",
                    "displayName": "Connector Attributes",
                    "preferredControlType": "tableViewAttributes",
                    "info": "These attributes will be available in User Connector Attribute Mapping.",
                    "sections": [
                        "connectorAttributes"
                    ],
                    "value": [
                        {
                            "name": "id",
                            "description": "ID",
                            "type": "string",
                            "value": null,
                            "minLength": "1",
                            "maxLength": "300",
                            "required": true,
                            "attributeType": "sk"
                        },
                        {
                            "name": "name",
                            "description": "Display Name",
                            "type": "string",
                            "value": null,
                            "minLength": "1",
                            "maxLength": "250",
                            "required": false,
                            "attributeType": "sk"
                        },
                        {
                            "name": "email",
                            "description": "Email",
                            "type": "string",
                            "value": null,
                            "minLength": "1",
                            "maxLength": "250",
                            "required": false,
                            "attributeType": "sk"
                        }
                    ]
                },
                "returnToUrl": {
                    "displayName": "Application Return To URL",
                    "preferredControlType": "textField",
                    "info": "When using the embedded flow player widget and an IdP/Social Login connector, provide a callback URL to return back to the application."
                }
            }
        }
    },
    "updatedDate": 1763750085080,
    "connectionId": "3b51289bf0126ac190d61284920d99e4",
    "companyId": "4111cd46-25bf-4a5b-8c74-184a9d0c1826"
}
```