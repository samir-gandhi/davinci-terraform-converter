# Settings Field Format Fix

## Issue

The `settings` field was generating invalid HCL with nested curly braces:

```hcl
settings {
  {
    "csp": "worker-src 'self' blob:;",
    "logLevel": 2
  }
}
```

This is incorrect HCL syntax and would fail Terraform validation.

## Root Cause

The `generateSettings()` function in `internal/converter/converter.go` was treating settings as a **block** (`settings { ... }`) when it should be an **attribute assignment** (`settings = { ... }`).

The function was:
1. Writing `settings {` (opening a block)
2. Marshaling the settings map to JSON (which produces `{ ... }`)
3. Writing the JSON inside the block
4. Closing with `}`

This created the nested braces: `settings { { ... } }`

## Solution

Changed the `generateSettings()` function to use attribute assignment syntax:

**Before:**
```go
func generateSettings(hcl *strings.Builder, settings map[string]interface{}) error {
	hcl.WriteString("  settings {\n")
	// ... marshal JSON and write with indentation ...
	hcl.WriteString("  }\n")
	return nil
}
```

**After:**
```go
func generateSettings(hcl *strings.Builder, settings map[string]interface{}) error {
	hcl.WriteString("  settings = ")
	jsonBytes, err := json.MarshalIndent(settings, "  ", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	hcl.WriteString(string(jsonBytes))
	hcl.WriteString("\n\n")
	return nil
}
```

## Correct Output

Now generates valid HCL:

```hcl
resource "pingone_davinci_flow" "my_flow" {
  environment_id = var.environment_id
  
  name = "My Flow"
  
  settings = {
    "csp": "worker-src 'self' blob:;",
    "intermediateLoadingScreenCSS": "",
    "intermediateLoadingScreenHTML": "",
    "logLevel": 2
  }
}
```

## Test Coverage

### New Test Added

Created `TestSettingsAttributeFormat` that specifically validates:
1. Settings uses `settings = {` (attribute assignment)
2. No nested braces (`settings = {\n    {` pattern)
3. JSON content is properly formatted

### Updated Tests

Updated all existing tests that checked for `settings {`:
- `TestFlowWithSettings`
- `TestCompleteFlowWithAllAttributes`
- `TestMultiFlowExport`
- `TestRealMultiFlowFile`

All 26 tests now pass.

## TDD Process

1. **Red Phase**: Created `TestSettingsAttributeFormat` test that failed with current implementation
2. **Green Phase**: Fixed `generateSettings()` to use attribute assignment
3. **Refactor Phase**: Updated all existing tests to check for correct format
4. **Validation**: Verified all tests pass and real output is correct

## Related Enhancement: Terraform Validation Tests

This issue highlighted the need for integration tests that validate generated HCL against Terraform itself. Added Part 5 to the project prompt (`davinci-converter.prompt.md`) recommending:

- Integration tests that run `terraform validate` on generated HCL
- Tests that run `terraform plan` with mock configuration
- Tests that catch syntax issues that unit tests might miss

This would provide an additional layer of validation beyond Go unit tests.

## Files Changed

1. `internal/converter/converter.go` - Fixed `generateSettings()` function
2. `internal/converter/converter_test.go` - Added new test, updated existing tests
3. `internal/converter/real_file_test.go` - Updated test expectations
4. `.github/prompts/davinci-converter.prompt.md` - Added Part 5: Terraform validation tests

## Verification

Tested with:
```bash
# Run all tests
make all

# Verify real output
./davinci-convert -f .github/prompts/simple-demo-flow.json | grep -A 5 "settings"
```

Output now correctly shows:
```hcl
settings = {
  "csp": "default-src 'self';",
  "logLevel": 4
}
```
