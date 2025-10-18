package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
)

// ExportVariables exports all variables from the API to HCL format
func ExportVariables(ctx context.Context, client *api.Client, skipDeps bool) (string, error) {
	if client == nil {
		return "", fmt.Errorf("client cannot be nil")
	}

	// Get all variables from API
	variables, err := client.ListVariables(ctx, client.EnvironmentID)
	if err != nil {
		return "", fmt.Errorf("failed to list variables: %w", err)
	}

	if len(variables) == 0 {
		return "", nil
	}

	var hclBlocks []string

	// Convert each variable to HCL
	for _, variable := range variables {
		// Convert SDK response to JSON format expected by converter
		variableJSON, err := convertVariableToJSON(&variable)
		if err != nil {
			return "", fmt.Errorf("failed to convert variable %s to JSON: %w", variable.GetId(), err)
		}

		// Convert to HCL using existing converter
		hcl, err := converter.ConvertVariableWithOptions(variableJSON, skipDeps)
		if err != nil {
			return "", fmt.Errorf("failed to convert variable %s to HCL: %w", variable.GetId(), err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Combine all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), nil
}

// convertVariableToJSON converts SDK DaVinciVariableResponse to JSON format expected by converter
func convertVariableToJSON(variable interface{}) ([]byte, error) {
	// The SDK response structure should map directly to the converter's VariableResponse
	// We use a generic marshal/unmarshal approach to handle the conversion
	return json.Marshal(variable)
}
