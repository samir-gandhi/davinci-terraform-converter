---
mode: agent
---

# Part 2 - Phase 2.1: Comprehensive Flow Structure Conversion

**Status**: 🔧 IN PROGRESS (Current Work Focus)

**Test Pass Rate**: 24/27 tests passing (88.9%)

## Overview

This phase focuses on converting complete DaVinci flow structures to valid HCL for the `pingone_davinci_flow` resource. The goal is to handle ALL flow attributes, not just basic metadata.

## TDD Approach

### 1. Create Test Case for Complete Flow

Location: `internal/converter/converter_test.go`

Test function: `TestCompleteFlowConversion`

**Test Structure**:
```go
func TestCompleteFlowConversion(t *testing.T) {
    // Define complete DaVinci flow JSON including:
    // - Top-level metadata (name, description, flowId, enabled, etc.)
    // - Complete graphData structure with nodes and edges
    // - Settings object with all nested properties
    // - Variables array
    // - InputSchema array
    // - OutputSchema structure
    
    // Define expected HCL output as string
    expectedHCL := `...` // Complete HCL with all structures
    
    // Call Convert() and assert output matches
    actualHCL, err := Convert(flowJSON)
    require.NoError(t, err)
    assert.Equal(t, expectedHCL, actualHCL)
}
```

### 2. Map Complete Flow Structure

**Reference Materials**:
- `pingone-go-client` structs for DaVinci flow structure
- DaVinci OpenAPI spec (@.github/prompts/davinci-openapi.yaml)

**Key Structures to Handle**:

1. **Top-Level Flow Attributes**:
   - `name` (string, required)
   - `description` (string, optional)
   - `flow_configuration_json` (object, required) - Contains graphData, settings, etc.
   - `environment_id` (string, required)
   - `deploy` (boolean, optional)

2. **graphData Structure**:
   - `elements.nodes[]` - Array of node objects:
     - `data.id` (string)
     - `data.nodeType` (string)
     - `data.connectionId` (string, for CONNECTION nodes)
     - `data.connectorId` (string)
     - `data.capabilityName` (string)
     - `data.properties` (object, capability-specific)
     - `position.x`, `position.y` (numbers)
     - `data.label`, `data.name` (strings)
   - `elements.edges[]` - Array of edge objects:
     - `data.id` (string)
     - `data.source` (string, node ID)
     - `data.target` (string, node ID)
   - Viewport properties: `pan`, `zoom`, `minZoom`, `maxZoom`
   - Boolean flags: `panningEnabled`, `zoomingEnabled`, etc.

3. **Settings Object**:
   - `csp` (string, optional)
   - `css` (string, optional)
   - `cssLinks[]` (array of strings)
   - `jsLinks[]` (array of objects with label, value, defer, etc.)
   - `logLevel` (integer, 1-4)
   - `customTitle` (string, optional)
   - `intermediateLoadingScreenHTML` (string, optional)
   - `intermediateLoadingScreenCSS` (string, optional)
   - Various boolean flags (useCustomCSS, validateOnSave, etc.)
   - Timeout settings (flowTimeoutInSeconds, flowHttpTimeoutInSeconds)

4. **Variables Array**:
   - Each variable has: `name`, `context`, `type`, `value`, `mutable`
   - Handle different contexts: company, flow, flowInstance, user
   - Handle different types: string, number, boolean, object, secret

5. **Input/Output Schemas**:
   - `inputSchema[]` - Array of schema objects:
     - `propertyName`, `preferredDataType`, `required`, `description`
   - `outputSchema` - Object with output properties

### 3. Implement Deep Conversion Logic

**Function**: `Convert(flowJSON []byte) (string, error)`

**HCL Syntax Requirements** (from Terraform Provider):

```hcl
resource "pingone_davinci_flow" "example" {
  environment_id = pingone_environment.my_env.id
  name           = "My Flow"
  description    = "Optional description"
  
  # flow_configuration_json is an OBJECT, not a JSON string
  flow_configuration_json = {
    graphData = {
      elements = {
        nodes = [
          {
            data = {
              id = "node1"
              nodeType = "CONNECTION"
              connectionId = "abc123"
              # ... all node properties
            }
            position = {
              x = 100
              y = 200
            }
          }
        ]
        edges = [
          {
            data = {
              id = "edge1"
              source = "node1"
              target = "node2"
            }
          }
        ]
      }
      pan = { x = 0, y = 0 }
      zoom = 1
      # ... other viewport properties
    }
    settings = {
      logLevel = 2
      csp = "worker-src 'self' blob:; ..."
      # ... all settings properties
    }
    inputSchema = [
      {
        propertyName = "email"
        preferredDataType = "string"
        required = true
      }
    ]
    # ... other configuration
  }
  
  deploy = true
}
```

**CRITICAL**: 
- `flow_configuration_json` is an **attribute** (uses `=`), not a block (uses `{ }`)
- The value is an **object/map**, not a JSON string
- All nested structures should be HCL objects/arrays, not JSON strings
- No large JSON strings - everything must be proper HCL types

**Conversion Logic**:

1. Parse JSON flow into Go structs
2. Convert to HCL object structure:
   - Use maps for objects: `map[string]interface{}`
   - Use slices for arrays: `[]interface{}`
   - Preserve types: string, bool, number (int/float)
   - Handle null/optional fields (omit or use null)
3. Generate HCL syntax:
   - Simple attributes: `name = "value"`
   - Nested objects: `settings = { ... }`
   - Arrays of objects: `nodes = [{ ... }, { ... }]`
4. Proper formatting and indentation

### 4. Handle Edge Cases

**Test Scenarios**:

1. **Empty or Null Fields**:
   - Optional fields not present
   - Empty arrays: `nodes = []`
   - Null values: omit or explicit `null`

2. **Large graphData Structures**:
   - Flows with hundreds of nodes
   - Deep nesting in node properties
   - Performance considerations

3. **Special Characters in Strings**:
   - Quotes, backslashes, newlines
   - HTML/CSS/JavaScript in settings
   - Proper escaping for HCL strings

4. **Boolean and Numeric Edge Cases**:
   - Boolean: true/false (not strings)
   - Numbers: integers vs floats
   - Scientific notation handling

5. **Nested Arrays and Objects**:
   - Arrays of objects with nested arrays
   - Complex node properties (formFieldsList, variableInputList)
   - Deeply nested settings (jsLinks with multiple properties)

## Current Implementation Status

**Passing Tests** (24/27):
- Basic flow metadata conversion
- Simple graphData structures
- Basic settings handling
- Most node types
- Most edge cases

**Failing Tests** (3/27):
- Complex nested node properties
- Some settings edge cases
- Specific attribute handling

## Next Steps

1. **Review failing tests**: Identify specific issues
2. **Fix conversion logic**: Handle remaining edge cases
3. **Add comprehensive tests**: Cover all flow variations
4. **Validate with Terraform**: Ensure generated HCL is valid

## Testing Strategy

Run tests frequently:
```bash
go test ./internal/converter -v
go test ./internal/converter -run TestCompleteFlowConversion
```

Test with real flow examples:
```bash
# Use the sample flow in .github/prompts/davinci-api-protect-reg-authn-flow.json
go run main.go --flow-json .github/prompts/davinci-api-protect-reg-authn-flow.json --out test-output.tf
```

## Success Criteria

- All 27 tests passing (100% pass rate)
- Complex nested structures handled correctly
- Special characters properly escaped
- Valid HCL syntax (no JSON strings)
- Tested with real DaVinci flows

## References

- OpenAPI Spec: `.github/prompts/davinci-openapi.yaml`
- Sample Flow: `.github/prompts/davinci-api-protect-reg-authn-flow.json`
- Terraform Provider Schema: Validate with `terraform validate`
- pingone-go-client: Check structs for field names/types
