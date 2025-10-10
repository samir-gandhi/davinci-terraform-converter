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

### Test Case

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

## Phase 2.4: Connection (Connector Instance) Resource Conversion

**Resource**: `pingone_davinci_connection`

### Test Case

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

### Test Case

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
2. **Connections** (no dependencies)
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
