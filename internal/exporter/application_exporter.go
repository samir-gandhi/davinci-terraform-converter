package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/importgen"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

// ExportApplications exports all DaVinci applications from the API to HCL format
func ExportApplications(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph) (string, error) {
	return ExportApplicationsWithImports(ctx, client, skipDeps, graph, nil)
}

// ExportApplicationsWithImports exports applications with optional import blocks
func ExportApplicationsWithImports(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph, importGen *importgen.ImportBlockGenerator) (string, error) {
	if client == nil {
		return "", fmt.Errorf("client cannot be nil")
	}

	// Get all applications from API
	applications, err := client.ListApplications(ctx, client.EnvironmentID)
	if err != nil {
		return "", fmt.Errorf("failed to list applications: %w", err)
	}

	if len(applications) == 0 {
		return "", nil
	}

	// First pass: Register all applications in the dependency graph
	for _, application := range applications {
		appName := application.GetName()
		appID := application.GetId()
		sanitizedName := resolver.SanitizeName(appName, nil)
		graph.AddResource("pingone_davinci_application", appID, sanitizedName)
	}

	var hclBlocks []string

	// Second pass: Convert each application to HCL
	for _, application := range applications {
		appID := application.GetId()

		// Get the actual resource name from the graph (includes deduplication suffix if needed)
		actualName, err := graph.GetReferenceName("pingone_davinci_application", appID)
		if err != nil {
			return "", fmt.Errorf("failed to get resource name for application %s: %w", appID, err)
		}

		// Generate import block if import generator provided
		if importGen != nil {
			importBlock, err := importGen.GenerateImportBlock(
				"pingone_davinci_application",
				actualName,
				appID,
				client.EnvironmentID,
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate import block for application %s: %w", appID, err)
			}
			hclBlocks = append(hclBlocks, importBlock)
		}

		// Convert SDK response to JSON format expected by converter
		appJSON, err := convertApplicationToJSON(&application)
		if err != nil {
			return "", fmt.Errorf("failed to convert application %s to JSON: %w", application.GetId(), err)
		}

		// Determine environment ID based on skipDeps flag
		var environmentID string
		if skipDeps {
			environmentID = client.EnvironmentID // Will be quoted by converter
		} else {
			environmentID = "var.environment_id" // Will be written as-is by converter
		}

		// Convert to HCL using converter with environment ID and graph
		hcl, err := converter.ConvertApplicationWithEnvironmentAndGraph(appJSON, environmentID, graph)
		if err != nil {
			return "", fmt.Errorf("failed to convert application %s to HCL: %w", application.GetId(), err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Combine all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), nil
}

// ensureUniqueResourceName checks if a resource name is already used and appends a suffix if needed
func ensureUniqueResourceName(hcl string, usedNames map[string]int) string {
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

// convertApplicationToJSON converts SDK DaVinciApplicationResponse to JSON format expected by converter
func convertApplicationToJSON(application interface{}) ([]byte, error) {
	// The SDK response structure should map directly to the converter's expected format
	return json.Marshal(application)
}
