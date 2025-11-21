# Bug Fix 08 - Connector Instance Properties Formatting

## Summary

Fixed the `pingone_davinci_connector_instance.properties` field formatting to preserve the full API response structure (`{"type": "...", "value": "..."}`) instead of flattening to simple key-value pairs.

## Changes Made

### 1. Properties Structure Preservation

**Before:**
```hcl
properties = jsonencode({
  "clientId"     : "3642f58b-b0c2-4a35-b1b1-e24d051de546",
  "clientSecret" : "TODO: Replace with actual client secret",
  "envId"        : "4111cd46-25bf-4a5b-8c74-184a9d0c1826",
  "region"       : "NA"
})
```

**After:**
```hcl
properties = jsonencode({
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
```

### 2. Dynamic Variable Extraction

Implemented **structure-based** variable extraction:
- ANY property following the standard `{"type": "...", "value": "..."}` structure is automatically extracted as a variable
- No hardcoded list of "variable-eligible" properties needed
- Configuration only required for:
  - Secret property names (marked `sensitive = true`)
  - Excluded properties (computed/read-only fields)
  - Unstructured properties (future enhancement)

### 3. Files Modified

#### Core Implementation
- `internal/converter/connector_instance_converter.go`
  - Updated `writePropertiesBlock()` to preserve type/value structure
  - Updated `writePropertiesBlockWithVariables()` to inject variables into `value` field
  - Updated `GetConnectorInstanceVariableEligibleAttributes()` to use dynamic extraction

#### Configuration
- `internal/converter/property_mapping_config.go` (NEW)
  - `PropertyMappingConfig` - defines exceptions to dynamic extraction
  - `DefaultPropertyMappingConfig()` - lists secret and excluded properties
  - `HasStandardStructure()` - validates property structure
  - `GenerateVariableName()` - creates variable names dynamically

#### Tests
- `internal/converter/connector_properties_test.go` (NEW)
  - Tests standard structure preservation
  - Tests variable injection into nested value fields
  - Tests complex and edge cases

#### Updated Tests
- `internal/converter/connector_instance_converter_test.go`
  - Updated expectations for new nested structure
- `internal/converter/variable_eligible_test.go`
  - Updated to expect dynamic extraction (timeout now extracted)

#### Documentation
- `internal/converter/PROPERTY_MAPPING.md` (NEW)
  - Complete guide to property-to-variable mapping
  - Examples and migration notes

## Test Results

All tests passing:
```
✓ TestConnectorInstancePropertiesFormatting
✓ TestConnectorInstancePropertiesWithVariables
✓ TestConnectorInstancePropertiesComplexValues
✓ TestConnectorInstancePropertiesEmptyAndNil
✓ TestVariableEligibleAttributesWithNestedStructure
✓ TestConnectorInstanceConversion
✓ TestGetConnectorInstanceVariableEligibleAttributes
```

## Benefits

1. **Accurate API Representation**: Properties now match the exact structure returned by the DaVinci API
2. **Future-Proof**: Dynamic extraction works for any new connector without code changes
3. **Maintainable**: Only exceptions need configuration updates
4. **Type-Safe**: Preserves type information from the API
5. **Variable Flexibility**: Variables are injected into the value field while preserving the full structure

## Migration Notes

This change affects the generated Terraform HCL format. Modules generated with the new format will:
- Preserve the full `{"type": "...", "value": "..."}` structure
- Work correctly with the `pingone_davinci_connector_instance` resource
- Extract more properties as variables (anything with standard structure)

## Future Enhancements

Support for unstructured properties (like `customAuth` in `genericConnector`) can be added via the `UnstructuredPropertyPaths` configuration when needed.
