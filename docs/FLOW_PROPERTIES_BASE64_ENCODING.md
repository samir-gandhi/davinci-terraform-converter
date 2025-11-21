# Flow Properties Encoding Solution

**Date**: 2025-11-14  
**Status**: ~~Base64 encoding (Deprecated)~~ **Updated to jsonencode() HCL map literals**  
**Previous Issue**: Base64 encoding made properties unreadable and hard to debug

---

## Update: jsonencode() HCL Map Literals (Current Implementation)

**As of November 2025**, the generator now uses `jsonencode()` with properly formatted HCL map literals instead of base64 encoding. This provides:

- ✅ **Readable output**: Properties are visible as structured HCL
- ✅ **Easy debugging**: Can see actual values without decoding
- ✅ **Git-friendly**: Diffs show actual property changes
- ✅ **Editable**: Users can modify properties in the HCL file
- ✅ **No special character issues**: Proper escaping handles single quotes, newlines, etc.

### Current Format

```hcl
resource "pingone_davinci_flow" "my_flow" {
  environment_id = var.pingone_environment_id
  
  name = "My Flow"
  
  graph_data = {
    elements = {
      nodes = [
        {
          data = {
            id = "node123"
            properties = jsonencode({
              "code" = {
                "value" = "module.exports = a = async ({ params }) => {\n  return { 'message': '' };\n}"
              },
              "formFieldsList" = {
                "value" = "[{\"type\":\"text\",\"label\":\"Username\"}]"
              }
            })
          }
        }
      ]
    }
  }
}
```

The properties are written as HCL map literals with proper escaping for special characters, making them both readable and valid Terraform syntax.

---

## Historical Context: Base64 Encoding (Deprecated)

When exporting DaVinci flows to Terraform HCL, the `properties` field within flow nodes contains complex JSON data including:
- JavaScript code with single quotes (`'`)
- Multi-line strings with newlines
- Special characters that HCL's parser interprets as syntax

### Original Error

```
Error: Invalid character

  on flows_export.tf line 5867, in resource "pingone_davinci_flow" "...":
  5867:             properties = jsonencode({"code":{"value":"module.exports = a = async ({ params }) => {
                   return { 'message': '' }; ..."}})

Single quotes are not valid. Use double quotes (") to enclose strings.
```

### Root Cause

The flow properties contain JavaScript code (from HTTP connector nodes, function nodes, etc.) that includes:
- Single quotes: `'message': ''`
- Template literals: `` `Your account will unlock in ${formattedTime}.` ``
- Complex nested JSON structures

When using `jsonencode()` in HCL, Terraform's parser attempts to parse the content as HCL syntax before encoding, causing it to interpret single quotes and other characters as invalid HCL syntax rather than content within a JSON string.

---

## Solutions Attempted

### 1. ❌ Direct `jsonencode()` Usage
```hcl
properties = jsonencode({"code":{"value":"return { 'message': '' };"}})
```
**Issue**: HCL parser sees single quotes inside the function call and rejects them as invalid HCL syntax.

### 2. ❌ Heredoc Syntax (Unquoted)
```hcl
properties = <<-EOT
{"code":{"value":"return { 'message': '' };"}}
EOT
```
**Issue**: HCL still parses and interprets the content, seeing single quotes as invalid syntax.

### 3. ❌ Heredoc Syntax (Quoted Delimiter)
```hcl
properties = <<-'EOT'
{"code":{"value":"return { 'message': '' };"}}
EOT
```
**Issue**: Terraform's HCL parser still processes the content even with quoted heredoc delimiters.

### 4. ❌ `jsondecode()` Wrapper
```hcl
properties = jsondecode(<<-EOT
{"code":{"value":"return { 'message': '' };"}}
EOT
)
```
**Issue**: Same problem - HCL parser still interprets the content before jsondecode runs.

---

## Solution Implemented: Base64 Encoding

### Approach

Encode the JSON properties as base64 strings and use Terraform's `base64decode()` function to decode them at apply time.

### Generated HCL Format

```hcl
resource "pingone_davinci_flow" "pingcli__My-0020-Flow" {
  environment_id = "62f10a04-6c54-40c2-a97d-80a98522ff9a"
  
  name = "My Flow"
  
  subflow_link {
    id = "node123"
    
    # Properties base64 encoded to handle complex JSON with special characters
    properties = base64decode("eyJjb2RlIjp7InZhbHVlIjoibW9kdWxlLmV4cG9ydHMgPSBhID0gYXN5bmMgKHsgcGFyYW1zIH0pID0+IHtcbiAgICByZXR1cm4geyAnbWVzc2FnZSc6ICcnIH07XG59In19")
  }
}
```

### Implementation Details

**File**: `internal/converter/flow_converter.go`

**Function Modified**: `writePropertiesField()`

```go
func writePropertiesField(hcl *strings.Builder, properties interface{}, indentLevel int) error {
    indent := strings.Repeat("  ", indentLevel)
    
    // Convert properties to JSON
    propertiesJSON, err := json.Marshal(properties)
    if err != nil {
        return fmt.Errorf("failed to marshal properties: %w", err)
    }
    
    // Base64 encode the JSON to avoid HCL parsing issues with special characters
    encoded := base64.StdEncoding.EncodeToString(propertiesJSON)
    
    // Write as base64decode() call with comment explaining why
    hcl.WriteString(fmt.Sprintf("%s# Properties base64 encoded to handle complex JSON with special characters\n", indent))
    hcl.WriteString(fmt.Sprintf("%sproperties = base64decode(\"%s\")\n", indent, encoded))
    
    return nil
}
```

**Files Modified**:
1. `internal/converter/flow_converter.go` - Added base64 encoding logic
2. `internal/exporter/flow_exporter.go` - Added duplicate name handling and environment_id logic

---

## Benefits

### ✅ Advantages

1. **Terraform Validation Passes**: All HCL syntax errors eliminated
2. **Handles All Special Characters**: Single quotes, backticks, newlines, unicode, etc.
3. **No Manual Escaping Needed**: Base64 encoding handles all edge cases automatically
4. **Provider Compatible**: The provider receives the decoded JSON correctly at apply time
5. **Deterministic**: Same input always produces same base64 output (no random formatting)

### ❌ Disadvantages

1. **Reduced Readability**: Cannot see the actual JSON content without decoding
2. **Harder to Edit**: Users cannot easily modify properties in the HCL file
3. **Diff Visibility**: Git diffs show base64 string changes, not semantic JSON changes
4. **Debugging Difficulty**: Need to manually decode to inspect property values
5. **Not Terraform Idiomatic**: Most Terraform resources use human-readable HCL syntax

---

## User Experience Impact

### Before (Attempted - Failed Validation)
```hcl
subflow_link {
  id = "node123"
  properties = jsonencode({
    "code" = {
      "value" = "module.exports = async ({ params }) => {\n  return { 'message': '' };\n}"
    }
    "nodeTitle" = {
      "value" = "Calculate Expiration"
    }
  })
}
```
**Pros**: Human readable, easy to understand  
**Cons**: Fails terraform validate with "Invalid character" errors

### After (Current - Passes Validation)
```hcl
subflow_link {
  id = "node123"
  # Properties base64 encoded to handle complex JSON with special characters
  properties = base64decode("eyJjb2RlIjp7InZhbHVlIjoibW9kdWxlLmV4cG9ydHMgPSBhc3luYyAoeyBwYXJhbXMgfSkgPT4ge1xuICByZXR1cm4geyAnbWVzc2FnZSc6ICcnIH07XG59In0sIm5vZGVUaXRsZSI6eyJ2YWx1ZSI6IkNhbGN1bGF0ZSBFeHBpcmF0aW9uIn19")
}
```
**Pros**: Passes validation, handles all edge cases  
**Cons**: Not human readable, difficult to edit

---

## Alternative Solutions to Consider

### Option 1: Use External JSON Files

Instead of embedding in HCL, reference external JSON files:

```hcl
subflow_link {
  id = "node123"
  properties = file("flows/my_flow_node123_properties.json")
}
```

**Pros**: 
- Fully readable JSON in separate files
- Easy to edit with JSON tools
- Clear git diffs

**Cons**:
- Generates many files (one per node with properties)
- More complex file structure
- Terraform `file()` function still might have parsing issues

### Option 2: Selective Base64 Encoding

Only encode properties that contain problematic characters:

```hcl
# Simple properties - human readable
properties = jsonencode({
  "nodeTitle" = {
    "value" = "User Login"
  }
})

# Complex properties with JavaScript - base64 encoded
properties = base64decode("...")
```

**Pros**:
- Best of both worlds
- Readable when possible

**Cons**:
- Inconsistent format
- Complex detection logic needed
- Users still encounter base64 in many cases

### Option 3: Terraform Provider Enhancement

Request that the PingOne Terraform provider accept properties as a simple string type instead of complex nested structure:

```hcl
properties_json = <<-EOT
{
  "code": {
    "value": "module.exports = async () => { return { 'message': '' }; }"
  }
}
EOT
```

**Pros**:
- Most natural solution
- Fully readable
- Standard Terraform pattern

**Cons**:
- Requires provider changes
- Not under our control
- May take time to implement

### Option 4: Custom HCL Escaping Function

Create a comprehensive escaping function that properly escapes all special characters for HCL:

```go
func escapeForHCL(jsonStr string) string {
    // Escape single quotes, backticks, special chars
    // Convert to valid HCL string literal
}
```

**Pros**:
- Maintains readability
- No base64 needed

**Cons**:
- Very complex to implement correctly
- May still miss edge cases
- HCL parsing rules are complex

---

## Testing Results

### Terraform Validation Status

All 5 terraform validation tests now **PASSING** with base64 encoding:

```
=== RUN   TestTerraformValidateVariablesFromAPI
--- PASS: TestTerraformValidateVariablesFromAPI (4.52s)

=== RUN   TestTerraformValidateConnectorInstancesFromAPI  
--- PASS: TestTerraformValidateConnectorInstancesFromAPI (4.85s)

=== RUN   TestTerraformValidateApplicationsFromAPI
--- PASS: TestTerraformValidateApplicationsFromAPI (3.41s)

=== RUN   TestTerraformValidateFlowsFromAPI
--- PASS: TestTerraformValidateFlowsFromAPI (10.23s)

=== RUN   TestTerraformValidateAllResourcesFromAPI
--- PASS: TestTerraformValidateAllResourcesFromAPI (11.45s)
```

### Test Environment

- **Flows Tested**: 8 flows from production environment
- **Flow Size**: 435KB total HCL output
- **Complex Properties**: JavaScript code, template literals, single quotes, unicode
- **Terraform Version**: Compatible with 1.0+
- **Provider**: pingidentity/pingone (development override)

---

## Recommendations

### Short Term (Current Implementation)

✅ **Keep base64 encoding** for now to ensure:
- All terraform validation passes
- Exports work reliably for all flow types
- No user-facing errors during terraform apply

Include clear documentation:
- Comment in HCL explaining base64 encoding
- Documentation on how to decode for inspection
- Tool/script to decode base64 properties for debugging

### Medium Term (User Experience Improvement)

Consider implementing:
1. **Decode utility command**: `davinci-converter decode-properties <file>` to show readable JSON
2. **Documentation**: Add examples showing how to inspect/modify base64 properties
3. **Selective encoding**: Only base64 encode when special characters detected

### Long Term (Ideal Solution)

Engage with PingOne Terraform provider team:
1. Request `properties_json` string field addition to provider schema
2. Allow raw JSON string assignment instead of complex nested structures
3. Let provider handle JSON parsing/validation internally

This would eliminate the need for base64 encoding entirely while maintaining compatibility.

---

## Migration Path

If base64 encoding is deemed unacceptable:

1. **Phase 1**: Document limitations and keep current implementation
2. **Phase 2**: Implement selective encoding (only problematic properties)
3. **Phase 3**: Work with provider team on schema enhancement
4. **Phase 4**: Migrate to provider-native JSON string support

---

## Related Issues

- **Phase 3.4c**: Terraform Validation Testing
- **Variables**: Fixed attribute name mismatches (bool/float32/json_object)
- **Connector Instances**: Fixed empty environment_id 
- **Applications**: Fixed duplicate resource names
- **Flows**: Fixed environment_id quoting, duplicate names, and properties encoding

---

## Appendix: How to Decode Base64 Properties

### Command Line
```bash
echo "eyJjb2RlIjp7InZhbHVlIjoibW9kdWxlLmV4cG9ydHMgPSBhc3luYyAoKSA9PiB7IHJldHVybiB7ICdtZXNzYWdlJzogJycgfTsgfSJ9fQ==" | base64 -d | jq .
```

### Terraform Console
```bash
terraform console
> base64decode("eyJjb2RlIjp7InZhbHVlIjoibW9kdWxlLmV4cG9ydHMgPSBhc3luYyAoKSA9PiB7IHJldHVybiB7ICdtZXNzYWdlJzogJycgfTsgfSJ9fQ==")
```

### Go Script
```go
package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
)

func main() {
    encoded := "eyJjb2RlIjp7InZhbHVlIjoibW9kdWxlLmV4cG9ydHMgPSBhc3luYyAoKSA9PiB7IHJldHVybiB7ICdtZXNzYWdlJzogJycgfTsgfSJ9fQ=="
    decoded, _ := base64.StdEncoding.DecodeString(encoded)
    
    var jsonData map[string]interface{}
    json.Unmarshal(decoded, &jsonData)
    
    pretty, _ := json.MarshalIndent(jsonData, "", "  ")
    fmt.Println(string(pretty))
}
```

---

## Questions for Decision

1. **Is base64 encoding acceptable for production use?**
   - Consider: user experience, debugging, maintenance

2. **Should we implement selective encoding?**
   - Encode only when special characters detected
   - Keep simple properties human-readable

3. **Should we provide decode utilities?**
   - CLI command to decode properties
   - Web tool for inspection

4. **Should we pursue provider enhancement?**
   - Timeline for engagement
   - Alternative workarounds

5. **Is documentation sufficient?**
   - Clear explanation of encoding
   - Examples for inspection/modification

---

**Decision Required**: Determine if base64 encoding is an acceptable user experience or if alternative solutions should be implemented before Phase 3 completion.
