# Raw HTTP Workaround for SDK Position Field Issue

**Date Implemented:** October 12, 2025  
**Status:** TEMPORARY WORKAROUND  
**Related Issue:** See `SDK_POSITION_FIELD_ISSUE.md`

## Overview

The `GetFlow()` method in `internal/api/flows.go` uses raw HTTP requests instead of the SDK's `GetFlowById()` method to bypass strict validation of the optional `Position` field in flow graph data structures.

## Affected Code

**File:** `internal/api/flows.go`  
**Method:** `func (c *Client) GetFlow(ctx context.Context, flowID string) (*FlowDetail, error)`

## Current Implementation

```go
// GetFlow retrieves detailed flow data including graph structure
// Uses raw HTTP request to bypass SDK's strict validation of optional fields
func (c *Client) GetFlow(ctx context.Context, flowID string) (*FlowDetail, error) {
	// ... validation code ...

	// Get an access token using the SDK's token source
	tokenSource, err := c.serviceCfg.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %w", err)
	}

	token, err := (*tokenSource).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Manual HTTP request construction
	baseURL := fmt.Sprintf("https://api.pingone.%s/v1/environments/%s/flows/%s",
		getRegionDomain(c.Region), envID.String(), flowID)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	// ... rest of manual HTTP handling ...
}
```

## How to Revert Once SDK is Fixed

### Step 1: Verify SDK Fix

Check that the SDK has been updated with the Position field fixes:

```bash
# In the pingone-go-client repository
cd /Users/samirgandhi/go/src/github.com/pingidentity/pingone-go-client
git log --oneline --grep="position" --grep="DaVinci" --all
```

Look for commits that:
- Make `Position` field optional (pointer with `omitempty`)
- Update helper methods (`GetPosition`, `GetPositionOk`, `SetPosition`, `New*`)
- Update `UnmarshalJSON` to not require Position

### Step 2: Update SDK Dependency

```bash
cd /Users/samirgandhi/go/src/github.com/samir-gandhi/davinci-terraform-converter

# If using local replace directive, remove it from go.mod:
# Delete or comment out the replace directive

# Update to the fixed SDK version:
go get github.com/pingidentity/pingone-go-client@v0.x.x  # Use actual fixed version
go mod tidy
```

### Step 3: Revert GetFlow() Implementation

Replace the current raw HTTP implementation with the SDK method:

```go
// GetFlow retrieves detailed flow data including graph structure
func (c *Client) GetFlow(ctx context.Context, flowID string) (*FlowDetail, error) {
	envID, err := uuid.Parse(c.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Call the GetFlowById API using SDK
	resp, httpResp, err := c.apiClient.DaVinciFlowsApi.GetFlowById(ctx, envID, flowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get flow %s: %w", flowID, err)
	}
	defer httpResp.Body.Close()

	// Extract flow details
	detail := &FlowDetail{
		FlowID:      resp.Id,
		Name:        resp.Name,
		Description: stringValue(resp.Description),
	}

	// Convert graph data using SDK's ToMap() method (now fixed)
	if resp.GraphData != nil {
		graphMap, err := resp.GraphData.ToMap()
		if err != nil {
			return nil, fmt.Errorf("failed to convert graph data: %w", err)
		}
		detail.GraphData = graphMap
	}

	return detail, nil
}
```

### Step 4: Remove Workaround Dependencies

Update imports in `internal/api/flows.go`:

```go
// Remove these if no longer needed:
import (
	"io"           // May no longer be needed
	"net/http"     // May no longer be needed
)
```

### Step 5: Remove serviceCfg from Client

If `serviceCfg` was only added for token access in the workaround:

**File:** `internal/api/client.go`

```go
// Remove serviceCfg field from Client struct:
type Client struct {
	apiClient         *pingone.APIClient
	// serviceCfg        *config.Configuration  // DELETE THIS LINE
	AuthEnvironmentID string
	EnvironmentID     string
	Region            string
}

// Remove from NewClient initialization:
client := &Client{
	apiClient:         apiClient,
	// serviceCfg:        serviceCfg,  // DELETE THIS LINE
	AuthEnvironmentID: authEnvironmentID,
	EnvironmentID:     targetEnvironmentID,
	Region:            region,
}
```

### Step 6: Run Tests

Verify that the SDK-based implementation works:

```bash
# Run unit tests
go test ./internal/api -v

# Run acceptance tests (requires credentials)
go test -tags=acceptance ./tests/acceptance -v

# Specifically test flow retrieval
go test -tags=acceptance ./tests/acceptance -run TestGetSingleFlowFromAPI -v
go test -tags=acceptance ./tests/acceptance -run TestExportFlowsFromAPI -v
```

### Step 7: Cleanup Documentation

After successful reversion:
1. Delete this file (`WORKAROUND_RAW_HTTP.md`)
2. Archive `SDK_POSITION_FIELD_ISSUE.md` to a `docs/historical/` folder
3. Update any comments in code that reference the workaround

## Testing Checklist

After reverting to SDK-based implementation, verify:

- [ ] All unit tests pass
- [ ] All acceptance tests pass
- [ ] Flow retrieval works for flows with nodes/edges without position data
- [ ] Flow retrieval works for flows with position data
- [ ] Flow export to HCL works correctly
- [ ] No compilation errors or warnings
- [ ] Performance is equivalent or better than workaround

## Additional Notes

### Why This Workaround Was Necessary

The SDK's `UnmarshalJSON` methods for `DaVinciFlowGraphDataResponseElementsNode` and `DaVinciFlowGraphDataResponseElementsEdge` required the `Position` field to be present, but the PingOne API returns flows where nodes and edges may not have position data. This caused unmarshaling to fail during `Execute()`.

### Trade-offs of the Workaround

**Pros:**
- Allows flow retrieval to work immediately
- Maintains type safety for other parts of the response
- Uses SDK's authentication mechanism

**Cons:**
- Manual HTTP request handling
- No type safety for the raw JSON response
- Must manually parse response into map[string]interface{}
- Bypasses SDK's error handling and retries
- Duplicates URL construction logic

### Performance Impact

The workaround should have minimal performance impact:
- Same number of network calls
- Similar JSON unmarshaling overhead
- No additional allocations for intermediate structures

## Contact

If you encounter issues reverting this workaround, contact:
- SDK Team: File an issue at github.com/pingidentity/pingone-go-client
- Reference the original issue document: `SDK_POSITION_FIELD_ISSUE.md`
