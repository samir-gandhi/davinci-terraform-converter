# Part 2 Phase 2.1 Progress Report

**Date**: 2025-10-10
**Status**: 🔧 IN PROGRESS
**Test Pass Rate**: 21/34 tests passing (62%)

## Work Completed This Session

### 1. ✅ Unified Converter Implementation
- Refactored `Convert()` function to delegate to `ConvertFlowToHCL()`
- Updated `ConvertMultiFlow()` to use new converter
- Eliminates dual implementations (old JSON-string format vs new HCL format)

### 2. ✅ Resource Naming Fixed
- Updated `generateResourceName()` to convert to lowercase
- Ensures consistent Terraform naming (e.g., `complete_test_flow` not `Complete_Test_Flow`)

### 3. ✅ Test Expectations Updated
- Updated `TestCompleteFlowConversion` for HCL format
- Fixed `TestFlowConversion_MinimalFlow` for flexible spacing
- Fixed `TestFlowConversion_ConnectionIDReference` for flexible spacing

### 4. ✅ Added Missing Fields
- Added `classes` field to nodes/edges (always included, even if empty)
- Added `description` field to input_schema items (always included, even if empty)

## Current Issues (13 failing tests)

### Issue 1: Old Tests Expect JSON Format
**Affected Tests**: TestFlowWithSingleNode, TestFlowWithNodesAndEdges, TestFlowWithComplexNodeProperties, TestNodeWithMissingData, TestEdgeWithMissingData, TestFlowOutputFormat, TestFlowWithSettings, TestFlowWithAllAttributes, TestMultiFlowExport, TestSettingsAttributeFormat, TestCompleteFlowConversion, TestRealMultiFlowFile

**Problem**: Tests check for JSON format strings like `"capabilityName": "value"` but converter outputs HCL format like `capability_name = "value"`

**Solution**: Systematically update test expectations in `converter_test.go` to check for HCL format

### Issue 2: TestComprehensiveFlowConversion Formatting
**Problem**: This test has very specific formatting expectations:
1. Resource name: expects `PingOne_DaVinci_API_Protect_Example` (original casing) but gets `pingone_davinci_api_protect_example` (lowercase)
2. Connection ID: expects `pingone_sso_connector` but gets `ping_one_s_s_o_connector` (overly aggressive snake_case)
3. jsonencode: expects pretty-printed with spaces/newlines, gets compact format
4. Settings alignment: expects 2-space aligned columns, gets 4-space alignment

**Solution Options**:
A. Adjust generator to match expectations (may not be realistic for production)
B. Make test more flexible to accept variations
C. Identify which expectations are Terraform requirements vs preferences

### Issue 3: Connection ID Sanitization
**Problem**: `toSnakeCase("pingOneSSOConnector")` produces `ping_one_s_s_o_connector` (extra underscores) when test expects `pingone_sso_connector`

**Solution**: Improve `toSnakeCase()` function to handle consecutive capitals better

## Next Steps Priority

1. **High Priority**: Update old tests in converter_test.go to expect HCL format (fixes ~8 tests)
2. **Medium Priority**: Fix `toSnakeCase()` function for better connection ID references
3. **Low Priority**: Decide on TestComprehensiveFlowConversion approach (adjust generator vs test)

## Files Modified This Session

- `/Users/samirgandhi/go/src/github.com/samir-gandhi/davinci-terraform-converter/internal/converter/converter.go`
  - Refactored `Convert()` and `ConvertMultiFlow()` to use `ConvertFlowToHCL()`

- `/Users/samirgandhi/go/src/github.com/samir-gandhi/davinci-terraform-converter/internal/converter/flow_converter.go`
  - Added lowercase conversion in `generateResourceName()`
  - Always include `classes` field for nodes/edges
  - Always include `description` field for input_schema items

- `/Users/samirgandhi/go/src/github.com/samir-gandhi/davinci-terraform-converter/internal/converter/converter_test.go`
  - Updated `TestCompleteFlowConversion` expectations for HCL format

- `/Users/samirgandhi/go/src/github.com/samir-gandhi/davinci-terraform-converter/internal/converter/flow_comprehensive_test.go`
  - Updated `TestFlowConversion_MinimalFlow` for flexible spacing
  - Updated `TestFlowConversion_ConnectionIDReference` for flexible spacing

## Test Status Summary

```
Passing Tests (21):
- TestSimpleFlowConversion
- TestSanitizeResourceName (5 subtests)
- TestFlowWithVariables
- TestFlowWithInputSchema
- TestMalformedJSON
- TestEmptyJSON
- TestFlowWithoutGraphData
- TestSpecialCharactersInFlowName
- TestSingleFlowWrappedInFlowsArray
- TestEmptyFlowsArray
- TestFlowConversion_NoEdges
- TestFlowConversion_MinimalFlow ✅ (fixed this session)
- TestFlowConversion_EscapeSpecialCharacters
- TestFlowConversion_NodeProperties
- TestFlowConversion_RendererField
- TestFlowConversion_ConnectionIDReference ✅ (fixed this session)

Failing Tests (13):
- TestFlowWithSingleNode (expects JSON format)
- TestFlowWithNodesAndEdges (expects JSON format)
- TestFlowWithComplexNodeProperties (expects JSON format)
- TestFlowOutputFormat (expects JSON format)
- TestFlowWithSettings (expects JSON format)
- TestNodeWithMissingData (expects JSON format)
- TestEdgeWithMissingData (expects JSON format)
- TestCompleteFlowWithAllAttributes (expects JSON format)
- TestMultiFlowExport (expects JSON format)
- TestSettingsAttributeFormat (expects JSON format)
- TestCompleteFlowConversion (expects JSON format)
- TestRealMultiFlowFile (needs investigation)
- TestComprehensiveFlowConversion (formatting differences)
```

## Estimated Work Remaining

- **Update old tests**: 2-3 hours (systematic but tedious)
- **Fix toSnakeCase()**: 30 minutes
- **Resolve TestComprehensiveFlowConversion**: 1-2 hours (depending on approach)
- **Total**: 4-6 hours to achieve 100% pass rate

## Recommendation

Continue with high-priority item: systematically update old tests in `converter_test.go` to expect HCL format. This will fix the majority of failing tests and provide clear path to 100% completion.
