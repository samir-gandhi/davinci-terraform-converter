package exporter

import (
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
)

// RegenerateHCLWithVariableReferences regenerates resource HCL with variable references
// This is called after all variables are extracted to replace hardcoded values with var.{name}
func RegenerateHCLWithVariableReferences(data *ExportedData, skipDependencies bool) error {
	// Build variable maps for each resource type
	variableVariableMap := buildVariableVariableMap(data.ExtractedVariables)
	connectorVariableMap := buildConnectorVariableMap(data.ExtractedVariables)

	// Regenerate Variables HCL with variable references
	if data.VariablesHCL != "" && len(variableVariableMap) > 0 {
		// Parse each variable resource and regenerate
		// For now, we'll need to store the original JSON data to regenerate
		// This is a TODO for future enhancement
	}

	// Regenerate Connectors HCL with variable references
	if data.ConnectorsHCL != "" && len(connectorVariableMap) > 0 {
		// Parse each connector resource and regenerate
		// For now, we'll need to store the original JSON data to regenerate
		// This is a TODO for future enhancement
	}

	return nil
}

// buildVariableVariableMap creates a map of variable_name -> variable_name for DaVinci variables
func buildVariableVariableMap(extracted []converter.VariableEligibleAttribute) map[string]string {
	varMap := make(map[string]string)
	for _, attr := range extracted {
		if attr.ResourceType == "variable" {
			// For variables, the key is the resource name
			varMap[attr.ResourceName] = attr.VariableName
		}
	}
	return varMap
}

// buildConnectorVariableMap creates a map of properties.{prop} -> variable_name for connectors
func buildConnectorVariableMap(extracted []converter.VariableEligibleAttribute) map[string]map[string]string {
	// Map of resource_name -> map of property_path -> variable_name
	connMap := make(map[string]map[string]string)
	for _, attr := range extracted {
		if attr.ResourceType == "connection" {
			if connMap[attr.ResourceName] == nil {
				connMap[attr.ResourceName] = make(map[string]string)
			}
			connMap[attr.ResourceName][attr.AttributePath] = attr.VariableName
		}
	}
	return connMap
}

// parseResourceBlocks parses HCL string into individual resource blocks
func parseResourceBlocks(hcl string) ([]resourceBlock, error) {
	// Simple parsing - split on "resource" keyword
	blocks := []resourceBlock{}

	lines := strings.Split(hcl, "\n")
	var currentBlock *resourceBlock
	var blockContent strings.Builder
	braceCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of new resource block
		if strings.HasPrefix(trimmed, "resource \"") {
			// Save previous block if exists
			if currentBlock != nil {
				currentBlock.HCL = blockContent.String()
				blocks = append(blocks, *currentBlock)
			}

			// Parse resource type and name
			parts := strings.Split(trimmed, "\"")
			if len(parts) >= 4 {
				currentBlock = &resourceBlock{
					ResourceType: parts[1],
					ResourceName: parts[3],
				}
				blockContent.Reset()
				braceCount = 0
			}
		}

		if currentBlock != nil {
			blockContent.WriteString(line + "\n")

			// Track braces to know when block ends
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			if braceCount == 0 && strings.Contains(line, "}") {
				// Block complete
				currentBlock.HCL = blockContent.String()
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				blockContent.Reset()
			}
		}
	}

	return blocks, nil
}

type resourceBlock struct {
	ResourceType string
	ResourceName string
	HCL          string
	JSONData     []byte // Original JSON data if available
}

// extractJSONFromComment tries to extract JSON data from special comments in HCL
// Format: # __JSON_DATA__: {...}
func extractJSONFromComment(hcl string) ([]byte, error) {
	marker := "# __JSON_DATA__:"
	if idx := strings.Index(hcl, marker); idx >= 0 {
		start := idx + len(marker)
		end := strings.Index(hcl[start:], "\n")
		if end == -1 {
			end = len(hcl)
		} else {
			end += start
		}
		jsonStr := strings.TrimSpace(hcl[start:end])
		return []byte(jsonStr), nil
	}
	return nil, fmt.Errorf("no JSON data found in HCL comment")
}

// regenerateResourceWithVars regenerates a single resource with variable references
func regenerateResourceWithVars(block resourceBlock, variableMap map[string]string, skipDeps bool) (string, error) {
	switch block.ResourceType {
	case "pingone_davinci_variable":
		// Regenerate variable HCL with var reference
		if block.JSONData != nil {
			return converter.GenerateVariableHCLWithVariableReferences(block.JSONData, skipDeps, variableMap[block.ResourceName])
		}
	case "pingone_davinci_connector_instance":
		// Regenerate connector HCL with property var references
		if block.JSONData != nil {
			return converter.GenerateConnectorInstanceHCLWithVariableReferences(block.JSONData, skipDeps, variableMap)
		}
	}

	// If we can't regenerate, return original
	return block.HCL, nil
}

// ConvertJSONToVariableMap converts a JSON map to map[string]string for variable references
func ConvertJSONToVariableMap(jsonData map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, value := range jsonData {
		if str, ok := value.(string); ok {
			result[key] = str
		}
	}
	return result
}
