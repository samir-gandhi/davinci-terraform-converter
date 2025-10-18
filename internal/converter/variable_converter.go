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

	// Determine if we'll actually write a value (needed for mutable logic)
	hasValue := variable.Value != nil && !isEmptyValue(variable.Value) && variable.DataType != "secret"
	willWriteValue := false
	if hasValue {
		// Check if writeVariableValueBlock would actually write something
		willWriteValue = canWriteValue(variable.DataType, variable.Value)
	}

	// Mutable must be true when no value is set (provider requirement)
	mutable := variable.Mutable
	if !willWriteValue && !mutable {
		mutable = true // Provider requires mutable=true when value is not set
	}
	hcl.WriteString(fmt.Sprintf("  mutable        = %t\n", mutable))

	// Note if we overrode mutable
	if !variable.Mutable && mutable {
		hcl.WriteString("  # NOTE: mutable overridden to true because no value is provided (provider requirement)\n")
	} // Optional display_name
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
	// Only write value block if variable actually has a meaningful value
	// Never output secret values for security
	valueWritten := false
	if hasValue {
		hcl.WriteString("\n")
		valueWritten = writeVariableValueBlock(&hcl, variable.DataType, variable.Value)

		// If no value was actually written, add TODO comment
		if !valueWritten {
			if variable.DataType == "secret" {
				hcl.WriteString("  # TODO: Add secret value manually\n")
				hcl.WriteString("  # value = {\n")
				hcl.WriteString("  #   secret_string = \"your-secret-value\"\n")
				hcl.WriteString("  # }\n")
			} else {
				hcl.WriteString(fmt.Sprintf("  # TODO: Add %s value\n", variable.DataType))
				hcl.WriteString("  # Value omitted - will be set dynamically by flow execution\n")
			}
		}
	} else {
		// For variables without values, add a TODO comment
		hcl.WriteString("\n")
		if variable.DataType == "secret" {
			hcl.WriteString("  # TODO: Add secret value manually\n")
			hcl.WriteString("  # value = {\n")
			hcl.WriteString("  #   secret_string = \"your-secret-value\"\n")
			hcl.WriteString("  # }\n")
		} else {
			hcl.WriteString(fmt.Sprintf("  # TODO: Add %s value\n", variable.DataType))
			hcl.WriteString("  # Value omitted - will be set dynamically by flow execution\n")
		}
	}

	hcl.WriteString("}\n")

	return hcl.String()
}

// isEmptyValue checks if a value is considered empty
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	// Check for empty string or masked secret value
	if str, ok := value.(string); ok {
		return str == "" || str == "******"
	}

	// Check for zero numbers (but don't treat 0 as empty, only nil/missing)
	// Other types (bool, object) are valid even if "empty" looking
	return false
}

// canWriteValue checks if writeVariableValueBlock would actually write content for this value
func canWriteValue(dataType string, value interface{}) bool {
	switch dataType {
	case "string":
		str, ok := value.(string)
		return ok && str != "" && str != "******"
	case "number":
		switch value.(type) {
		case float64, int:
			return true
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		jsonBytes, err := json.Marshal(value)
		return err == nil && len(jsonBytes) > 2 // More than just "{}"
	case "secret":
		return false // Never write secret values
	}
	return false
}

// writeVariableValueBlock writes the value block based on data type
// Returns true if a value was written, false if nothing was written
func writeVariableValueBlock(hcl *strings.Builder, dataType string, value interface{}) bool {
	var valueContent strings.Builder
	hasContent := false

	switch dataType {
	case "string":
		if str, ok := value.(string); ok && str != "" {
			valueContent.WriteString(fmt.Sprintf("    string = \"%s\"\n", str))
			hasContent = true
		}
	case "number":
		switch v := value.(type) {
		case float64:
			// Check if it's an integer
			if v == float64(int64(v)) {
				valueContent.WriteString(fmt.Sprintf("    float32 = %d\n", int64(v)))
			} else {
				valueContent.WriteString(fmt.Sprintf("    float32 = %f\n", v))
			}
			hasContent = true
		case int:
			valueContent.WriteString(fmt.Sprintf("    float32 = %d\n", v))
			hasContent = true
		}
	case "boolean":
		if b, ok := value.(bool); ok {
			valueContent.WriteString(fmt.Sprintf("    bool = %t\n", b))
			hasContent = true
		}
	case "object":
		// Marshal the object to JSON
		jsonBytes, err := json.Marshal(value)
		if err == nil && len(jsonBytes) > 2 { // More than just "{}"
			// Use json_object attribute per provider schema
			valueContent.WriteString(fmt.Sprintf("    json_object = %s\n", string(jsonBytes)))
			hasContent = true
		}
	}

	// Only write the value block if we have actual content
	if hasContent {
		hcl.WriteString("  value = {\n")
		hcl.WriteString(valueContent.String())
		hcl.WriteString("  }\n")
	}

	return hasContent
}
