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

// ExportVariables exports all variables from the API to HCL format
func ExportVariables(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph) (string, error) {
	return ExportVariablesWithImports(ctx, client, skipDeps, graph, nil)
}

// ExportVariablesWithImports exports all variables with optional import blocks
func ExportVariablesWithImports(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph, importGen *importgen.ImportBlockGenerator) (string, error) {
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

	// First pass: Register all variables in the dependency graph
	for _, variable := range variables {
		variableName := variable.GetName()
		variableID := variable.GetId()
		sanitizedName := resolver.SanitizeName(variableName, nil)
		graph.AddResource("pingone_davinci_variable", variableID.String(), sanitizedName)
	}

	var hclBlocks []string

	// Second pass: Convert each variable to HCL with optional import blocks
	for _, variable := range variables {
		variableID := variable.GetId().String()

		// Get the actual resource name from the graph (includes deduplication suffix if needed)
		actualName, err := graph.GetReferenceName("pingone_davinci_variable", variableID)
		if err != nil {
			return "", fmt.Errorf("failed to get resource name for variable %s: %w", variableID, err)
		}

		// Generate import block if import generator provided
		if importGen != nil {
			importBlock, err := importGen.GenerateImportBlock(
				"pingone_davinci_variable",
				actualName,
				variableID,
				client.EnvironmentID,
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate import block for variable %s: %w", variableID, err)
			}
			hclBlocks = append(hclBlocks, importBlock)
		}

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
