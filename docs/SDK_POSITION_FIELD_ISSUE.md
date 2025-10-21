# PingOne Go SDK - DaVinci Flow Position Field Validation Issue

**Date:** October 12, 2025  
**SDK Version:** v0.2.0  
**SDK Branch:** DvFlowsFixes  
**Reporter:** Samir Gandhi

## Issue Summary

The PingOne Go SDK enforces strict validation on the `Position` field in DaVinci flow graph data structures, requiring it to be present even though the PingOne API returns flows where nodes and edges may not have position data. This causes API calls to fail during response parsing.

## Error Message

```
Error: no value given for required property position
```

This error occurs when calling `DaVinciFlowsApi.GetFlowById().Execute()` for flows that contain nodes or edges without position information.

## Affected SDK Models

### 1. `model_da_vinci_flow_graph_data_response_elements_node.go`

**Current Definition (Line ~45):**
```go
type DaVinciFlowGraphDataResponseElementsNode struct {
    // ... other fields ...
    Position DaVinciFlowGraphDataResponseElementsNodePosition `json:"position"`
    // ... other fields ...
}
```

**Issue:** The `Position` field lacks the `omitempty` JSON tag and is not a pointer, making it required during unmarshaling.

### 2. `model_da_vinci_flow_graph_data_response_elements_edge.go`

**Current Definition (Line ~45):**
```go
type DaVinciFlowGraphDataResponseElementsEdge struct {
    // ... other fields ...
    Position DaVinciFlowGraphDataResponseElementsEdgePosition `json:"position"`
    // ... other fields ...
}
```

**Issue:** Same as above - `Position` field is required but API may not return it.

## Root Cause

The SDK's `UnmarshalJSON` methods in both model files validate that all required fields are present. Since `Position` lacks the `omitempty` tag, the SDK treats it as mandatory. When the PingOne API returns flow data without position information (which is valid), unmarshaling fails.

## API Behavior

The PingOne DaVinci Flows API (`GET /v1/environments/{environmentId}/davinci/flows/{flowId}`) returns flow objects where:
- Some nodes may not have `position` data
- Some edges may not have `position` data
- This is valid API behavior - position is optional

## Attempted Solutions

### Solution 1: Make Position Optional with Pointer + omitempty
```go
Position *DaVinciFlowGraphDataResponseElementsNodePosition `json:"position,omitempty"`
```

**Result:** Compilation errors in helper methods at lines 54, 228, 237, 242:
- `NewDaVinciFlowGraphDataResponseElementsNode()` - Cannot use struct type as pointer type
- `GetPosition()` - Return type mismatch
- `GetPositionOk()` - Return type mismatch  
- `SetPosition()` - Parameter type mismatch

All helper methods expect the non-pointer type, causing cascading compilation failures.

### Solution 2: Marshal/Unmarshal Workaround
```go
graphBytes, _ := json.Marshal(resp.GraphData)
json.Unmarshal(graphBytes, &graphMap)
```

**Result:** Failed - validation occurs during `Execute()` before we can access the data.

### Solution 3: Raw HTTP Request
Bypassing the SDK entirely with manual HTTP requests works but loses SDK authentication handling and type safety.

## Impact

This issue blocks:
- Reading DaVinci flows with nodes/edges that lack position data
- Integration testing against real PingOne environments
- Flow export/import functionality
- Any automation tools that need to retrieve flow configurations

## Recommended Fix

Make the `Position` field optional in both models AND update all helper methods:

### 1. Update struct definitions:
```go
// In model_da_vinci_flow_graph_data_response_elements_node.go
Position *DaVinciFlowGraphDataResponseElementsNodePosition `json:"position,omitempty"`

// In model_da_vinci_flow_graph_data_response_elements_edge.go
Position *DaVinciFlowGraphDataResponseElementsEdgePosition `json:"position,omitempty"`
```

### 2. Update helper methods in both files:

**Line ~54 - New function:**
```go
// Before:
Position: DaVinciFlowGraphDataResponseElementsNodePosition{},

// After:
Position: nil,
```

**Line ~228 - GetPosition:**
```go
// Before:
func (o *DaVinciFlowGraphDataResponseElementsNode) GetPosition() DaVinciFlowGraphDataResponseElementsNodePosition {
    if o == nil {
        var ret DaVinciFlowGraphDataResponseElementsNodePosition
        return ret
    }
    return o.Position
}

// After:
func (o *DaVinciFlowGraphDataResponseElementsNode) GetPosition() *DaVinciFlowGraphDataResponseElementsNodePosition {
    if o == nil {
        return nil
    }
    return o.Position
}
```

**Line ~237 - GetPositionOk:**
```go
// Before:
func (o *DaVinciFlowGraphDataResponseElementsNode) GetPositionOk() (*DaVinciFlowGraphDataResponseElementsNodePosition, bool) {
    if o == nil {
        return nil, false
    }
    return &o.Position, true
}

// After:
func (o *DaVinciFlowGraphDataResponseElementsNode) GetPositionOk() (*DaVinciFlowGraphDataResponseElementsNodePosition, bool) {
    if o == nil || o.Position == nil {
        return nil, false
    }
    return o.Position, true
}
```

**Line ~242 - SetPosition:**
```go
// Before:
func (o *DaVinciFlowGraphDataResponseElementsNode) SetPosition(v DaVinciFlowGraphDataResponseElementsNodePosition) {
    o.Position = v
}

// After:
func (o *DaVinciFlowGraphDataResponseElementsNode) SetPosition(v *DaVinciFlowGraphDataResponseElementsNodePosition) {
    o.Position = v
}
```

### 3. Update UnmarshalJSON validation

Remove `Position` from the required fields list in the `UnmarshalJSON` method.

## Validation

After applying these changes:
1. Flows without position data should unmarshal successfully
2. Flows with position data should continue working as before
3. All helper methods should compile without errors
4. Existing tests should pass

## Additional Context

- The Terraform Provider (`terraform-provider-pingone`) uses the same SDK and may have worked around this issue or is using a version with the fix
- The SDK's `DvFlowsFixes` branch contains other recent fixes but still has this position field issue
- This appears to be an oversight where the API schema allows optional position but the SDK generated code doesn't reflect that

## Environment

- **Go Version:** 1.23+
- **SDK Import Path:** `github.com/pingidentity/pingone-go-client`
- **API Endpoint:** `GET /v1/environments/{envId}/davinci/flows/{flowId}`
- **Affected API:** `DaVinciFlowsApi.GetFlowById()`

## Reproducibility

1. Set up PingOne environment with DaVinci flows
2. Configure SDK client with proper authentication
3. Call `GetFlowById()` on a flow containing nodes/edges without position data
4. Observe validation error during `Execute()`

## Workaround

**Status:** IMPLEMENTED (October 12, 2025)

Currently using raw HTTP requests with manual authentication header management to bypass SDK validation until this is fixed.

**Implementation Details:**
- **File:** `internal/api/flows.go`
- **Method:** `GetFlow()` uses raw HTTP request instead of SDK's `GetFlowById()`
- **Authentication:** Uses SDK's `TokenSource` to get OAuth2 token
- **Response Parsing:** Manually unmarshals JSON into `map[string]interface{}`
- **Reversion Guide:** See `WORKAROUND_RAW_HTTP.md` for detailed instructions on reverting once SDK is fixed

**Code Changes:**
1. Added `serviceCfg *config.Configuration` field to `Client` struct (for token access)
2. Modified `GetFlow()` to construct manual HTTP request to `/v1/environments/{id}/flows/{id}`
3. Uses `http.DefaultClient` for request execution
4. Parses response as raw JSON to bypass SDK validation

**Test Status:**
- All acceptance tests passing (11/11)
- Flow export working correctly
- Position data handled gracefully when present or absent
