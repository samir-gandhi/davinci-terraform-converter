package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// FlowSummary represents a summary of a DaVinci flow from the API
type FlowSummary struct {
	FlowID      string
	Name        string
	Description string
	// Add other relevant fields as needed
}

// FlowDetail represents detailed flow data including graph structure
type FlowDetail struct {
	FlowID      string
	Name        string
	Description string
	GraphData   map[string]interface{} // The full flow graph structure
	// Add other relevant fields as needed
}

// ListFlows retrieves all flows from the environment
//
// WORKAROUND: Uses raw HTTP request to bypass SDK's strict validation of optional fields.
// The SDK requires the Version field in flow responses, but the API returns flows where
// this field may be absent. This causes unmarshaling to fail.
//
// TODO: Revert to SDK's GetFlows() once the Version field is fixed in the SDK.
// See WORKAROUND_RAW_HTTP.md for detailed reversion instructions.
//
// Related SDK Issue: Version field in DaVinciFlowResponse lacks omitempty tag.
func (c *Client) ListFlows(ctx context.Context) ([]FlowSummary, error) {
	envID, err := uuid.Parse(c.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Get an access token using the SDK's token source
	tokenSource, err := c.serviceCfg.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %w", err)
	}

	token, err := (*tokenSource).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("token source returned empty access token")
	}

	// Make raw HTTP request
	// Use the correct path structure matching the SDK
	baseURL := fmt.Sprintf("https://api.pingone.%s/v1/environments/%s/flows",
		getRegionDomain(c.Region), envID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	// Create HTTP client
	httpClient := http.DefaultClient

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", httpResp.StatusCode, string(body))
	}

	// Parse response as raw JSON
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract flows from _embedded structure
	summaries := []FlowSummary{}

	if embedded, ok := rawResponse["_embedded"].(map[string]interface{}); ok {
		if flowsData, ok := embedded["flows"].([]interface{}); ok {
			for _, flowItem := range flowsData {
				if flowMap, ok := flowItem.(map[string]interface{}); ok {
					summary := FlowSummary{}

					if id, ok := flowMap["id"].(string); ok {
						summary.FlowID = id
					}
					if name, ok := flowMap["name"].(string); ok {
						summary.Name = name
					}
					if desc, ok := flowMap["description"].(string); ok {
						summary.Description = desc
					}

					summaries = append(summaries, summary)
				}
			}
		}
	}

	return summaries, nil
}

// GetFlow retrieves detailed flow data including graph structure
//
// WORKAROUND: Uses raw HTTP request to bypass SDK's strict validation of optional fields.
// The SDK requires the Position field in flow graph nodes/edges, but the API returns flows
// where these fields may be absent. This causes unmarshaling to fail.
//
// TODO: Revert to SDK's GetFlowById() once the Position field is fixed in the SDK.
// See WORKAROUND_RAW_HTTP.md for detailed reversion instructions.
//
// Related SDK Issue: Position field in DaVinciFlowGraphDataResponseElementsNode and
// DaVinciFlowGraphDataResponseElementsEdge lacks omitempty tag and is not a pointer.
func (c *Client) GetFlow(ctx context.Context, flowID string) (*FlowDetail, error) {
	envID, err := uuid.Parse(c.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Get an access token using the SDK's token source
	tokenSource, err := c.serviceCfg.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %w", err)
	}

	token, err := (*tokenSource).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("token source returned empty access token")
	}

	// Make raw HTTP request
	// Use the correct path structure matching the SDK
	baseURL := fmt.Sprintf("https://api.pingone.%s/v1/environments/%s/flows/%s",
		getRegionDomain(c.Region), envID.String(), flowID)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	// Create HTTP client
	httpClient := http.DefaultClient

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", httpResp.StatusCode, string(body))
	}

	// Parse response as raw JSON
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract flow details
	detail := &FlowDetail{
		FlowID: flowID,
	}

	if name, ok := rawResponse["name"].(string); ok {
		detail.Name = name
	}

	if desc, ok := rawResponse["description"].(string); ok {
		detail.Description = desc
	}

	if graphData, ok := rawResponse["graphData"].(map[string]interface{}); ok {
		detail.GraphData = graphData
	}

	return detail, nil
}

// stringValue safely extracts string value from pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
