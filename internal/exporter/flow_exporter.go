package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
)

// ExportFlows retrieves flows from the API and converts them to Terraform HCL
func ExportFlows(ctx context.Context, client *api.Client, skipDeps bool) (string, error) {
	if client == nil {
		return "", fmt.Errorf("API client is required")
	}

	// Retrieve all flows from the environment
	flowSummaries, err := client.ListFlows(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list flows: %w", err)
	}

	if len(flowSummaries) == 0 {
		return "# No flows found in environment\n", nil
	}

	var hclBlocks []string

	// Retrieve detailed flow data and convert each flow
	for _, summary := range flowSummaries {
		flowDetail, err := client.GetFlow(ctx, summary.FlowID)
		if err != nil {
			return "", fmt.Errorf("failed to get flow %s (%s): %w", summary.Name, summary.FlowID, err)
		}

		// Convert the flow detail to the format expected by the converter
		flowData, err := convertFlowDetailToMap(flowDetail)
		if err != nil {
			return "", fmt.Errorf("failed to convert flow %s to map: %w", summary.Name, err)
		}

		// Convert to HCL using the existing converter
		hcl, err := converter.ConvertFlowToHCL(flowData, client.EnvironmentID, skipDeps)
		if err != nil {
			return "", fmt.Errorf("failed to convert flow %s to HCL: %w", summary.Name, err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Combine all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), nil
}

// convertFlowDetailToMap converts FlowDetail to map[string]interface{} for the converter
func convertFlowDetailToMap(flow *api.FlowDetail) (map[string]interface{}, error) {
	// Create a flow structure compatible with the converter's expected format
	flowMap := map[string]interface{}{
		"name":        flow.Name,
		"description": flow.Description,
		"flowId":      flow.FlowID,
	}

	// Add graph data if present
	if flow.GraphData != nil {
		flowMap["graphData"] = flow.GraphData
	}

	return flowMap, nil
}

// ExportFlowsJSON exports flows in JSON format (for debugging/inspection)
func ExportFlowsJSON(ctx context.Context, client *api.Client) (string, error) {
	if client == nil {
		return "", fmt.Errorf("API client is required")
	}

	flowSummaries, err := client.ListFlows(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list flows: %w", err)
	}

	var flows []map[string]interface{}
	for _, summary := range flowSummaries {
		flowDetail, err := client.GetFlow(ctx, summary.FlowID)
		if err != nil {
			return "", fmt.Errorf("failed to get flow %s: %w", summary.FlowID, err)
		}

		flowData, err := convertFlowDetailToMap(flowDetail)
		if err != nil {
			return "", fmt.Errorf("failed to convert flow: %w", err)
		}

		flows = append(flows, flowData)
	}

	jsonData, err := json.MarshalIndent(flows, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal flows to JSON: %w", err)
	}

	return string(jsonData), nil
}
