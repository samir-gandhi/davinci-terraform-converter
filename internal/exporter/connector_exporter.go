package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/importgen"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

// ExportConnectorInstances retrieves connector instances from the API and converts them to Terraform HCL
func ExportConnectorInstances(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph) (string, []converter.VariableEligibleAttribute, error) {
	return ExportConnectorInstancesWithImports(ctx, client, skipDeps, graph, nil)
}

// ExportConnectorInstancesWithImports exports connector instances with optional import blocks
// Returns HCL string and extracted variable-eligible attributes for module generation
func ExportConnectorInstancesWithImports(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph, importGen *importgen.ImportBlockGenerator) (string, []converter.VariableEligibleAttribute, error) {
	if client == nil {
		return "", nil, fmt.Errorf("API client is required")
	}

	// Retrieve all connector instances from the environment
	instanceSummaries, err := client.ListConnectorInstances(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to list connector instances: %w", err)
	}

	if len(instanceSummaries) == 0 {
		return "# No connector instances found in environment\n", nil, nil
	}

	// First pass: Register all connector instances in the dependency graph
	for _, summary := range instanceSummaries {
		sanitizedName := resolver.SanitizeName(summary.Name, nil)
		graph.AddResource("pingone_davinci_connector_instance", summary.InstanceID, sanitizedName)
	}

	var hclBlocks []string
	var extractedVariables []converter.VariableEligibleAttribute

	// Second pass: Retrieve detailed connector instance data and convert each instance
	for _, summary := range instanceSummaries {
		// Get the actual resource name from the graph (includes deduplication suffix if needed)
		actualName, err := graph.GetReferenceName("pingone_davinci_connector_instance", summary.InstanceID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get resource name for connector instance %s: %w", summary.InstanceID, err)
		}

		// Generate import block if import generator provided
		if importGen != nil {
			// Skip import for special connector IDs that don't follow UUID format
			// User Pool connector uses "defaultUserPool" which isn't a valid UUID
			if !isSpecialConnectorID(summary.InstanceID) {
				importBlock, err := importGen.GenerateImportBlock(
					"pingone_davinci_connector_instance",
					actualName,
					summary.InstanceID,
					client.EnvironmentID,
				)
				if err != nil {
					return "", nil, fmt.Errorf("failed to generate import block for connector instance %s: %w", summary.InstanceID, err)
				}
				hclBlocks = append(hclBlocks, importBlock)
			}
		}

		instanceDetail, err := client.GetConnectorInstance(ctx, summary.InstanceID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get connector instance %s (%s): %w", summary.Name, summary.InstanceID, err)
		}

		// Convert the instance detail to JSON for the converter
		instanceJSON, err := convertInstanceDetailToJSON(instanceDetail, client.EnvironmentID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert instance %s to JSON: %w", summary.Name, err)
		}

		// Extract variable-eligible attributes for module generation
		connectorAttrs, err := converter.GetConnectorInstanceVariableEligibleAttributes(instanceJSON, actualName)
		if err != nil {
			return "", nil, fmt.Errorf("failed to extract connector attributes for %s: %w", summary.Name, err)
		}
		extractedVariables = append(extractedVariables, connectorAttrs...)

		// Convert to HCL using the existing converter
		hcl, err := converter.ConvertConnectorInstanceWithOptions(instanceJSON, skipDeps)
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert connector instance %s to HCL: %w", summary.Name, err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Join all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), extractedVariables, nil
}

// isSpecialConnectorID checks if a connector instance ID is a special case that doesn't follow UUID format
func isSpecialConnectorID(instanceID string) bool {
	// User Pool connector uses "defaultUserPool" instead of UUID
	specialIDs := []string{
		"defaultUserPool",
	}

	for _, specialID := range specialIDs {
		if instanceID == specialID {
			return true
		}
	}
	return false
}

// convertInstanceDetailToJSON converts connector instance detail to JSON format expected by converter
func convertInstanceDetailToJSON(detail *api.ConnectorInstanceDetail, environmentID string) ([]byte, error) {
	// Build the structure expected by the converter
	instanceData := map[string]interface{}{
		"id":   detail.InstanceID,
		"name": detail.Name,
		"environment": map[string]interface{}{
			"id": environmentID,
		},
		"connector": map[string]interface{}{
			"id": detail.ConnectorID,
		},
	}

	// Add properties if present
	if detail.Properties != nil {
		instanceData["properties"] = detail.Properties
	}

	return json.Marshal(instanceData)
}
