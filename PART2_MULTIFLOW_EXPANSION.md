# Part 2 Multi-Flow Expansion - COMPLETE

## ✅ Multi-Flow Export Support Added

### What Was Implemented

Extended the converter to handle DaVinci multi-flow exports where a parent flow and its subflows are exported together in a single JSON file with a `"flows"` array wrapper.

### New Structures

**MultiFlowExport struct** - Added to `converter.go`:
```go
type MultiFlowExport struct {
    Flows      []FlowExport `json:"flows"`
    CompanyID  string       `json:"companyId,omitempty"`
    CustomerID string       `json:"customerId,omitempty"`
}
```

### New Functions

**ConvertMultiFlow()** - Added to `converter.go`:
- Takes JSON with `"flows"` array containing multiple flow exports
- Returns slice of HCL strings (one per flow)
- Handles empty arrays gracefully
- Provides detailed error messages with flow names/indices

**Function Signature:**
```go
func ConvertMultiFlow(multiFlowJSON []byte) ([]string, error)
```

### Multi-Flow Export Format

DaVinci can export flows in two formats:

**1. Single Flow Export** (existing):
```json
{
  "name": "My Flow",
  "flowId": "abc123",
  "graphData": { ... }
}
```

**2. Multi-Flow Export** (NEW):
```json
{
  "flows": [
    {
      "name": "Parent Flow",
      "flowId": "parent-id",
      "graphData": { ... }
    },
    {
      "name": "Subflow One",
      "flowId": "subflow-one-id",
      "parentFlowId": "parent-id",
      "graphData": { ... }
    }
  ],
  "companyId": "company-123",
  "customerId": "customer-456"
}
```

### Test Coverage (4 New Tests)

#### 1. TestMultiFlowExport
- Tests conversion of 3 flows (1 parent + 2 subflows)
- Verifies each flow generates correct HCL resource
- Validates all flow attributes preserved
- Checks nodes, edges, settings, variables handling

**Test Scenario:**
- Main Flow: With nodes, edges, and settings
- Subflow One: With EVAL nodes
- Subflow Two: With variables

#### 2. TestSingleFlowWrappedInFlowsArray
- Tests backwards compatibility
- Single flow wrapped in `"flows"` array still works
- Ensures existing exports with wrapper don't break

#### 3. TestEmptyFlowsArray
- Edge case: Empty flows array
- Returns empty slice without error
- Graceful handling of no flows

#### 4. TestRealMultiFlowFile
- Integration test with actual DaVinci export
- File: `PingOne_Sign On with Sessions_multiflow.json`
- Contains 2 real production flows:
  - Flow 1: "PingOne Sign On with Sessions" (31KB)
  - Flow 2: "PingOne Sign On with Registration, Password Reset and Recovery" (359KB)
- Validates complete conversion of complex real-world flows
- Verifies all major components present

### Output Example

**Input:** Multi-flow export with 2 flows

**Output:** Array of 2 HCL strings:

**Flow 1:**
```hcl
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id

  name        = "Main Flow"
  description = "Parent flow"

  graph_data {
    elements {
      nodes = [
        {
          "data": {
            "connectionId": "conn-123",
            "connectorId": "httpConnector",
            "id": "node1",
            "nodeType": "CONNECTION"
          }
        },
      ]
    }
  }

  settings {
    {
      "logLevel": 4
    }
  }
}
```

**Flow 2:**
```hcl
resource "pingone_davinci_flow" "subflow_one" {
  environment_id = var.environment_id

  name        = "Subflow One"
  description = "First subflow"

  graph_data {
    elements {
      nodes = [
        {
          "data": {
            "id": "node2",
            "nodeType": "EVAL"
          }
        },
      ]
    }
  }
}
```

### Usage Pattern

```go
// For multi-flow exports (flows array)
results, err := converter.ConvertMultiFlow(multiFlowJSON)
if err != nil {
    // Handle error
}
for i, hcl := range results {
    // Write each flow to separate file or combine
    fmt.Printf("Flow %d:\n%s\n\n", i+1, hcl)
}

// For single flow exports (direct flow object)
hcl, err := converter.Convert(singleFlowJSON)
if err != nil {
    // Handle error
}
fmt.Println(hcl)
```

### Key Features

✅ **Multiple Flows**: Converts all flows in export (parent + subflows)
✅ **Separate Resources**: Each flow becomes distinct `pingone_davinci_flow` resource
✅ **Full Attributes**: All flow features preserved (nodes, edges, settings, variables)
✅ **Error Context**: Error messages include flow name and index for debugging
✅ **Empty Handling**: Gracefully handles empty flows array
✅ **Real-World Tested**: Validated against actual DaVinci production exports

### Files Modified

```
internal/converter/
├── converter.go          # Added MultiFlowExport struct and ConvertMultiFlow()
├── converter_test.go     # Added 3 unit tests for multi-flow scenarios
└── real_file_test.go     # NEW: Integration test with real export file
```

### Test Results

```bash
$ go test ./internal/converter/... -v 2>&1 | grep -c "^=== RUN"
25

$ go test ./internal/converter/... -cover
coverage: 91.6% of statements

$ make all
✓ All tests pass
✓ Binary builds successfully
```

### Integration Notes

**For Part 4 (Command Integration):**
- Need to detect if input has `"flows"` array
- If yes: Call `ConvertMultiFlow()`, write multiple files or concatenate
- If no: Call `Convert()`, write single file
- Consider flags like `--output-dir` for multi-flow exports

**Detection Logic:**
```go
var check map[string]interface{}
json.Unmarshal(fileData, &check)

if _, hasFlows := check["flows"]; hasFlows {
    // Multi-flow export - use ConvertMultiFlow()
} else {
    // Single flow export - use Convert()
}
```

### Real-World Validation

Tested with production DaVinci export containing:
- 15+ nodes per flow
- 120+ edges
- Complex nested properties
- Settings with CSP, CSS, HTML
- Variables with multiple contexts
- All connector types (CONNECTION, EVAL, ANNOTATION, etc.)

**Result:** Complete successful conversion with all data preserved.

## Summary

Multi-flow export support complete. The converter now handles both single-flow and multi-flow DaVinci exports. All 25 tests passing with 91.6% coverage. Ready for Part 3 (environment-specific dependency handling) or Part 4 (command integration).

### Test Count Evolution
- Part 2 Initial: 6 tests
- Part 2 Expanded: 21 tests (settings/variables/errors)
- Part 2 Multi-Flow: **25 tests** (multi-flow support + real file validation)
