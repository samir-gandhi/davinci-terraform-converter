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
func ExportVariables(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph) (string, []converter.VariableEligibleAttribute, error) {
	return ExportVariablesWithImports(ctx, client, skipDeps, graph, nil)
}

// ExportVariablesForModule exports variables with JSON data for module generation
// Returns HCL, extracted variables, JSON map, and resource names map
func ExportVariablesForModule(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph, importGen *importgen.ImportBlockGenerator) (string, []converter.VariableEligibleAttribute, map[string][]byte, map[string]string, error) {
	hcl, extracted, err := ExportVariablesWithImports(ctx, client, skipDeps, graph, importGen)
	if err != nil {
		return "", nil, nil, nil, err
	}

	// Re-fetch to get JSON (this is inefficient but keeps changes minimal)
	variables, err := client.ListVariables(ctx, client.EnvironmentID)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("failed to re-fetch variables for JSON: %w", err)
	}

	jsonMap := make(map[string][]byte)
	namesMap := make(map[string]string)

	for _, variable := range variables {
		variableJSON, err := convertVariableToJSON(&variable)
		if err != nil {
			return "", nil, nil, nil, fmt.Errorf("failed to convert variable to JSON: %w", err)
		}

		variableID := variable.GetId().String()
		jsonMap[variableID] = variableJSON

		// Get actual resource name from graph
		actualName, err := graph.GetReferenceName("pingone_davinci_variable", variableID)
		if err != nil {
			return "", nil, nil, nil, fmt.Errorf("failed to get resource name for variable %s: %w", variableID, err)
		}
		namesMap[variableID] = actualName
	}

	return hcl, extracted, jsonMap, namesMap, nil
}

// ExportVariablesWithImports exports all variables with optional import blocks
// Returns HCL string and extracted variable-eligible attributes for module generation
func ExportVariablesWithImports(ctx context.Context, client *api.Client, skipDeps bool, graph *resolver.DependencyGraph, importGen *importgen.ImportBlockGenerator) (string, []converter.VariableEligibleAttribute, error) {
	if client == nil {
		return "", nil, fmt.Errorf("client cannot be nil")
	}

	// Get all variables from API
	variables, err := client.ListVariables(ctx, client.EnvironmentID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to list variables: %w", err)
	}

	if len(variables) == 0 {
		return "", nil, nil
	}

	// First pass: Register all variables in the dependency graph
	for _, variable := range variables {
		variableName := variable.GetName()
		variableID := variable.GetId()
		sanitizedName := resolver.SanitizeName(variableName, nil)
		graph.AddResource("pingone_davinci_variable", variableID.String(), sanitizedName)
	}

	var hclBlocks []string
	var extractedVariables []converter.VariableEligibleAttribute

	// Second pass: Convert each variable to HCL with optional import blocks
	for _, variable := range variables {
		variableID := variable.GetId().String()

		// Get the actual resource name from the graph (includes deduplication suffix if needed)
		actualName, err := graph.GetReferenceName("pingone_davinci_variable", variableID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get resource name for variable %s: %w", variableID, err)
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
				return "", nil, fmt.Errorf("failed to generate import block for variable %s: %w", variableID, err)
			}
			hclBlocks = append(hclBlocks, importBlock)
		}

		// Convert SDK response to JSON format expected by converter
		variableJSON, err := convertVariableToJSON(&variable)
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert variable %s to JSON: %w", variable.GetId(), err)
		}

		// Extract variable-eligible attributes for module generation
		variableAttrs, err := converter.GetVariableEligibleAttributes(variableJSON, actualName)
		if err != nil {
			return "", nil, fmt.Errorf("failed to extract variable attributes for %s: %w", variable.GetId(), err)
		}
		extractedVariables = append(extractedVariables, variableAttrs...)

		// Convert to HCL using existing converter
		hcl, err := converter.ConvertVariableWithOptions(variableJSON, skipDeps)
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert variable %s to HCL: %w", variable.GetId(), err)
		}

		hclBlocks = append(hclBlocks, hcl)
	}

	// Combine all HCL blocks with blank lines between them
	return strings.Join(hclBlocks, "\n\n"), extractedVariables, nil
}

// convertVariableToJSON converts SDK DaVinciVariableResponse to JSON format expected by converter
func convertVariableToJSON(variable interface{}) ([]byte, error) {
	// The SDK response structure should map directly to the converter's VariableResponse
	// We use a generic marshal/unmarshal approach to handle the conversion
	return json.Marshal(variable)
}
