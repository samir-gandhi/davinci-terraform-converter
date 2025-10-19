package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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
	usedNames := make(map[string]int) // Track resource name usage for uniqueness

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

		// Determine environment_id value based on skipDeps flag
		envID := "var.environment_id"
		if skipDeps {
			envID = client.EnvironmentID
		}

		// Convert to HCL using the existing converter
		hcl, err := converter.ConvertFlowToHCL(flowData, envID, skipDeps)
		if err != nil {
			return "", fmt.Errorf("failed to convert flow %s to HCL: %w", summary.Name, err)
		}

		// Ensure unique resource names by appending suffix if name already used
		hcl = ensureUniqueFlowResourceName(hcl, usedNames)

		hclBlocks = append(hclBlocks, hcl)
	}

	// Combine all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), nil
}

// ensureUniqueFlowResourceName checks if a resource name is already used and appends a suffix if needed
func ensureUniqueFlowResourceName(hcl string, usedNames map[string]int) string {
	// Extract resource name from HCL (format: resource "type" "name" {)
	re := regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(hcl)

	if len(matches) < 3 {
		return hcl // No resource declaration found, return as-is
	}

	resourceType := matches[1]
	originalName := matches[2]
	key := resourceType + "." + originalName

	// Check if name has been used
	if count, exists := usedNames[key]; exists {
		// Name already used, append counter suffix
		usedNames[key] = count + 1
		newName := fmt.Sprintf("%s_%d", originalName, count+1)
		newKey := resourceType + "." + newName
		usedNames[newKey] = 0

		// Replace the resource name in HCL
		hcl = re.ReplaceAllString(hcl, fmt.Sprintf(`resource "%s" "%s"`, resourceType, newName))
	} else {
		// First use of this name
		usedNames[key] = 0
	}

	return hcl
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
