package converter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

// ConnectorInstanceResponse represents the DaVinci API response for a connector instance
type ConnectorInstanceResponse struct {
	ID          string `json:"id"`
	Environment struct {
		ID string `json:"id"`
	} `json:"environment"`
	Connector struct {
		ID string `json:"id"`
	} `json:"connector"`
	Name       string                            `json:"name"`
	Properties map[string]ConnectorPropertyValue `json:"properties,omitempty"`
}

// ConnectorPropertyValue represents a connector property with type and value
type ConnectorPropertyValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ConvertConnectorInstance converts a DaVinci connector instance JSON to Terraform HCL
func ConvertConnectorInstance(instanceJSON []byte) (string, error) {
	return ConvertConnectorInstanceWithOptions(instanceJSON, false)
}

// ConvertConnectorInstanceWithOptions converts a connector instance with optional skip-dependencies flag
func ConvertConnectorInstanceWithOptions(instanceJSON []byte, skipDependencies bool) (string, error) {
	var instance ConnectorInstanceResponse
	if err := json.Unmarshal(instanceJSON, &instance); err != nil {
		return "", fmt.Errorf("failed to parse connector instance JSON: %w", err)
	}

	if instance.Name == "" {
		return "", fmt.Errorf("connector instance name is required")
	}

	if instance.Connector.ID == "" {
		return "", fmt.Errorf("connector.id is required")
	}

	return generateConnectorInstanceHCL(instance, skipDependencies), nil
}

// generateConnectorInstanceHCL generates the Terraform HCL for a connector instance
func generateConnectorInstanceHCL(instance ConnectorInstanceResponse, skipDependencies bool) string {
	var hcl strings.Builder

	// Resource name using pingcli format
	resourceName := utils.SanitizeResourceName(instance.Name)
	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_connector_instance\" \"%s\" {\n", resourceName))

	// Environment ID
	if skipDependencies {
		hcl.WriteString(fmt.Sprintf("  environment_id = \"%s\"\n", instance.Environment.ID))
	} else {
		hcl.WriteString("  environment_id = var.environment_id\n")
	}

	hcl.WriteString("\n")

	// Name
	hcl.WriteString(fmt.Sprintf("  name           = \"%s\"\n", instance.Name))

	hcl.WriteString("\n")

	// Connector reference
	hcl.WriteString("  connector = {\n")
	hcl.WriteString(fmt.Sprintf("    id = \"%s\"\n", instance.Connector.ID))
	hcl.WriteString("  }\n")

	// Properties (if present)
	if len(instance.Properties) > 0 {
		hcl.WriteString("\n")
		writePropertiesBlock(&hcl, instance.Properties)
	}

	hcl.WriteString("}\n")

	return hcl.String()
}

// writePropertiesBlock writes the properties block with jsonencode
func writePropertiesBlock(hcl *strings.Builder, properties map[string]ConnectorPropertyValue) {
	hcl.WriteString("  properties = jsonencode({\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Find max key length for alignment
	maxKeyLen := 0
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	// Write each property
	for _, key := range keys {
		prop := properties[key]
		value := prop.Value

		// Check if this is a masked secret
		isMasked := false
		if strVal, ok := value.(string); ok && strings.Contains(strVal, "***") {
			isMasked = true
		}

		// Format the value
		var formattedValue string
		if isMasked {
			// Replace masked secrets with TODO comments
			// Convert camelCase to space-separated words
			fieldNameWords := utils.CamelCaseToWords(key)
			formattedValue = fmt.Sprintf("\"TODO: Replace with actual %s\"", strings.ToLower(fieldNameWords))
		} else {
			// Format based on type
			switch v := value.(type) {
			case string:
				formattedValue = fmt.Sprintf("\"%s\"", v)
			case bool:
				formattedValue = fmt.Sprintf("%t", v)
			case float64:
				// Check if it's an integer
				if v == float64(int64(v)) {
					formattedValue = fmt.Sprintf("%d", int64(v))
				} else {
					formattedValue = fmt.Sprintf("%f", v)
				}
			default:
				formattedValue = fmt.Sprintf("\"%v\"", v)
			}
		}

		// Write property with alignment
		padding := strings.Repeat(" ", maxKeyLen-len(key))
		hcl.WriteString(fmt.Sprintf("    \"%s\"%s : %s", key, padding, formattedValue))

		// Add comma except for last item
		if key != keys[len(keys)-1] {
			hcl.WriteString(",")
		}

		hcl.WriteString("\n")
	}

	hcl.WriteString("  })\n")
}

// GetConnectorInstanceVariableEligibleAttributes extracts variable-eligible properties from a connector instance
// Properties that match common patterns (URLs, client IDs, etc.) become module variables
func GetConnectorInstanceVariableEligibleAttributes(instanceJSON []byte, resourceName string) ([]VariableEligibleAttribute, error) {
	var instance ConnectorInstanceResponse
	if err := json.Unmarshal(instanceJSON, &instance); err != nil {
		return nil, fmt.Errorf("failed to parse connector instance JSON: %w", err)
	}

	if instance.Name == "" {
		return nil, fmt.Errorf("connector instance name is required")
	}

	// Use provided resource name or sanitize from instance name
	if resourceName == "" {
		resourceName = utils.SanitizeResourceName(instance.Name)
	}

	var attributes []VariableEligibleAttribute

	// Define properties that should typically be variables
	variableEligibleProperties := map[string]bool{
		"baseUrl":     true,
		"url":         true,
		"endpoint":    true,
		"apiUrl":      true,
		"tokenUrl":    true,
		"authUrl":     true,
		"issuerUrl":   true,
		"redirectUri": true,
		"callbackUrl": true,
		"clientId":    true,
		"tenantId":    true,
		"domain":      true,
		"region":      true,
		"environment": true,
		"namespace":   true,
		"scope":       true,
		"audience":    true,
		"realm":       true,
	}

	// Secret properties that should be variables but marked as secrets
	secretProperties := map[string]bool{
		"clientSecret": true,
		"apiKey":       true,
		"accessToken":  true,
		"password":     true,
		"secret":       true,
		"privateKey":   true,
		"certificate":  true,
	}

	// Extract variables from properties
	for propName, propValue := range instance.Properties {
		value := propValue.Value

		// Skip nil or empty values
		if value == nil {
			continue
		}

		// Skip empty strings
		if strVal, ok := value.(string); ok && strVal == "" {
			continue
		}

		// Skip masked secrets (they get TODO comments in normal HCL generation)
		if strVal, ok := value.(string); ok && strings.Contains(strVal, "****") {
			continue
		}

		// Check if this property should become a variable
		isVariableEligible := variableEligibleProperties[propName]
		isSecret := secretProperties[propName]

		// Extract if it's variable-eligible or a secret
		if !isVariableEligible && !isSecret {
			continue
		}

		// Determine Terraform type
		tfType := "string"
		switch value.(type) {
		case bool:
			tfType = "bool"
		case float64, int:
			tfType = "number"
		}

		// Create variable name: davinci_connection_{resourceName}_{propertyName}
		// Remove pingcli__ prefix for cleaner variable names
		cleanResourceName := strings.TrimPrefix(resourceName, "pingcli__")
		varName := fmt.Sprintf("davinci_connection_%s_%s", cleanResourceName, propName)

		// Create description
		description := fmt.Sprintf("%s for %s connector", propName, instance.Name)

		attr := VariableEligibleAttribute{
			ResourceType:  "connection",
			ResourceName:  resourceName,
			ResourceID:    instance.ID,
			AttributePath: fmt.Sprintf("properties.%s", propName),
			CurrentValue:  value,
			VariableName:  varName,
			VariableType:  tfType,
			Description:   description,
			Sensitive:     isSecret,
			IsSecret:      isSecret,
		}

		attributes = append(attributes, attr)
	}

	return attributes, nil
}

// GenerateConnectorInstanceHCLWithVariableReferences generates HCL with variable references for properties
func GenerateConnectorInstanceHCLWithVariableReferences(instanceJSON []byte, skipDependencies bool, variableMap map[string]string) (string, error) {
	var instance ConnectorInstanceResponse
	if err := json.Unmarshal(instanceJSON, &instance); err != nil {
		return "", fmt.Errorf("failed to parse connector instance JSON: %w", err)
	}

	if instance.Name == "" {
		return "", fmt.Errorf("connector instance name is required")
	}

	if instance.Connector.ID == "" {
		return "", fmt.Errorf("connector.id is required")
	}

	var hcl strings.Builder

	// Resource name using pingcli format
	resourceName := utils.SanitizeResourceName(instance.Name)
	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_connector_instance\" \"%s\" {\n", resourceName))

	// Environment ID
	if skipDependencies {
		hcl.WriteString(fmt.Sprintf("  environment_id = \"%s\"\n", instance.Environment.ID))
	} else {
		hcl.WriteString("  environment_id = var.environment_id\n")
	}

	hcl.WriteString("\n")

	// Name
	hcl.WriteString(fmt.Sprintf("  name           = \"%s\"\n", instance.Name))

	hcl.WriteString("\n")

	// Connector reference
	hcl.WriteString("  connector = {\n")
	hcl.WriteString(fmt.Sprintf("    id = \"%s\"\n", instance.Connector.ID))
	hcl.WriteString("  }\n")

	// Properties (if present) - with variable references
	if len(instance.Properties) > 0 {
		hcl.WriteString("\n")
		writePropertiesBlockWithVariables(&hcl, instance.Properties, variableMap)
	}

	hcl.WriteString("}\n")

	return hcl.String(), nil
}

// writePropertiesBlockWithVariables writes the properties block with variable references where applicable
func writePropertiesBlockWithVariables(hcl *strings.Builder, properties map[string]ConnectorPropertyValue, variableMap map[string]string) {
	hcl.WriteString("  properties = jsonencode({\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Find max key length for alignment
	maxKeyLen := 0
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	// Write each property
	for _, key := range keys {
		prop := properties[key]
		value := prop.Value

		// Check if this property has a variable mapping
		propertyPath := fmt.Sprintf("properties.%s", key)
		varName, hasVariable := variableMap[propertyPath]

		var formattedValue string
		if hasVariable {
			// Use variable reference
			formattedValue = fmt.Sprintf("var.%s", varName)
		} else {
			// Check if this is a masked secret
			isMasked := false
			if strVal, ok := value.(string); ok && strings.Contains(strVal, "***") {
				isMasked = true
			}

			if isMasked {
				// Replace masked secrets with TODO comments
				fieldNameWords := utils.CamelCaseToWords(key)
				formattedValue = fmt.Sprintf("\"TODO: Replace with actual %s\"", strings.ToLower(fieldNameWords))
			} else {
				// Format based on type
				switch v := value.(type) {
				case string:
					formattedValue = fmt.Sprintf("\"%s\"", v)
				case bool:
					formattedValue = fmt.Sprintf("%t", v)
				case float64:
					if v == float64(int64(v)) {
						formattedValue = fmt.Sprintf("%d", int64(v))
					} else {
						formattedValue = fmt.Sprintf("%f", v)
					}
				default:
					formattedValue = fmt.Sprintf("\"%v\"", v)
				}
			}
		}

		// Write property with alignment
		padding := strings.Repeat(" ", maxKeyLen-len(key))
		hcl.WriteString(fmt.Sprintf("    \"%s\"%s : %s", key, padding, formattedValue))

		// Add comma except for last item
		if key != keys[len(keys)-1] {
			hcl.WriteString(",")
		}

		hcl.WriteString("\n")
	}

	hcl.WriteString("  })\n")
}
