package api

import (
	"context"
	"fmt"

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
func (c *Client) ListFlows(ctx context.Context) ([]FlowSummary, error) {
	envID, err := uuid.Parse(c.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Call the GetFlows API
	resp, httpResp, err := c.apiClient.DaVinciFlowsApi.GetFlows(ctx, envID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list flows: %w", err)
	}
	defer httpResp.Body.Close()

	// Extract flows from the embedded response
	embedded, ok := resp.GetEmbeddedOk()
	if !ok || embedded == nil {
		return []FlowSummary{}, nil
	}

	flows, ok := embedded.GetFlowsOk()
	if !ok || flows == nil {
		return []FlowSummary{}, nil
	}

	// Convert to FlowSummary
	summaries := make([]FlowSummary, 0, len(flows))
	for _, flow := range flows {
		summary := FlowSummary{
			FlowID:      flow.Id,
			Name:        flow.Name,
			Description: stringValue(flow.Description),
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetFlow retrieves detailed flow data including graph structure
func (c *Client) GetFlow(ctx context.Context, flowID string) (*FlowDetail, error) {
	envID, err := uuid.Parse(c.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Call the GetFlowById API
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

	// Convert graph data to map[string]interface{} for compatibility with converter
	if resp.GraphData != nil {
		graphMap, err := resp.GraphData.ToMap()
		if err != nil {
			return nil, fmt.Errorf("failed to convert graph data to map: %w", err)
		}
		detail.GraphData = graphMap
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
