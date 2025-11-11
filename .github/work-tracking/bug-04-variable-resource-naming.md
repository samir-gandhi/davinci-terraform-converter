# Bug 04: Duplicate Variable Resource Names

## Status: Completed

## Problem

Variable resource names are generated using only the variable name, ignoring the context. This creates duplicate resource names when multiple variables share the same name but have different contexts.

Example of duplicate resources:
```hcl
resource "pingone_davinci_variable" "pingcli__origin" {
  name    = "origin"
  context = "company"
}

resource "pingone_davinci_variable" "pingcli__origin" {
  name    = "origin"
  context = "flowInstance"
}
```

## Root Cause

In `internal/converter/variable_converter.go`:
- Line 123: `resourceName := utils.SanitizeResourceName(variable.Name)`
- Line 334: `resourceName := utils.SanitizeResourceName(variable.Name)`

The resource name generation only considers `variable.Name`, not `variable.Context`.

## Solution

Updated resource name generation to include both name and context:
`pingcli__origin_company` and `pingcli__origin_flowInstance`

Created a generic multi-key sanitization function for reusability.

## Test-Driven Development Approach

### Phase 1: Add Failing Tests
- [x] Create test cases for duplicate variable names with different contexts
- [x] Verify tests fail with current implementation

### Phase 2: Update Resource Naming Logic
- [x] Create new function `SanitizeMultiKeyResourceName(...keys string) string`
- [x] Create wrapper `SanitizeVariableResourceName(name, context string) string`
- [x] Update `generateVariableHCL()` to use new function
- [x] Update `generateVariableHCLWithVarReference()` to use new function
- [x] Update `GetVariableEligibleAttributes()` to use new function

### Phase 3: Update Variable Name References
- [x] Ensure extracted variable names also include context suffix
- [x] Update module variable name generation to match
- [x] Verify variable references remain consistent

### Phase 4: Verify Tests Pass
- [x] Run tests and verify all pass
- [x] Test integration with module generation
- [x] Validate HCL output format

## Files Modified

1. `internal/utils/sanitize.go`
   - Added `SanitizeMultiKeyResourceName(...keys string) string` - generic function
   - Added `SanitizeVariableResourceName(name, context string) string` - wrapper for variables

2. `internal/converter/variable_converter.go`
   - Updated line 123 in `generateVariableHCL()` to use `SanitizeVariableResourceName`
   - Updated line 334 in `generateVariableHCLWithVarReference()` to use `SanitizeVariableResourceName`
   - Updated line 71 in `GetVariableEligibleAttributes()` to use `SanitizeVariableResourceName`

3. `internal/utils/sanitize_test.go`
   - Added `TestSanitizeMultiKeyResourceName` with 9 test cases
   - Updated `TestSanitizeVariableResourceName` to verify wrapper compatibility

4. `internal/converter/variable_converter_test.go`
   - Added `TestDuplicateVariableNamesWithDifferentContexts` with 7 test cases
   - Updated existing tests to expect context-suffixed resource names

5. `internal/converter/multi_resource_converter_test.go`
   - Updated test expectations for new naming convention

6. `internal/converter/variable_eligible_test.go`
   - Updated test expectations for new naming convention

## Implementation Details

### Generic Multi-Key Function
```go
func SanitizeMultiKeyResourceName(keys ...string) string {
    if len(keys) == 0 {
        return "pingcli__"
    }

    // Sanitize each key individually
    sanitizedKeys := make([]string, len(keys))
    for i, key := range keys {
        sanitizedKeys[i] = regexp.MustCompile(`[^0-9A-Za-z_\-]`).ReplaceAllStringFunc(key, func(s string) string {
            return fmt.Sprintf("-%04X-", s)
        })
    }

    // Join with underscores and prefix
    return fmt.Sprintf("pingcli__%s", strings.Join(sanitizedKeys, "_"))
}
```

### Variable-Specific Wrapper
```go
func SanitizeVariableResourceName(name, context string) string {
    return SanitizeMultiKeyResourceName(name, context)
}
```

## Results

All tests pass:
- 7 new test cases for duplicate variable names
- 9 test cases for multi-key sanitization
- All existing tests updated and passing
- No duplicate resource names in generated HCL

## Example Output

Before fix:
```hcl
resource "pingone_davinci_variable" "pingcli__origin" {
  name    = "origin"
  context = "company"
}

resource "pingone_davinci_variable" "pingcli__origin" {  # DUPLICATE!
  name    = "origin"
  context = "flowInstance"
}
```

After fix:
```hcl
resource "pingone_davinci_variable" "pingcli__origin_company" {
  name    = "origin"
  context = "company"
}

resource "pingone_davinci_variable" "pingcli__origin_flowInstance" {
  name    = "origin"
  context = "flowInstance"
}
```

## Progress Log

### 2025-01-11
- Created work tracking document
- Identified root cause in variable_converter.go
- Defined phased TDD approach
- Implemented generic SanitizeMultiKeyResourceName function
- Updated all variable converter functions
- Added comprehensive test coverage
- Updated existing tests for new convention
- All tests passing ✅
