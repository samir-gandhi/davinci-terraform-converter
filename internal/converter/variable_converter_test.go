package converter

import (
	"strings"
	"testing"
)

// TestVariableConversion tests the conversion of DaVinci variables to Terraform HCL
func TestVariableConversion(t *testing.T) {
	tests := []struct {
		name     string
		varJSON  string
		expected []string
	}{
		{
			name: "Company context string variable",
			varJSON: `{
				"id": "cccde8f2-c5e2-46f5-80cf-efe5c4e40ee2",
				"environment": {
					"id": "62f10a04-6c54-40c2-a97d-80a98522ff9a"
				},
				"name": "SampleVariable",
				"dataType": "string",
				"displayName": "Sample Company Context Variable",
				"context": "company",
				"value": "FOOBAR",
				"mutable": true,
				"min": 0,
				"max": 2000,
				"createdAt": "2025-10-11T22:27:54.402Z",
				"updatedAt": "2025-10-11T22:27:54.402Z"
			}`,
			expected: []string{
				`resource "pingone_davinci_variable" "pingcli__SampleVariable_company"`,
				`environment_id = var.environment_id`,
				`name           = "SampleVariable"`,
				`context        = "company"`,
				`data_type      = "string"`,
				`mutable        = true`,
				`display_name   = "Sample Company Context Variable"`,
				`min            = 0`,
				`max            = 2000`,
				`value = {`,
				`string = "FOOBAR"`,
			},
		},
		{
			name: "Number variable with min/max",
			varJSON: `{
				"id": "var-num",
				"environment": {"id": "env-456"},
				"name": "maxRetries",
				"dataType": "number",
				"context": "company",
				"value": 3,
				"mutable": false,
				"min": 0,
				"max": 10
			}`,
			expected: []string{
				`resource "pingone_davinci_variable" "pingcli__maxRetries_company"`,
				`environment_id = var.environment_id`,
				`name           = "maxRetries"`,
				`context        = "company"`,
				`data_type      = "number"`,
				`mutable        = false`,
				`min            = 0`,
				`max            = 10`,
				`value = {`,
				`float32 = 3`,
			},
		},
		{
			name: "Boolean variable",
			varJSON: `{
				"id": "var-bool",
				"environment": {"id": "env-456"},
				"name": "featureEnabled",
				"dataType": "boolean",
				"context": "company",
				"value": true,
				"mutable": true
			}`,
			expected: []string{
				`resource "pingone_davinci_variable" "pingcli__featureEnabled_company"`,
				`environment_id = var.environment_id`,
				`name           = "featureEnabled"`,
				`context        = "company"`,
				`data_type      = "boolean"`,
				`mutable        = true`,
				`value = {`,
				`bool = true`,
			},
		},
		{
			name: "Secret variable (value should not be included)",
			varJSON: `{
				"id": "var-secret",
				"environment": {"id": "env-456"},
				"name": "apiKey",
				"dataType": "secret",
				"context": "company",
				"value": "super-secret-value",
				"mutable": false
			}`,
			expected: []string{
				`resource "pingone_davinci_variable" "pingcli__apiKey_company"`,
				`environment_id = var.environment_id`,
				`name           = "apiKey"`,
				`context        = "company"`,
				`data_type      = "secret"`,
				`mutable        = true`, // Must be true when no value is provided (provider requirement)
				`# TODO: Add secret value manually`,
			},
		},
		{
			name: "Object variable",
			varJSON: `{
				"id": "var-obj",
				"environment": {"id": "env-456"},
				"name": "appConfig",
				"dataType": "object",
				"context": "company",
				"value": {"timeout": 30, "retries": 3},
				"mutable": true
			}`,
			expected: []string{
				`resource "pingone_davinci_variable" "pingcli__appConfig_company"`,
				`environment_id = var.environment_id`,
				`name           = "appConfig"`,
				`context        = "company"`,
				`data_type      = "object"`,
				`mutable        = true`,
				`value = {`,
				`json_object =`,
			},
		},
		{
			name: "Flow context variable (no value, requires flow reference)",
			varJSON: `{
				"id": "var-flow",
				"environment": {"id": "env-456"},
				"name": "flowVar",
				"dataType": "string",
				"context": "flow",
				"mutable": true,
				"flow": {
					"id": "flow-123"
				}
			}`,
			expected: []string{
				`resource "pingone_davinci_variable" "pingcli__flowVar_flow"`,
				`environment_id = var.environment_id`,
				`name           = "flowVar"`,
				`context        = "flow"`,
				`data_type      = "string"`,
				`mutable        = true`,
				`flow = {`,
				`id = "flow-123"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertVariable([]byte(tt.varJSON))
			if err != nil {
				t.Fatalf("ConvertVariable() returned error: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("ConvertVariable() missing expected element: %s\nGot:\n%s", expected, result)
				}
			}

			if !strings.HasSuffix(strings.TrimSpace(result), "}") {
				t.Error("ConvertVariable() result doesn't end with closing brace")
			}
		})
	}
}

// TestVariableConversionWithSkipDependencies tests variables when skip-dependencies is true
func TestVariableConversionWithSkipDependencies(t *testing.T) {
	varJSON := `{
		"id": "var-123",
		"environment": {"id": "env-456"},
		"name": "testVar",
		"dataType": "string",
		"context": "company",
		"value": "test",
		"mutable": true,
		"flow": {
			"id": "flow-789"
		}
	}`

	result, err := ConvertVariableWithOptions([]byte(varJSON), true)
	if err != nil {
		t.Fatalf("ConvertVariableWithOptions() returned error: %v", err)
	}

	// Should use hardcoded IDs instead of references
	if strings.Contains(result, "var.environment_id") {
		t.Error("Result should use hardcoded environment ID when skip-dependencies is true")
	}

	if !strings.Contains(result, `environment_id = "env-456"`) {
		t.Errorf("Result should contain hardcoded environment ID. Got:\n%s", result)
	}

	if strings.Contains(result, "pingone_davinci_flow.") {
		t.Error("Result should use hardcoded flow ID when skip-dependencies is true")
	}

	if !strings.Contains(result, `id = "flow-789"`) {
		t.Errorf("Result should contain hardcoded flow ID. Got:\n%s", result)
	}
}

// TestDuplicateVariableNamesWithDifferentContexts tests that variables with the same name
// but different contexts generate unique resource names
func TestDuplicateVariableNamesWithDifferentContexts(t *testing.T) {
	tests := []struct {
		name             string
		varJSON          string
		expectedResource string
		expectedContext  string
	}{
		{
			name: "origin variable with company context",
			varJSON: `{
				"id": "var-origin-company",
				"environment": {"id": "env-123"},
				"name": "origin",
				"dataType": "string",
				"context": "company",
				"value": "https://company.example.com",
				"mutable": true,
				"min": 0,
				"max": 2000
			}`,
			expectedResource: `resource "pingone_davinci_variable" "pingcli__origin_company"`,
			expectedContext:  `context        = "company"`,
		},
		{
			name: "origin variable with flowInstance context",
			varJSON: `{
				"id": "var-origin-flowInstance",
				"environment": {"id": "env-123"},
				"name": "origin",
				"dataType": "string",
				"context": "flowInstance",
				"value": "https://flow.example.com",
				"mutable": true,
				"min": 0,
				"max": 2000
			}`,
			expectedResource: `resource "pingone_davinci_variable" "pingcli__origin_flowInstance"`,
			expectedContext:  `context        = "flowInstance"`,
		},
		{
			name: "enableFeatureX variable with company context (boolean)",
			varJSON: `{
				"id": "var-feature-company",
				"environment": {"id": "env-123"},
				"name": "enableFeatureX",
				"dataType": "boolean",
				"context": "company",
				"value": true,
				"mutable": false
			}`,
			expectedResource: `resource "pingone_davinci_variable" "pingcli__enableFeatureX_company"`,
			expectedContext:  `context        = "company"`,
		},
		{
			name: "enableFeatureX variable with flowInstance context (string)",
			varJSON: `{
				"id": "var-feature-flowInstance",
				"environment": {"id": "env-123"},
				"name": "enableFeatureX",
				"dataType": "string",
				"context": "flowInstance",
				"value": "enabled",
				"mutable": true
			}`,
			expectedResource: `resource "pingone_davinci_variable" "pingcli__enableFeatureX_flowInstance"`,
			expectedContext:  `context        = "flowInstance"`,
		},
		{
			name: "apiKey variable with user context",
			varJSON: `{
				"id": "var-apikey-user",
				"environment": {"id": "env-123"},
				"name": "apiKey",
				"dataType": "string",
				"context": "user",
				"value": "user-api-key",
				"mutable": true
			}`,
			expectedResource: `resource "pingone_davinci_variable" "pingcli__apiKey_user"`,
			expectedContext:  `context        = "user"`,
		},
		{
			name: "apiKey variable with flow context",
			varJSON: `{
				"id": "var-apikey-flow",
				"environment": {"id": "env-123"},
				"name": "apiKey",
				"dataType": "string",
				"context": "flow",
				"mutable": true,
				"flow": {
					"id": "flow-456"
				}
			}`,
			expectedResource: `resource "pingone_davinci_variable" "pingcli__apiKey_flow"`,
			expectedContext:  `context        = "flow"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertVariable([]byte(tt.varJSON))
			if err != nil {
				t.Fatalf("ConvertVariable() returned error: %v", err)
			}

			// Verify unique resource name includes context
			if !strings.Contains(result, tt.expectedResource) {
				t.Errorf("ConvertVariable() missing expected resource declaration: %s\nGot:\n%s", tt.expectedResource, result)
			}

			// Verify context is preserved in the resource
			if !strings.Contains(result, tt.expectedContext) {
				t.Errorf("ConvertVariable() missing expected context: %s\nGot:\n%s", tt.expectedContext, result)
			}
		})
	}

	// Test that we can generate multiple variables with same name but different contexts
	// without name collision
	t.Run("Multiple variables with same name generate unique resources", func(t *testing.T) {
		var results []string

		// Generate all origin variables
		originCompany := `{
			"id": "var-1",
			"environment": {"id": "env-123"},
			"name": "origin",
			"dataType": "string",
			"context": "company",
			"value": "company-origin",
			"mutable": true
		}`
		result1, _ := ConvertVariable([]byte(originCompany))
		results = append(results, result1)

		originFlowInstance := `{
			"id": "var-2",
			"environment": {"id": "env-123"},
			"name": "origin",
			"dataType": "string",
			"context": "flowInstance",
			"value": "flow-origin",
			"mutable": true
		}`
		result2, _ := ConvertVariable([]byte(originFlowInstance))
		results = append(results, result2)

		originUser := `{
			"id": "var-3",
			"environment": {"id": "env-123"},
			"name": "origin",
			"dataType": "string",
			"context": "user",
			"value": "user-origin",
			"mutable": true
		}`
		result3, _ := ConvertVariable([]byte(originUser))
		results = append(results, result3)

		// Verify all have unique resource names
		resourceNames := make(map[string]int)
		for i, result := range results {
			lines := strings.Split(result, "\n")
			if len(lines) > 0 {
				firstLine := lines[0]
				if count, exists := resourceNames[firstLine]; exists {
					t.Errorf("Duplicate resource declaration found (occurrence %d):\n%s\nFrom result %d:\n%s",
						count+1, firstLine, i+1, result)
				}
				resourceNames[firstLine] = resourceNames[firstLine] + 1
			}
		}

		// Verify we have 3 unique resource names
		if len(resourceNames) != 3 {
			t.Errorf("Expected 3 unique resource names, got %d", len(resourceNames))
		}
	})
}
