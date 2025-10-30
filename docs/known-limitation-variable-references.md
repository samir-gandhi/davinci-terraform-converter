# Known Limitation: Variable References in Child Module Resources

## Current Status

**Issue**: Variable values appear in `variables.tf` and `module.tf`, but child module resources still contain hardcoded values instead of `var.{name}` references.

**Example**:
```hcl
# variables.tf (child module) - ✅ CORRECT
variable "davinci_variable_api_endpoint_value" {
  type        = string
  default     = "https://api.example.com"
  description = "Value for apiEndpoint DaVinci variable"
}

# module.tf (root module) - ✅ CORRECT  
module "davinci" {
  source = "./modules/davinci"
  davinci_variable_api_endpoint_value = "https://api.example.com"
}

# Child module resource - ❌ INCORRECT (has hardcoded value)
resource "pingone_davinci_variable" "apiEndpoint" {
  name  = "apiEndpoint"
  value = "https://api.example.com"  # Should be: var.davinci_variable_api_endpoint_value
}
```

## Root Cause

The export functions generate HCL once during the initial export using normal converters. To generate HCL with variable references, we need to:
1. Collect all variables first
2. Build a complete variable map
3. Regenerate HCL using the `GenerateXXXWithVariableReferences()` functions

Currently, HCL is generated during export without knowledge of which attributes will become variables.

## Workaround

**Manual Find-and-Replace**: After module generation, manually replace hardcoded values with variable references in child module resources.

## Proper Fix (Not Yet Implemented)

### Option 1: Two-Pass Generation
1. First pass: Export resources and extract variables
2. Build complete variable map from extracted variables
3. Second pass: Regenerate resource HCL with variable references

### Option 2: Store Raw JSON During Export
1. Store original JSON data alongside HCL during export
2. After variable extraction is complete, regenerate HCL from JSON with variable references
3. Requires adding JSON storage to `ExportedData`

### Option 3: Module-Aware Export Mode
1. Add `moduleMode bool` parameter to all export functions
2. When `moduleMode=true`, use variable-reference generation functions
3. Build variable maps incrementally during export
4. Pass variable map to each resource as it's generated

## Implementation Complexity

- **Option 1**: Medium complexity, requires refactoring export flow
- **Option 2**: High complexity, significant memory overhead for large environments
- **Option 3**: Medium-high complexity, requires coordinating variable map across exporters

**Estimated effort**: 8-12 hours for complete implementation and testing

## Current Functionality

✅ **What Works**:
- Variable extraction from resources
- Variable collection during export
- `variables.tf` generation with all extracted variables
- `module.tf` generation with variable values
- Module structure and organization

❌ **What Doesn't Work**:
- Child module resources using `var.{name}` references
- Actual parameterization of child module resources

## Impact

**Severity**: Medium
- Generated modules are structurally correct
- Variables are defined and passed correctly
- Child modules don't actually use the variables (defeating the purpose)
- Manual editing required to make modules truly reusable

## Next Steps

1. **Short term**: Document workaround for users
2. **Medium term**: Implement Option 3 (module-aware export mode)
3. **Long term**: Consider Option 2 for maximum flexibility

## Related Files

- `internal/exporter/variable_exporter.go` - Needs module mode support
- `internal/exporter/connector_exporter.go` - Needs module mode support
- `internal/exporter/module_export.go` - Orchestrates module generation
- `internal/converter/variable_converter.go` - Has `GenerateVariableHCLWithVariableReferences()`
- `internal/converter/connector_instance_converter.go` - Has `GenerateConnectorInstanceHCLWithVariableReferences()`
