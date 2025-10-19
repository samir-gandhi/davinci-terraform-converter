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

// ExportApplications exports all DaVinci applications from the API to HCL format
func ExportApplications(ctx context.Context, client *api.Client, skipDeps bool) (string, error) {
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

	var hclBlocks []string
	usedNames := make(map[string]int) // Track resource name usage for uniqueness

	// Convert each application to HCL
	for _, application := range applications {
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

		// Convert to HCL using converter with environment ID
		hcl, err := converter.ConvertApplicationWithEnvironment(appJSON, environmentID)
		if err != nil {
			return "", fmt.Errorf("failed to convert application %s to HCL: %w", application.GetId(), err)
		}

		// Ensure unique resource names by appending suffix if name already used
		hcl = ensureUniqueResourceName(hcl, usedNames)

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
