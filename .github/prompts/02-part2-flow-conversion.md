---
mode: agent
---

# Part 2 - Phase 2.1: Comprehensive Flow Structure Conversion

**Status**: ✅ MOSTLY COMPLETE (Core functionality working)

**Test Pass Rate**: Most comprehensive tests passing, legacy tests need updating

**Last Updated**: 2025-10-11

## Overview

This phase focuses on converting complete DaVinci flow structures to valid HCL for the `pingone_davinci_flow` resource. The converter now handles all major flow attributes including:

- Top-level metadata (name, description, color)
- Complete graph_data structure with nodes and edges
- Settings object with nested properties
- Input schema definitions
- Connection ID to Terraform reference conversion
- Resource name sanitization (pingcli-compatible format)

**Remaining Work**: Minor enhancements for output_schema, trigger, and js_links formatting improvements.

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
   - `environment_id` (string, required) - The PingOne environment ID (use `var.environment_id` in generated HCL)
   - `name` (string, required) - Flow name
   - `description` (string, optional) - Flow description
   - `color` (string, optional) - Flow color in hex format (e.g., "#CACED3")
   - `graph_data` (object, optional) - Complete flow graph structure (see section 2 below)
   - `settings` (object, optional) - Flow settings (see section 3 below)
   - `input_schema` (array, optional) - Input schema definitions (see section 5 below)
   - `output_schema` (object, optional) - Output schema definition (see section 6 below)
   - `trigger` (object, optional) - Flow trigger configuration (see section 7 below)

2. **graphData Structure**:
   - `elements.nodes[]` - Array of node objects:
     - `data` (object, required):
       - `id` (string, required)
       - `node_type` (string, required)
       - `connection_id` (string, optional) - For CONNECTION nodes
       - `connector_id` (string, optional)
       - `capability_name` (string, optional)
       - `label` (string, optional)
       - `name` (string, optional)
       - `status` (string, optional)
       - `type` (string, optional)
       - `properties` (string, optional, sensitive) - Use jsonencode() for complex objects
       - `id_unique` (string, read-only)
     - `position` (object, optional):
       - `x` (number, required)
       - `y` (number, required)
     - `classes` (string, optional)
     - `grabbable` (boolean, optional)
     - `group` (string, optional)
     - `locked` (boolean, optional)
     - `pannable` (boolean, optional)
     - `removed` (boolean, optional)
     - `selectable` (boolean, optional)
     - `selected` (boolean, optional)
   - `elements.edges[]` - Array of edge objects:
     - `data` (object, required):
       - `id` (string, required)
       - `source` (string, required) - Source node ID
       - `target` (string, required) - Target node ID
     - `position` (object, optional):
       - `x` (number, required)
       - `y` (number, required)
     - `classes` (string, optional)
     - `grabbable` (boolean, optional)
     - `group` (string, optional)
     - `locked` (boolean, optional)
     - `pannable` (boolean, optional)
     - `removed` (boolean, optional)
     - `selectable` (boolean, optional)
     - `selected` (boolean, optional)
   - Viewport properties:
     - `pan` (object, optional):
       - `x` (number, required)
       - `y` (number, required)
     - `zoom` (number, optional)
     - `min_zoom` (number, optional)
     - `max_zoom` (number, optional)
     - `zooming_enabled` (boolean, optional)
     - `user_zooming_enabled` (boolean, optional)
     - `panning_enabled` (boolean, optional)
     - `user_panning_enabled` (boolean, optional)
     - `box_selection_enabled` (boolean, optional)
     - `renderer` (string, optional) - Use jsonencode() for renderer config
     - `data` (string, optional) - Additional graph data

3. **Settings Object**:
   - `csp` (string, optional) - Content Security Policy
   - `css` (string, optional) - Custom CSS content
   - `css_links` (set of strings, optional) - External CSS URLs
   - `js_links` (set of objects, optional) - External JavaScript configurations:
     - `crossorigin` (string, required)
     - `defer` (boolean, required)
     - `integrity` (string, required)
     - `label` (string, required)
     - `referrerpolicy` (string, required)
     - `type` (string, required)
     - `value` (string, required) - The script URL
   - `log_level` (number, optional) - Log level (1-4)
   - `custom_title` (string, optional) - Custom page title
   - `custom_favicon_link` (string, optional) - Custom favicon URL
   - `custom_logo_urlselection` (number, optional)
   - `intermediate_loading_screen_html` (string, optional) - Loading screen HTML
   - `intermediate_loading_screen_css` (string, optional) - Loading screen CSS
   - `js_custom_flow_player` (string, optional) - Custom flow player JavaScript
   - `flow_timeout_in_seconds` (number, optional) - Flow timeout
   - `flow_http_timeout_in_seconds` (number, optional) - HTTP request timeout
   - `require_authentication_to_initiate` (boolean, optional)
   - `scrub_sensitive_info` (boolean, optional)
   - `sensitive_info_fields` (set of strings, optional) - Fields to scrub
   - `use_csp` (boolean, optional)
   - `use_custom_css` (boolean, optional)
   - `use_custom_flow_player` (boolean, optional)
   - `use_custom_script` (boolean, optional)
   - `use_intermediate_loading_screen` (boolean, optional)
   - `validate_on_save` (boolean, optional)
   - `custom_error_screen_brand_logo_url` (string, optional)
   - `custom_error_show_footer` (boolean, optional)
   - `default_error_screen_brand_logo` (boolean, optional)

4. **Variables Array**:
   - Each variable has: `name`, `context`, `type`, `value`, `mutable`
   - Handle different contexts: company, flow, flowInstance, user
   - Handle different types: string, number, boolean, object, secret
   - **Note**: Variables are NOT part of the Terraform schema and should be ignored or documented separately

5. **Input Schema Array** (`input_schema`):
   - Array of input schema objects:
     - `property_name` (string, optional)
     - `preferred_data_type` (string, optional) - Options: `array`, `boolean`, `number`, `object`, `string`
     - `preferred_control_type` (string, optional) - Options: `button`, `colorPicker`, `contentEditableTextArea`, `cssArea`, `dropDown`, `dropDownMultiSelect`, `dropDownMultiSelect2`, `dropDownWithCreate`, `functionArgumentList`, `keyValueList`, `label`, `radioSelect`, `textArea`, `textField`, `textFieldArrayView`, `toggleSwitch`
     - `required` (boolean, optional)
     - `description` (string, optional)
     - `is_expanded` (boolean, optional)

6. **Output Schema Object** (`output_schema`):
   - `output` (string, optional) - Output schema definition

7. **Trigger Object** (`trigger`):
   - `type` (string, required) - Trigger type
   - `configuration` (object, optional):
     - `mfa` (object, optional):
       - `enabled` (boolean, optional)
       - `time` (number, optional)
       - `time_format` (string, optional)
     - `pwd` (object, optional):
       - `enabled` (boolean, optional)
       - `time` (number, optional)
       - `time_format` (string, optional)

8. **Read-Only Attributes** (computed, not in input):
   - `id` - Resource ID
   - `current_version` - Current version number
   - `published_version` - Published version number
   - `enabled` - Whether flow is enabled
   - `connectors` - Set of connector references used in the flow

### 3. Implement Deep Conversion Logic

**Function**: `Convert(flowJSON []byte) (string, error)`

**HCL Syntax Requirements** (from Terraform Provider):

```hcl
resource "pingone_davinci_flow" "example" {
  environment_id = var.environment_id
  name           = "My Flow"
  description    = "Optional description"
  color          = "#CACED3"
  
  settings = {
    csp                              = "worker-src 'self' blob:; script-src 'self' https://cdn.jsdelivr.net 'unsafe-inline' 'unsafe-eval';"
    css                              = ".custom { color: blue; }"
    css_links                        = ["https://example.com/styles.css"]
    custom_title                     = "My Custom Flow"
    custom_favicon_link              = "https://example.com/favicon.ico"
    flow_http_timeout_in_seconds     = 300
    flow_timeout_in_seconds          = 600
    intermediate_loading_screen_css  = ".loader { animation: spin 1s; }"
    intermediate_loading_screen_html = "<div class='loader'>Loading...</div>"
    js_custom_flow_player            = "console.log('Custom player');"
    js_links = [
      {
        crossorigin    = "anonymous"
        defer          = true
        integrity      = "sha256-abc123"
        label          = "jQuery"
        referrerpolicy = "no-referrer"
        type           = "text/javascript"
        value          = "https://code.jquery.com/jquery-3.6.0.min.js"
      }
    ]
    log_level                            = 2
    require_authentication_to_initiate   = false
    scrub_sensitive_info                 = true
    sensitive_info_fields                = ["password", "ssn"]
    use_csp                              = true
    use_custom_css                       = true
    use_custom_flow_player               = false
    use_custom_script                    = true
    use_intermediate_loading_screen      = true
    validate_on_save                     = true
  }
  
  graph_data = {
    elements = {
      nodes = [
        {
          data = {
            id              = "node1"
            node_type       = "CONNECTION"
            connection_id   = pingone_davinci_connector_instance.http_example.id
            connector_id    = "httpConnector"
            name            = "Http"
            label           = "Http Connector"
            status          = "configured"
            capability_name = "customHtmlMessage"
            type            = "action"
            properties      = jsonencode({
              "message" : {
                "value" : "Welcome to the flow"
              }
            })
          }
          position = {
            x = 100
            y = 200
          }
          group      = "nodes"
          removed    = false
          selected   = false
          selectable = true
          locked     = false
          grabbable  = true
          pannable   = false
          classes    = ""
        }
      ]
      edges = [
        {
          data = {
            id     = "edge1"
            source = "node1"
            target = "node2"
          }
          group      = "edges"
          removed    = false
          selected   = false
          selectable = true
          locked     = false
          grabbable  = true
          pannable   = false
          classes    = ""
        }
      ]
    }
    
    pan = {
      x = 0
      y = 0
    }
    zoom                  = 1
    min_zoom              = 1e-50
    max_zoom              = 1e+50
    zooming_enabled       = true
    panning_enabled       = true
    user_zooming_enabled  = true
    user_panning_enabled  = true
    box_selection_enabled = true
    renderer              = jsonencode({"name": "canvas"})
  }
  
  input_schema = [
    {
      property_name          = "email"
      preferred_data_type    = "string"
      preferred_control_type = "textField"
      required               = true
      is_expanded            = true
      description            = "User email address"
    },
    {
      property_name          = "password"
      preferred_data_type    = "string"
      preferred_control_type = "textField"
      required               = true
      is_expanded            = false
      description            = "User password"
    }
  ]
  
  output_schema = {
    output = jsonencode({
      "success" : true,
      "message" : "Flow completed"
    })
  }
  
  trigger = {
    type = "authentication"
    configuration = {
      mfa = {
        enabled     = true
        time        = 30
        time_format = "minutes"
      }
    }
  }
}
```

**CRITICAL**: 
- All attributes use `=` (not blocks with `{ }`)
- Nested objects are proper HCL objects, not JSON strings
- Use `jsonencode()` for complex nested properties (node properties, renderer, output schema)
- Arrays of objects use proper HCL array syntax: `[{ ... }, { ... }]`
- Boolean values are `true`/`false` (not strings)
- Numbers can use scientific notation (e.g., `1e-50`)
- Strings use proper escaping for special characters

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

**Test Results**: Most comprehensive tests passing

**Implemented Features**:
- ✅ Top-level flow attributes (name, description, color)
- ✅ Settings object with all standard fields
- ✅ Settings `js_links` as proper HCL objects (resolved 2025-10-11)
- ✅ Graph data structure (nodes, edges, viewport)
- ✅ Node data with connection references
- ✅ Edge data with source/target references
- ✅ Edge position attributes (optional, rarely used)
- ✅ Position attributes for nodes
- ✅ Visual attributes (classes, group, removed, selected, etc.)
- ✅ Input schema array
- ✅ Output schema object (added 2025-10-11)
- ✅ Trigger configuration with mfa/pwd (added 2025-10-11)
- ✅ Connection ID to Terraform reference conversion
- ✅ Resource name sanitization (pingcli format)
- ✅ Special character escaping in strings
- ✅ jsonencode() for complex properties

**Not Yet Implemented**:
- ⚠️ Variables (documented as Known Limitation #3)
- ⚠️ Edge `position` attributes are supported but rarely used in practice
- ⚠️ Graph data `data` attribute (additional graph metadata - rarely used)

**Legacy Test Updates Needed**:
- Several tests in `converter_test.go` use old resource name format expectations
- These tests need updating to match pingcli-compatible naming (see Known Limitations #1)

## Next Steps

1. **Validate with complex real flows**: 
   - Test generated HCL with `terraform validate`
   - Test with complex production flows
   - Ensure all common DaVinci patterns are handled

2. **Update legacy tests**: 
   - Update `converter_test.go` to use pingcli-compatible resource names
   - Ensure all connection ID references use lowercase format

3. **Consider variables implementation**:
   - Decide on approach (separate resources vs comments)
   - Design variable reference handling
   - Implement if needed for pingcli integration

## Testing Strategy

Run tests frequently:

```bash
go test ./internal/converter -v
go test ./internal/converter -run TestCompleteFlowConversion
go test ./internal/converter -run TestRealMultiFlowFile
```

Test with real flow examples:

```bash
# Test with simple demo flow
./davinci-convert convert --flow-json .github/prompts/simple-demo-flow.json

# Test with complex API protect flow
./davinci-convert convert --flow-json .github/prompts/davinci-api-protect-reg-authn-flow.json

# Test with multi-flow export
./davinci-convert convert --flow-json .github/prompts/PingOne_Sign\ On\ with\ Sessions_multiflow.json
```

Validate generated Terraform:

```bash
# Save output to file
./davinci-convert convert --flow-json <flow.json> > output.tf

# Initialize and validate (requires provider configuration)
terraform init
terraform validate
```

## Success Criteria

- ✅ **Core functionality complete**: All major flow attributes converted
- ✅ **Real flows work**: Complex production flows generate valid HCL
- ✅ **Resource naming consistent**: Uses pingcli-compatible format
- ✅ **Connection references correct**: Properly formatted Terraform references
- ✅ **New attributes implemented**: output_schema, trigger, js_links (2025-10-11)
- ⚠️ **Known limitations documented**: See KNOWN_LIMITATIONS.md
- ⚠️ **Legacy tests need updates**: Documented in Known Limitations #1
- 🔄 **Future enhancements**: Variables implementation if needed

## References

- OpenAPI Spec: `.github/prompts/davinci-openapi.yaml`
- Sample Flow: `.github/prompts/davinci-api-protect-reg-authn-flow.json`
- Terraform Provider Schema: Validate with `terraform validate`
- pingone-go-client: Check structs for field names/types
