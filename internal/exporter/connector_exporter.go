package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

// ExportConnectorInstances retrieves connector instances from the API and converts them to Terraform HCL
func ExportConnectorInstances(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph) (string, error) {
	if client == nil {
		return "", fmt.Errorf("API client is required")
	}

	// Retrieve all connector instances from the environment
	instanceSummaries, err := client.ListConnectorInstances(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list connector instances: %w", err)
	}

	if len(instanceSummaries) == 0 {
		return "# No connector instances found in environment\n", nil
	}

	// First pass: Register all connector instances in the dependency graph
	for _, summary := range instanceSummaries {
		sanitizedName := resolver.SanitizeName(summary.Name, nil)
		graph.AddResource("pingone_davinci_connector_instance", summary.InstanceID, sanitizedName)
	}

	var hclBlocks []string

	// Second pass: Retrieve detailed connector instance data and convert each instance
	for _, summary := range instanceSummaries {
		instanceDetail, err := client.GetConnectorInstance(ctx, summary.InstanceID)
		if err != nil {
			return "", fmt.Errorf("failed to get connector instance %s (%s): %w", summary.Name, summary.InstanceID, err)
		}

		// Convert the instance detail to JSON for the converter
		instanceJSON, err := convertInstanceDetailToJSON(instanceDetail, client.EnvironmentID)
		if err != nil {
			return "", fmt.Errorf("failed to convert instance %s to JSON: %w", summary.Name, err)
		}

		// Convert to HCL using the existing converter
		hcl, err := converter.ConvertConnectorInstanceWithOptions(instanceJSON, skipDeps)
		if err != nil {
			return "", fmt.Errorf("failed to convert connector instance %s to HCL: %w", summary.Name, err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Join all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), nil
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
