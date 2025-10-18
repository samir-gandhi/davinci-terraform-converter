package exporter

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Convert each application to HCL
	for _, application := range applications {
		// Convert SDK response to JSON format expected by converter
		appJSON, err := convertApplicationToJSON(&application)
		if err != nil {
			return "", fmt.Errorf("failed to convert application %s to JSON: %w", application.GetId(), err)
		}

		// Convert to HCL using existing converter
		hcl, err := converter.ConvertApplicationWithOptions(appJSON, skipDeps)
		if err != nil {
			return "", fmt.Errorf("failed to convert application %s to HCL: %w", application.GetId(), err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Combine all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), nil
}

// convertApplicationToJSON converts SDK DaVinciApplicationResponse to JSON format expected by converter
func convertApplicationToJSON(application interface{}) ([]byte, error) {
	// The SDK response structure should map directly to the converter's expected format
	return json.Marshal(application)
}
