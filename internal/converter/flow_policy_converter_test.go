package converter

import (
	"strings"
	"testing"
)

// TestFlowPolicyConversion tests the conversion of flow policies to Terraform HCL
func TestFlowPolicyConversion(t *testing.T) {
	tests := []struct {
		name     string
		policyJSON string
		expected []string
	}{
		{
			name: "Real API response - basic flow policy",
			policyJSON: `{
				"_links": {
					"self": {
						"href": "https://api.pingone.com/v1/environments/62f10a04-6c54-40c2-a97d-80a98522ff9a/davinciApplications/b0ca8a26522bfccb2dcee7e25cb73ae4/flowPolicies/655f6eb61dc2875e68614438fdfdbba5"
					}
				},
				"id": "655f6eb61dc2875e68614438fdfdbba5",
				"name": "DaVinci API Protect Sample Policy",
				"environment": {
					"id": "62f10a04-6c54-40c2-a97d-80a98522ff9a"
				},
				"application": {
					"id": "b0ca8a26522bfccb2dcee7e25cb73ae4"
				},
				"status": "enabled",
				"flowDistributions": [
					{
						"id": "320448fa8e9fe59eee802b3fcbb4dfb4",
						"version": -1,
						"weight": 100
					}
				],
				"createdAt": "2025-10-01T20:32:28.526Z",
				"updatedAt": "2025-10-01T20:32:28.526Z"
			}`,
			expected: []string{
				`resource "pingone_davinci_application_flow_policy" "pingcli__DaVinci-0020-API-0020-Protect-0020-Sample-0020-Policy"`,
				`environment_id = var.environment_id`,
				`name           = "DaVinci API Protect Sample Policy"`,
				`status         = "enabled"`,
				`da_vinci_application_id = "b0ca8a26522bfccb2dcee7e25cb73ae4"`,
				`flow_distributions = [`,
				`flow_id      = "320448fa8e9fe59eee802b3fcbb4dfb4"`,
				`version      = -1`,
				`weight       = 100`,
			},
		},
		{
			name: "Flow policy with multiple flows",
			policyJSON: `{
				"id": "policy-123",
				"name": "Multi Flow Policy",
				"environment": {"id": "env-456"},
				"application": {"id": "app-789"},
				"status": "enabled",
				"flowDistributions": [
					{
						"id": "flow-1",
						"version": 5,
						"weight": 75
					},
					{
						"id": "flow-2",
						"version": -1,
						"weight": 25
					}
				]
			}`,
			expected: []string{
				`resource "pingone_davinci_application_flow_policy" "pingcli__Multi-0020-Flow-0020-Policy"`,
				`environment_id = var.environment_id`,
				`name           = "Multi Flow Policy"`,
				`status         = "enabled"`,
				`da_vinci_application_id = "app-789"`,
				`flow_distributions = [`,
				`flow_id      = "flow-1"`,
				`version      = 5`,
				`weight       = 75`,
				`flow_id      = "flow-2"`,
				`version      = -1`,
				`weight       = 25`,
			},
		},
		{
			name: "Disabled flow policy",
			policyJSON: `{
				"id": "policy-disabled",
				"name": "Disabled Policy",
				"environment": {"id": "env-456"},
				"application": {"id": "app-789"},
				"status": "disabled",
				"flowDistributions": [
					{
						"id": "flow-1",
						"version": -1,
						"weight": 100
					}
				]
			}`,
			expected: []string{
				`resource "pingone_davinci_application_flow_policy" "pingcli__Disabled-0020-Policy"`,
				`status         = "disabled"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertFlowPolicy([]byte(tt.policyJSON))
			if err != nil {
				t.Fatalf("ConvertFlowPolicy() returned error: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("ConvertFlowPolicy() missing expected element: %s\nGot:\n%s", expected, result)
				}
			}

			if !strings.HasSuffix(strings.TrimSpace(result), "}") {
				t.Error("ConvertFlowPolicy() result doesn't end with closing brace")
			}
		})
	}
}

// TestFlowPolicyConversionWithSkipDependencies tests flow policies when skip-dependencies is true
func TestFlowPolicyConversionWithSkipDependencies(t *testing.T) {
	policyJSON := `{
		"id": "policy-123",
		"name": "Test Policy",
		"environment": {"id": "env-456"},
		"application": {"id": "app-789"},
		"status": "enabled",
		"flowDistributions": [
			{
				"id": "flow-abc",
				"version": -1,
				"weight": 100
			}
		]
	}`

	result, err := ConvertFlowPolicyWithOptions([]byte(policyJSON), true)
	if err != nil {
		t.Fatalf("ConvertFlowPolicyWithOptions() returned error: %v", err)
	}

	// Should use hardcoded IDs instead of references
	if strings.Contains(result, "var.environment_id") {
		t.Error("Result should use hardcoded environment ID when skip-dependencies is true")
	}

	if !strings.Contains(result, `environment_id = "env-456"`) {
		t.Errorf("Result should contain hardcoded environment ID. Got:\n%s", result)
	}

	// Application ID should still be hardcoded (no Terraform reference for application)
	if !strings.Contains(result, `da_vinci_application_id = "app-789"`) {
		t.Errorf("Result should contain hardcoded application ID. Got:\n%s", result)
	}

	// Flow IDs should be hardcoded
	if strings.Contains(result, "pingone_davinci_flow.") {
		t.Error("Result should use hardcoded flow IDs when skip-dependencies is true")
	}

	if !strings.Contains(result, `flow_id      = "flow-abc"`) {
		t.Errorf("Result should contain hardcoded flow ID. Got:\n%s", result)
	}
}
