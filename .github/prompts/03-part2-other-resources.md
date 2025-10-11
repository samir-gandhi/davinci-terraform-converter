---
mode: agent
---

# Part 2 - Phases 2.2-2.6: Other DaVinci Resources

**Status**: ⏳ NOT STARTED (After Phase 2.1)

## Overview

These phases extend conversion beyond flows to handle all DaVinci resource types:
- Applications (Phase 2.2)
- Flow Policies (Phase 2.3)
- Connections/Connector Instances (Phase 2.4)
- Variables (Phase 2.5)
- Multi-Resource Integration (Phase 2.6)

All phases follow TDD approach with tests written first.

## Phase 2.2: Application Resource Conversion

**Resource**: `pingone_davinci_application`

### Complete Schema Reference

**Reference Materials**:
- Terraform Provider Docs: `docs/resources/davinci_application.md`
- pingone-go-client structs for DaVinci application structure

**Key Structures to Handle**:

1. **Top-Level Application Attributes**:
   - `environment_id` (string, required) - The PingOne environment ID (use `var.environment_id` in generated HCL)
   - `name` (string, required) - Application name
   - `api_key` (object, optional) - API key configuration (see section 2 below)
   - `oauth` (object, optional) - OAuth configuration (see section 3 below)
   - `id` (string, read-only) - Resource ID

2. **api_key Object**:
   - `enabled` (boolean, optional) - Whether API key authentication is enabled
   - `value` (string, read-only, sensitive) - The actual API key value (managed by provider, not in input)

3. **oauth Object**:
   - `enforce_signed_request_openid` (boolean, optional) - Whether to enforce signed requests for OpenID
   - `grant_types` (set of strings, optional) - OAuth grant types. Options: `authorizationCode`, `clientCredentials`, `implicit`
   - `logout_uris` (set of strings, optional) - Logout redirect URIs
   - `redirect_uris` (set of strings, optional) - OAuth redirect URIs
   - `scopes` (set of strings, optional) - OAuth scopes. Options: `flow_analytics`, `offline_access`, `openid`, `profile`
   - `sp_jwks_openid` (string, optional) - Service provider JWKS for OpenID
   - `sp_jwks_url` (string, optional) - Service provider JWKS URL
   - `client_secret` (string, read-only, sensitive) - OAuth client secret (managed by provider, not in input)

**HCL Syntax Example**:
```hcl
resource "pingone_davinci_application" "my_application" {
  environment_id = var.environment_id
  name           = "My Application"
  
  api_key = {
    enabled = true
  }
  
  oauth = {
    grant_types                   = ["authorizationCode", "implicit"]
    scopes                        = ["openid", "profile"]
    enforce_signed_request_openid = false
    redirect_uris                 = ["https://example.com/callback"]
    logout_uris                   = ["https://example.com/logout"]
    sp_jwks_openid                = "{ ... }"  # Optional
    sp_jwks_url                   = "https://example.com/.well-known/jwks.json"  # Optional
  }
}
```

### Test Case

Function: `TestApplicationConversion`

**Sample JSON Structure**:
```json
{
  "id": "app-123",
  "name": "My Application",
  "environment": {"id": "env-123"},
  "apiKey": {
    "enabled": true,
    "value": "ak_xxx"
  },
  "oauth": {
    "clientSecret": "cs_xxx",
    "grantTypes": ["authorizationCode", "implicit"],
    "redirectUris": ["https://example.com/callback"],
    "logoutUris": ["https://example.com/logout"],
    "scopes": ["openid", "profile"],
    "enforceSignedRequestOpenid": false
  },
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-02T00:00:00Z"
}
```

**Expected HCL**:
```hcl
resource "pingone_davinci_application" "my_application" {
  environment_id = var.environment_id
  name           = "My Application"
  
  api_key_enabled = true
  
  oauth {
    enabled = true
    values = {
      enabled                    = true
      grant_types                = ["authorizationCode", "implicit"]
      redirect_uris              = ["https://example.com/callback"]
      logout_uris                = ["https://example.com/logout"]
      scopes                     = ["openid", "profile"]
      enforce_signed_request_openid = false
    }
  }
}
```

### Implementation

Function: `generateApplicationHCL()`

**Key Points**:
- Handle OAuth configuration block with enabled flag and values
- Handle API key configuration (enabled flag only, mask actual key)
- Mask sensitive values (client secrets, API keys) with TODO comments
- Handle optional fields gracefully

**Test Scenarios**:
- Application with OAuth and API key
- Application with only OAuth
- Application with only API key
- Optional fields missing

## Phase 2.3: Flow Policy Resource Conversion

**Resource**: `pingone_davinci_application_flow_policy`

### Complete Schema Reference

**Reference Materials**:
- Terraform Provider Docs: `docs/resources/davinci_application_flow_policy.md`
- pingone-go-client structs for DaVinci flow policy structure

**Key Structures to Handle**:

1. **Top-Level Flow Policy Attributes**:
   - `environment_id` (string, required) - The PingOne environment ID (use `var.environment_id` in generated HCL)
   - `da_vinci_application_id` (string, required) - Reference to the DaVinci application (use Terraform reference)
   - `flow_distributions` (set of objects, required) - Flow distributions (see section 2 below)
   - `name` (string, optional) - Flow policy name
   - `status` (string, optional) - Policy status. Options: `disabled`, `enabled`
   - `trigger` (object, optional) - Flow trigger configuration (see section 3 below)
   - `id` (string, read-only) - Resource ID

2. **flow_distributions Objects (Set)**:
   - `id` (string, required) - Flow ID (use Terraform reference to pingone_davinci_flow)
   - `version` (number, required) - Flow version (-1 for latest)
   - `weight` (number, optional) - Distribution weight for flow
   - `ip` (set of strings, optional) - IP ranges for conditional routing
   - `success_nodes` (set of objects, optional) - Success node configuration:
     - `id` (string, required) - Node ID

3. **trigger Object**:
   - `type` (string, optional) - Trigger type (e.g., "AUTHENTICATION")
   - `configuration` (object, optional):
     - `mfa` (object, optional):
       - `enabled` (boolean, optional) - Whether MFA timeout is enabled
       - `time` (number, optional) - MFA timeout value
       - `time_format` (string, optional) - Time format (e.g., "seconds", "minutes")
     - `pwd` (object, optional):
       - `enabled` (boolean, optional) - Whether password timeout is enabled
       - `time` (number, optional) - Password timeout value
       - `time_format` (string, optional) - Time format (e.g., "seconds", "minutes")

**HCL Syntax Example**:
```hcl
resource "pingone_davinci_application_flow_policy" "main_policy" {
  environment_id         = var.environment_id
  da_vinci_application_id = pingone_davinci_application.my_app.id
  
  name   = "Main Policy"
  status = "enabled"
  
  trigger = {
    type = "AUTHENTICATION"
    
    configuration = {
      mfa = {
        enabled     = true
        time        = 300
        time_format = "seconds"
      }
      
      pwd = {
        enabled     = true
        time        = 3600
        time_format = "seconds"
      }
    }
  }
  
  flow_distributions = [
    {
      id      = pingone_davinci_flow.registration.id
      version = -1
      weight  = 100
      
      success_nodes = [
        {
          id = "successNode1"
        },
        {
          id = "successNode2"
        }
      ]
      
      ip = ["10.0.0.0/8", "192.168.0.0/16"]
    },
    {
      id      = pingone_davinci_flow.authentication.id
      version = -1
      weight  = 50
    }
  ]
}
```

### Test Case (Flow Policy)

Function: `TestFlowPolicyConversion`

**Sample JSON Structure**:
```json
{
  "id": "policy-123",
  "name": "Main Policy",
  "status": "enabled",
  "application": {"id": "app-123"},
  "trigger": {
    "type": "AUTHENTICATION",
    "configuration": {
      "mfa": {
        "enabled": true,
        "time": 300,
        "timeFormat": "seconds"
      },
      "pwd": {
        "enabled": true,
        "time": 3600,
        "timeFormat": "seconds"
      }
    }
  },
  "flowDistributions": [
    {
      "id": "flow-123",
      "version": 1,
      "weight": 100,
      "successNodes": [
        {"id": "successNode1"}
      ],
      "ip": ["10.0.0.0/8"]
    }
  ]
}
```

**Expected HCL**:
```hcl
resource "pingone_davinci_application_flow_policy" "main_policy" {
  environment_id = var.environment_id
  application_id = pingone_davinci_application.my_application.id
  name           = "Main Policy"
  status         = "enabled"
  
  trigger {
    type = "AUTHENTICATION"
    
    mfa {
      enabled     = true
      time_value  = 300
      time_unit   = "seconds"
    }
    
    password {
      enabled    = true
      time_value = 3600
      time_unit  = "seconds"
    }
  }
  
  policy_flow {
    flow_id    = pingone_davinci_flow.my_flow.id
    version_id = 1
    weight     = 100
    
    success_node {
      node_id = "successNode1"
    }
    
    ip_range = ["10.0.0.0/8"]
  }
}
```

### Implementation

Function: `generateFlowPolicyHCL()`

**Key Points**:
- Handle trigger block with type and configuration
- Handle MFA and password settings under trigger
- Handle flow_policy block with status and optional priority
- Handle repeatable policy_flow blocks
- Handle success_nodes (repeatable nested blocks)
- Handle IP ranges for conditional routing

**Test Scenarios**:
- Single flow distribution
- Multiple flow distributions
- Various trigger types
- Optional fields (priority, IP ranges)

## Phase 2.4: Connection Resource Conversion

**Resource**: `pingone_davinci_connector_instance`

**Note**: In DaVinci terminology, "connections" are connector instances - specific configurations of connectors.

### Complete Schema Reference

**Reference Materials**:
- Terraform Provider Docs: `docs/resources/davinci_connector_instance.md`
- pingone-go-client structs for DaVinci connector instance structure

**Key Structures to Handle**:

1. **Top-Level Connector Instance Attributes**:
   - `environment_id` (string, required) - The PingOne environment ID (use `var.environment_id` in generated HCL)
   - `name` (string, required) - Name for this connector instance (this is the user-assigned name)
   - `connector` (object, required) - Connector type reference (see section 2 below)
   - `properties` (string, optional, sensitive) - JSON string of connector-specific properties
   - `id` (string, read-only) - Resource ID

2. **connector Object**:
   - `id` (string, required) - Connector type ID (e.g., "annotationConnector", "httpConnector", "crowdStrikeConnector")

**Special Handling**:
- **Connector Type ID**: The `connector.id` is the connector type (e.g., "httpConnector"), NOT a reference to another resource
- **Properties**: JSON string containing connector-specific configuration (varies by connector type)
  - Should be sensitive data for connectors with credentials
  - Properties structure is connector-specific (no universal schema)
  - Use `jsonencode()` in generated HCL for proper escaping
  - For export: mask sensitive values (API keys, secrets) with TODO comments
- **Naming**: Resource name should use pingcli format (hex-encoded)

**Common Connector Types**:
- `annotationConnector` - No properties needed
- `httpConnector` - Usually requires endpoint URL
- `flowConnector` - Requires flow reference
- `functionsConnector` - JavaScript functions
- OAuth connectors (Google, Facebook, etc.) - Require client ID/secret
- API connectors - Require API keys/credentials

**HCL Syntax Examples**:

```hcl
# Simple connector with no properties
resource "pingone_davinci_connector_instance" "pingcli__My-0020-Annotation" {
  environment_id = var.environment_id
  name           = "My Annotation"
  
  connector = {
    id = "annotationConnector"
  }
}

# HTTP connector with endpoint configuration
resource "pingone_davinci_connector_instance" "pingcli__External-0020-API" {
  environment_id = var.environment_id
  name           = "External API"
  
  connector = {
    id = "httpConnector"
  }
  
  properties = jsonencode({
    "url" : "https://api.example.com/endpoint"
  })
}

# OAuth connector with credentials (masked for export)
resource "pingone_davinci_connector_instance" "pingcli__Google-0020-Login" {
  environment_id = var.environment_id
  name           = "Google Login"
  
  connector = {
    id = "googleIdpConnector"
  }
  
  properties = jsonencode({
    "clientId" : "TODO: Replace with actual client ID",
    "clientSecret" : "TODO: Replace with actual client secret",
    "scope" : "openid profile email"
  })
}

# Flow connector (references another flow)
resource "pingone_davinci_connector_instance" "pingcli__Subflow-0020-Call" {
  environment_id = var.environment_id
  name           = "Subflow Call"
  
  connector = {
    id = "flowConnector"
  }
  
  properties = jsonencode({
    "flowId" : "TODO: Replace with flow reference or ID"
  })
}
```

### Test Case (Connector Instances)

Function: `TestConnectionConversion`

**Sample JSON Structure**:
```json
{
  "id": "conn-123",
  "name": "PingOne SSO Connection",
  "environment": {"id": "env-123"},
  "connector": {"id": "pingOneSSOConnector"},
  "properties": {
    "clientId": "client-123",
    "clientSecret": "secret-xxx",
    "region": "NA",
    "apiPath": "/api/v1"
  },
  "createdAt": "2024-01-01T00:00:00Z"
}
```

**Expected HCL**:
```hcl
resource "pingone_davinci_connection" "pingone_sso_connection" {
  environment_id = var.environment_id
  connector_id   = "pingOneSSOConnector"
  name           = "PingOne SSO Connection"
  
  property {
    name  = "clientId"
    value = "client-123"
  }
  
  property {
    name  = "clientSecret"
    value = "TODO: Sensitive value - replace with actual value"
  }
  
  property {
    name  = "region"
    value = "NA"
  }
  
  property {
    name  = "apiPath"
    value = "/api/v1"
  }
}
```

### Implementation

Function: `generateConnectionHCL()`

**Key Points**:
- Handle connector_id reference (will be resolved in Part 4)
- Convert properties object to property blocks
- Identify and mask sensitive fields with TODO comments
- Common sensitive patterns: password, secret, token, apiKey, clientSecret, privateKey

**Sensitive Field Detection**:
```go
func isSensitiveField(key string) bool {
    lowerKey := strings.ToLower(key)
    sensitivePatterns := []string{
        "password", "secret", "token", "apikey", 
        "clientsecret", "privatekey", "credential",
    }
    for _, pattern := range sensitivePatterns {
        if strings.Contains(lowerKey, pattern) {
            return true
        }
    }
    return false
}
```

**Test Scenarios**:
- Various connector types
- Sensitive field masking
- Empty or minimal properties
- Complex nested properties

## Phase 2.5: Variable Resource Conversion

**Resource**: `pingone_davinci_variable`

### Complete Schema Reference

**Reference Materials**:
- Terraform Provider Docs: `docs/resources/davinci_variable.md`
- pingone-go-client structs for DaVinci variable structure

**Key Structures to Handle**:

1. **Top-Level Variable Attributes**:
   - `environment_id` (string, required) - The PingOne environment ID (use `var.environment_id` in generated HCL)
   - `name` (string, required) - Variable name
   - `context` (string, required) - Variable context. Options: `company`, `flow`, `flowInstance`, `user`
   - `data_type` (string, required) - Variable data type. Options: `boolean`, `number`, `object`, `secret`, `string`
   - `mutable` (boolean, required) - Whether the variable value is mutable
   - `display_name` (string, optional) - Display name for the variable
   - `flow` (object, optional) - Flow reference (required for `flow` context) (see section 2 below)
   - `min` (number, optional) - Minimum value (for `number` data type)
   - `max` (number, optional) - Maximum value (for `number` data type)
   - `value` (object, optional) - Variable value (see section 3 below)
   - `id` (string, read-only) - Resource ID

2. **flow Object** (Required for `flow` context):
   - `id` (string, required) - Flow ID (use Terraform reference to pingone_davinci_flow)

3. **value Object** (Type-specific value):
   - For `string` data_type:
     - `string` (string, optional) - String value
   - For `number` data_type:
     - `number` (number, optional) - Numeric value
   - For `boolean` data_type:
     - `boolean` (boolean, optional) - Boolean value
   - For `object` data_type:
     - `object` (string, optional) - JSON string representation of object
   - For `secret` data_type:
     - `secret` (string, optional, sensitive) - Secret value (should be masked in export)

**Context-Specific Behavior**:
- `company`: Environment-wide variable, no flow reference needed
- `flow`: Flow-specific variable, requires `flow.id` reference
- `flowInstance`: Runtime variable per flow execution, typically no static value
- `user`: User-specific variable, no flow reference needed

**HCL Syntax Examples**:

```hcl
# Company-level string variable
resource "pingone_davinci_variable" "api_endpoint" {
  environment_id = var.environment_id
  name           = "apiEndpoint"
  context        = "company"
  data_type      = "string"
  mutable        = true
  display_name   = "API Endpoint URL"
  
  value = {
    string = "https://api.example.com"
  }
}

# Flow-level number variable with min/max
resource "pingone_davinci_variable" "max_retries" {
  environment_id = var.environment_id
  name           = "maxRetries"
  context        = "flow"
  data_type      = "number"
  mutable        = false
  min            = 0
  max            = 10
  
  flow = {
    id = pingone_davinci_flow.my_flow.id
  }
  
  value = {
    number = 3
  }
}

# Secret variable (value should be masked)
resource "pingone_davinci_variable" "api_key" {
  environment_id = var.environment_id
  name           = "apiKey"
  context        = "company"
  data_type      = "secret"
  mutable        = false
  
  value = {
    secret = "TODO: Sensitive value - replace with actual secret"
  }
}

# Boolean variable
resource "pingone_davinci_variable" "feature_enabled" {
  environment_id = var.environment_id
  name           = "featureEnabled"
  context        = "company"
  data_type      = "boolean"
  mutable        = true
  
  value = {
    boolean = true
  }
}

# Object variable
resource "pingone_davinci_variable" "config" {
  environment_id = var.environment_id
  name           = "appConfig"
  context        = "company"
  data_type      = "object"
  mutable        = true
  
  value = {
    object = jsonencode({
      "timeout" : 30,
      "retries" : 3
    })
  }
}
```

### Test Case (Variables)

Function: `TestVariableConversion`

**Sample JSON Structures**:

```json
// Company-level string variable
{
  "id": "var-123",
  "name": "apiEndpoint",
  "environment": {"id": "env-123"},
  "context": "company",
  "type": "string",
  "value": "https://api.example.com",
  "mutable": true,
  "description": "API endpoint URL"
}

// Flow-level number variable
{
  "id": "var-456",
  "name": "maxRetries",
  "context": "flow",
  "type": "number",
  "value": 3,
  "mutable": false,
  "min": 0,
  "max": 10,
  "flow": {"id": "flow-123"}
}

// Secret variable
{
  "id": "var-789",
  "name": "apiKey",
  "context": "company",
  "type": "secret",
  "value": "sk_xxx",
  "mutable": false
}
```

**Expected HCL**:
```hcl
resource "pingone_davinci_variable" "api_endpoint" {
  environment_id = var.environment_id
  name           = "apiEndpoint"
  context        = "company"
  type           = "string"
  value          = "https://api.example.com"
  mutable        = true
  description    = "API endpoint URL"
}

resource "pingone_davinci_variable" "max_retries" {
  environment_id = var.environment_id
  name           = "maxRetries"
  context        = "flow"
  type           = "number"
  value          = 3
  mutable        = false
  min            = 0
  max            = 10
  flow_id        = pingone_davinci_flow.my_flow.id
}

resource "pingone_davinci_variable" "api_key" {
  environment_id = var.environment_id
  name           = "apiKey"
  context        = "company"
  type           = "secret"
  value          = "TODO: Sensitive value - replace with actual secret"
  mutable        = false
}
```

### Implementation

Function: `generateVariableHCL()`

**Key Points**:
- Handle context types: company, flow, flowInstance, user
- Handle data types: string, number, boolean, object, secret
- Handle optional fields: description, min/max (for numbers), flow_id (for flow context)
- Mask secret values with TODO comments
- Convert object values to HCL map syntax

**Context-Specific Logic**:
- `flow` context: Requires flow_id reference (resolve in Part 4)
- `flowInstance`: Runtime variable, may not need conversion
- `company`: Environment-wide, no additional references
- `user`: User-specific, no additional references

**Test Scenarios**:
- Each context type
- Each data type
- Secret masking
- Optional fields
- Min/max for numeric types

## Phase 2.6: Multi-Resource Integration

### Comprehensive Test

Function: `TestMultiResourceConversion`

**Purpose**: Verify all resource types work together correctly

### Resource Ordering

Generate HCL in logical dependency order:

1. **Variables** (no dependencies)
2. **Connector** (no dependencies)
3. **Flows** (reference connections and variables, may reference other flows)
4. **Applications** (standalone, no dependencies)
5. **Flow Policies** (reference flows and applications)

### Test Scenarios

1. **All Resources Present**:
   - Verify each resource type generates correctly
   - Verify ordering is correct
   - Verify references between resources

2. **Subset of Resources**:
   - Only flows and connections
   - Only applications and flow policies
   - Verify graceful handling of missing resource types

3. **Resource Ordering Validation**:
   - Ensure dependencies come before dependents
   - Verify Terraform can parse the order without errors

### Implementation

Function: `ConvertAll(exportData []byte) (string, error)`

**Algorithm**:
```go
func ConvertAll(exportData []byte) (string, error) {
    // Parse export data into resource groups
    data := parseExportData(exportData)
    
    var hcl strings.Builder
    
    // 1. Generate variables first
    for _, variable := range data.Variables {
        hcl.WriteString(generateVariableHCL(variable))
        hcl.WriteString("\n\n")
    }
    
    // 2. Generate connections
    for _, connection := range data.Connections {
        hcl.WriteString(generateConnectionHCL(connection))
        hcl.WriteString("\n\n")
    }
    
    // 3. Generate flows
    for _, flow := range data.Flows {
        hcl.WriteString(generateFlowHCL(flow))
        hcl.WriteString("\n\n")
    }
    
    // 4. Generate applications
    for _, app := range data.Applications {
        hcl.WriteString(generateApplicationHCL(app))
        hcl.WriteString("\n\n")
    }
    
    // 5. Generate flow policies
    for _, policy := range data.FlowPolicies {
        hcl.WriteString(generateFlowPolicyHCL(policy))
        hcl.WriteString("\n\n")
    }
    
    return hcl.String(), nil
}
```

## Success Criteria

- All resource types convert correctly
- Proper resource ordering maintained
- References between resources handled (TODO comments for unresolved refs)
- All tests passing for each phase
- Integration tests validate HCL syntax

## Next Steps

After completing these phases:
1. Move to Terraform validation tests (Phase 2.7)
2. Then proceed to Part 3 (API Export)
