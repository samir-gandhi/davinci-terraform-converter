package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

// VariableResponse represents the DaVinci API response for a variable
type VariableResponse struct {
	ID          string `json:"id"`
	Environment struct {
		ID string `json:"id"`
	} `json:"environment"`
	Name        string      `json:"name"`
	DataType    string      `json:"dataType"`
	DisplayName string      `json:"displayName,omitempty"`
	Context     string      `json:"context"`
	Value       interface{} `json:"value,omitempty"`
	Mutable     bool        `json:"mutable"`
	Min         *int        `json:"min,omitempty"`
	Max         *int        `json:"max,omitempty"`
	Flow        *struct {
		ID string `json:"id"`
	} `json:"flow,omitempty"`
}

// ConvertVariable converts a DaVinci variable JSON to Terraform HCL
func ConvertVariable(variableJSON []byte) (string, error) {
	return ConvertVariableWithOptions(variableJSON, false)
}

// ConvertVariableWithOptions converts a variable with optional skip-dependencies flag
func ConvertVariableWithOptions(variableJSON []byte, skipDependencies bool) (string, error) {
	var variable VariableResponse
	if err := json.Unmarshal(variableJSON, &variable); err != nil {
		return "", fmt.Errorf("failed to parse variable JSON: %w", err)
	}

	if variable.Name == "" {
		return "", fmt.Errorf("variable name is required")
	}

	if variable.Context == "" {
		return "", fmt.Errorf("variable context is required")
	}

	if variable.DataType == "" {
		return "", fmt.Errorf("variable data_type is required")
	}

	return generateVariableHCL(variable, skipDependencies), nil
}

// generateVariableHCL generates the Terraform HCL for a variable
func generateVariableHCL(variable VariableResponse, skipDependencies bool) string {
	var hcl strings.Builder

	// Resource name using pingcli format
	resourceName := utils.SanitizeResourceName(variable.Name)
	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_variable\" \"%s\" {\n", resourceName))

	// Environment ID
	if skipDependencies {
		hcl.WriteString(fmt.Sprintf("  environment_id = \"%s\"\n", variable.Environment.ID))
	} else {
		hcl.WriteString("  environment_id = var.environment_id\n")
	}

	hcl.WriteString("\n")

	// Required attributes
	hcl.WriteString(fmt.Sprintf("  name           = \"%s\"\n", variable.Name))
	hcl.WriteString(fmt.Sprintf("  context        = \"%s\"\n", variable.Context))
	hcl.WriteString(fmt.Sprintf("  data_type      = \"%s\"\n", variable.DataType))
	hcl.WriteString(fmt.Sprintf("  mutable        = %t\n", variable.Mutable))

	// Optional display_name
	if variable.DisplayName != "" {
		hcl.WriteString(fmt.Sprintf("  display_name   = \"%s\"\n", variable.DisplayName))
	}

	// Optional min/max (for number type)
	if variable.Min != nil {
		hcl.WriteString(fmt.Sprintf("  min            = %d\n", *variable.Min))
	}
	if variable.Max != nil {
		hcl.WriteString(fmt.Sprintf("  max            = %d\n", *variable.Max))
	}

	// Flow reference (for flow context)
	if variable.Flow != nil {
		hcl.WriteString("\n")
		hcl.WriteString("  flow = {\n")
		if skipDependencies {
			hcl.WriteString(fmt.Sprintf("    id = \"%s\"\n", variable.Flow.ID))
		} else {
			// TODO: Part 4 - resolve flow reference
			hcl.WriteString(fmt.Sprintf("    id = \"%s\"  # TODO: Replace with flow reference\n", variable.Flow.ID))
		}
		hcl.WriteString("  }\n")
	}

	// Value block (type-specific)
	if variable.Value != nil {
		hcl.WriteString("\n")
		writeVariableValueBlock(&hcl, variable.DataType, variable.Value)
	} else if variable.DataType == "secret" {
		// For secrets without values, add a TODO comment
		hcl.WriteString("\n")
		hcl.WriteString("  # TODO: Add secret value manually\n")
		hcl.WriteString("  # value = {\n")
		hcl.WriteString("  #   secret = \"your-secret-value\"\n")
		hcl.WriteString("  # }\n")
	}

	hcl.WriteString("}\n")

	return hcl.String()
}

// writeVariableValueBlock writes the value block based on data type
func writeVariableValueBlock(hcl *strings.Builder, dataType string, value interface{}) {
	// Don't output secret values for security
	if dataType == "secret" {
		hcl.WriteString("  # TODO: Add secret value manually\n")
		hcl.WriteString("  # value = {\n")
		hcl.WriteString("  #   secret = \"your-secret-value\"\n")
		hcl.WriteString("  # }\n")
		return
	}

	hcl.WriteString("  value = {\n")

	switch dataType {
	case "string":
		if str, ok := value.(string); ok {
			hcl.WriteString(fmt.Sprintf("    string = \"%s\"\n", str))
		}
	case "number":
		switch v := value.(type) {
		case float64:
			// Check if it's an integer
			if v == float64(int64(v)) {
				hcl.WriteString(fmt.Sprintf("    number = %d\n", int64(v)))
			} else {
				hcl.WriteString(fmt.Sprintf("    number = %f\n", v))
			}
		case int:
			hcl.WriteString(fmt.Sprintf("    number = %d\n", v))
		}
	case "boolean":
		if b, ok := value.(bool); ok {
			hcl.WriteString(fmt.Sprintf("    boolean = %t\n", b))
		}
	case "object":
		// Marshal the object to JSON
		jsonBytes, err := json.Marshal(value)
		if err == nil {
			hcl.WriteString("    object = jsonencode({\n")
			// Pretty print the JSON content (simple approach)
			var obj map[string]interface{}
			if json.Unmarshal(jsonBytes, &obj) == nil {
				for k, v := range obj {
					switch val := v.(type) {
					case string:
						hcl.WriteString(fmt.Sprintf("      \"%s\" : \"%s\",\n", k, val))
					case float64:
						if val == float64(int64(val)) {
							hcl.WriteString(fmt.Sprintf("      \"%s\" : %d,\n", k, int64(val)))
						} else {
							hcl.WriteString(fmt.Sprintf("      \"%s\" : %f,\n", k, val))
						}
					case bool:
						hcl.WriteString(fmt.Sprintf("      \"%s\" : %t,\n", k, val))
					default:
						hcl.WriteString(fmt.Sprintf("      \"%s\" : %v,\n", k, val))
					}
				}
			}
			hcl.WriteString("    })\n")
		}
	}

	hcl.WriteString("  }\n")
}
