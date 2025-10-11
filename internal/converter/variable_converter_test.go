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
				`resource "pingone_davinci_variable" "pingcli__SampleVariable"`,
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
				"id": "var-123",
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
				`resource "pingone_davinci_variable" "pingcli__maxRetries"`,
				`environment_id = var.environment_id`,
				`name           = "maxRetries"`,
				`context        = "company"`,
				`data_type      = "number"`,
				`mutable        = false`,
				`min            = 0`,
				`max            = 10`,
				`value = {`,
				`number = 3`,
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
				`resource "pingone_davinci_variable" "pingcli__featureEnabled"`,
				`environment_id = var.environment_id`,
				`name           = "featureEnabled"`,
				`context        = "company"`,
				`data_type      = "boolean"`,
				`mutable        = true`,
				`value = {`,
				`boolean = true`,
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
				`resource "pingone_davinci_variable" "pingcli__apiKey"`,
				`environment_id = var.environment_id`,
				`name           = "apiKey"`,
				`context        = "company"`,
				`data_type      = "secret"`,
				`mutable        = false`,
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
				`resource "pingone_davinci_variable" "pingcli__appConfig"`,
				`environment_id = var.environment_id`,
				`name           = "appConfig"`,
				`context        = "company"`,
				`data_type      = "object"`,
				`mutable        = true`,
				`value = {`,
				`object = jsonencode({`,
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
				`resource "pingone_davinci_variable" "pingcli__flowVar"`,
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
