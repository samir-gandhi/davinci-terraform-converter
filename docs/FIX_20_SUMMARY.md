# Fix 20: Schema Update (Nodes/Edges as Maps)

## Overview

This fix updates the HCL generation for `pingone_davinci_flow` resources to align with the latest provider schema changes.

## Changes Implemented

### 1. HCL Structure Update

- **Nodes**: Changed from a List (`nodes = [...]`) to a Map (`nodes = { "id" = { ... } }`).
- **Edges**: Changed from a List (`edges = [...]`) to a Map (`edges = { "id" = { ... } }`).
- **Keys**: The map keys are derived from the element `id`. If an ID is missing, a fallback key (`node_N` or `edge_N`) is generated.

### 2. Type Updates

- **Position**: Ensured `x` and `y` coordinates in the `position` block are treated as numbers in the generated HCL, not strings.

### 3. File Changes

- Modified `internal/converter/flow_converter.go`: updated `writeNodesBlock` and `writeEdgesBlock`.
- Added `internal/converter/fix_20_test.go`: New test verifying the map structure.
- Updated `internal/converter/converter_test.go`: Fixed regressions in core unit tests.
- Updated `internal/converter/flow_comprehensive_test.go`: Fixed regressions in comprehensive tests.
- Updated `internal/converter/real_file_test.go`: Fixed regressions in integration tests.
- Updated `internal/converter/flow_elements_sensitive_test.go`: Fixed regressions in sensitive element tests.

## Verification

All tests in `internal/converter` passed successfully.

```text
pass: TestFix20_MapStructure
pass: TestAllTestdataFlowsConvertToValidHCL (validated against 34 real flow exports)
... [Full test suite passed]
```
