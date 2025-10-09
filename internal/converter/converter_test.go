// Copyright © 2025 Ping Identity Corporation

package converter

import (
	"strings"
	"testing"
)

// TestSimpleFlowConversion tests converting a minimal DaVinci flow to HCL.
// This flow has no connections, variables, or subflows - just basic metadata.
func TestSimpleFlowConversion(t *testing.T) {
	// Simple DaVinci flow JSON with minimal structure
	flowJSON := []byte(`{
		"name": "Simple Test Flow",
		"description": "A simple test flow",
		"flowId": "test-flow-123",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [],
				"edges": []
			}
		}
	}`)

	// Call the Convert function
	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify the result contains expected elements
	expectedElements := []string{
		`resource "pingone_davinci_flow" "simple_test_flow"`,
		`environment_id = var.environment_id`,
		`name        = "Simple Test Flow"`,
		`description = "A simple test flow"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Convert() output missing expected element: %s\nGot:\n%s", expected, result)
		}
	}
}

// TestFlowWithSingleNode tests converting a flow with one node (connection).
func TestFlowWithSingleNode(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Single Node Flow",
		"description": "Flow with one HTTP connector node",
		"flowId": "flow-single-node",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION",
							"connectionId": "conn-123-abc",
							"connectorId": "httpConnector",
							"name": "Http",
							"label": "Http",
							"capabilityName": "customHtmlMessage",
							"properties": {
								"message": {
									"value": "Hello World"
								}
							}
						}
					}
				],
				"edges": []
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify graph_data section is present
	expectedElements := []string{
		`graph_data {`,
		`elements {`,
		`nodes = [`,
		`"id": "node1"`,
		`"nodeType": "CONNECTION"`,
		`"connectionId": "conn-123-abc"`,
		`"connectorId": "httpConnector"`,
		`"capabilityName": "customHtmlMessage"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Convert() output missing expected element: %s\nGot:\n%s", expected, result)
		}
	}
}

// TestFlowWithNodesAndEdges tests a flow with multiple nodes and edges.
func TestFlowWithNodesAndEdges(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Multi Node Flow",
		"description": "Flow with nodes and edges",
		"flowId": "flow-multi",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION",
							"connectionId": "conn-123",
							"connectorId": "httpConnector",
							"capabilityName": "customHtmlMessage"
						}
					},
					{
						"data": {
							"id": "node2",
							"nodeType": "EVAL",
							"label": "Evaluator"
						}
					}
				],
				"edges": [
					{
						"data": {
							"id": "edge1",
							"source": "node1",
							"target": "node2"
						}
					}
				]
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify nodes and edges sections
	expectedElements := []string{
		`nodes = [`,
		`"id": "node1"`,
		`"id": "node2"`,
		`"nodeType": "EVAL"`,
		`edges = [`,
		`"id": "edge1"`,
		`"source": "node1"`,
		`"target": "node2"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Convert() output missing expected element: %s\nGot:\n%s", expected, result)
		}
	}
}

// TestFlowWithComplexNodeProperties tests a node with nested properties.
func TestFlowWithComplexNodeProperties(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Complex Properties Flow",
		"flowId": "flow-complex",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION",
							"connectionId": "pingone-conn",
							"connectorId": "pingOneSSOConnector",
							"capabilityName": "userLookup",
							"properties": {
								"matchAttributes": {
									"value": ["email", "username"]
								},
								"userIdentifierForFindUser": {
									"value": "{{global.parameters.email}}"
								}
							}
						}
					}
				],
				"edges": []
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify complex properties are preserved
	expectedElements := []string{
		`"properties"`,
		`"matchAttributes"`,
		`"userIdentifierForFindUser"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Convert() output missing expected element: %s\nGot:\n%s", expected, result)
		}
	}
}

// TestSanitizeResourceName tests the resource name sanitization function.
func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple name",
			input:    "My Flow",
			expected: "my_flow",
		},
		{
			name:     "Name with special characters",
			input:    "My-Flow@2024!",
			expected: "my_flow_2024",
		},
		{
			name:     "Name with multiple spaces",
			input:    "My   Test   Flow",
			expected: "my_test_flow",
		},
		{
			name:     "Already lowercase with underscores",
			input:    "my_test_flow",
			expected: "my_test_flow",
		},
		{
			name:     "Leading and trailing spaces",
			input:    "  My Flow  ",
			expected: "my_flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeResourceName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeResourceName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFlowOutputFormat verifies the HCL output format is readable.
func TestFlowOutputFormat(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Test Flow",
		"description": "A test flow for format verification",
		"flowId": "test-123",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION",
							"connectionId": "conn-abc-123",
							"connectorId": "httpConnector",
							"capabilityName": "customHtmlMessage",
							"properties": {
								"message": {
									"value": "Hello"
								}
							}
						}
					}
				],
				"edges": [
					{
						"data": {
							"id": "edge1",
							"source": "node1",
							"target": "node2"
						}
					}
				]
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Print the output for manual inspection
	t.Logf("Generated HCL:\n%s", result)

	// Verify structure
	if !strings.Contains(result, `resource "pingone_davinci_flow"`) {
		t.Error("Output missing resource declaration")
	}
	if !strings.Contains(result, "graph_data {") {
		t.Error("Output missing graph_data block")
	}
	if !strings.Contains(result, "elements {") {
		t.Error("Output missing elements block")
	}
}

// TestFlowWithSettings tests converting a flow with settings configuration.
func TestFlowWithSettings(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Flow With Settings",
		"flowId": "flow-settings",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [],
				"edges": []
			}
		},
		"settings": {
			"csp": "worker-src 'self' blob:;",
			"logLevel": 2,
			"intermediateLoadingScreenCSS": "",
			"intermediateLoadingScreenHTML": "",
			"flowHttpTimeoutInSeconds": 300
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify settings section is present
	expectedElements := []string{
		`settings {`,
		`"csp"`,
		`"logLevel"`,
		`"flowHttpTimeoutInSeconds"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Convert() output missing expected element: %s\nGot:\n%s", expected, result)
		}
	}
}

// TestFlowWithVariables tests converting a flow with variable definitions.
func TestFlowWithVariables(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Flow With Variables",
		"flowId": "flow-vars",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [],
				"edges": []
			}
		},
		"variables": [
			{
				"context": "flow",
				"name": "myVariable##SK##flow##SK##flowid",
				"fields": {
					"type": "string",
					"displayName": "My Variable",
					"value": "test value",
					"mutable": true
				}
			},
			{
				"context": "company",
				"name": "globalVar##SK##company",
				"fields": {
					"type": "number",
					"displayName": "Global Variable",
					"value": "42",
					"mutable": false
				}
			}
		]
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify variables section is present (if we implement it)
	// For now, just ensure no error and basic structure
	if !strings.Contains(result, `resource "pingone_davinci_flow"`) {
		t.Error("Output missing resource declaration")
	}

	// Check if variables are mentioned somewhere
	t.Logf("Variables test output:\n%s", result)
}

// TestFlowWithInputSchema tests converting a flow with input schema.
func TestFlowWithInputSchema(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Flow With Input Schema",
		"flowId": "flow-input",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [],
				"edges": []
			}
		},
		"inputSchemaCompiled": {
			"parameters": {
				"type": "object",
				"properties": {
					"email": {
						"type": "string",
						"description": "User email"
					},
					"password": {
						"type": "string",
						"description": "User password"
					}
				},
				"required": ["email", "password"]
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	// Verify basic structure
	if !strings.Contains(result, `resource "pingone_davinci_flow"`) {
		t.Error("Output missing resource declaration")
	}
}

// TestMalformedJSON tests error handling for invalid JSON.
func TestMalformedJSON(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Broken Flow",
		"flowId": "broken"
		"missing": "comma"
	}`)

	_, err := Convert(flowJSON)
	if err == nil {
		t.Error("Convert() should return error for malformed JSON")
	}

	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("Expected unmarshal error, got: %v", err)
	}
}

// TestEmptyJSON tests error handling for empty input.
func TestEmptyJSON(t *testing.T) {
	flowJSON := []byte(`{}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error for empty JSON: %v", err)
	}

	// Should still generate valid HCL with minimal content
	if !strings.Contains(result, `resource "pingone_davinci_flow"`) {
		t.Error("Output missing resource declaration for empty flow")
	}
}

// TestNodeWithMissingData tests handling of nodes with incomplete data.
func TestNodeWithMissingData(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Incomplete Node Flow",
		"flowId": "incomplete",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1"
						}
					}
				],
				"edges": []
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error for incomplete node: %v", err)
	}

	// Should handle gracefully and include what's available
	if !strings.Contains(result, `"id": "node1"`) {
		t.Error("Output missing node id")
	}
}

// TestEdgeWithMissingData tests handling of edges with incomplete data.
func TestEdgeWithMissingData(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Incomplete Edge Flow",
		"flowId": "incomplete-edge",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [],
				"edges": [
					{
						"data": {
							"id": "edge1",
							"source": "node1"
						}
					}
				]
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error for incomplete edge: %v", err)
	}

	// Should handle gracefully
	if !strings.Contains(result, `"id": "edge1"`) {
		t.Error("Output missing edge id")
	}
	if !strings.Contains(result, `"source": "node1"`) {
		t.Error("Output missing edge source")
	}
}

// TestFlowWithoutGraphData tests handling when graphData is missing.
func TestFlowWithoutGraphData(t *testing.T) {
	flowJSON := []byte(`{
		"name": "No Graph Flow",
		"flowId": "no-graph",
		"flowStatus": "enabled"
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error for flow without graphData: %v", err)
	}

	// Should still generate valid HCL
	if !strings.Contains(result, `resource "pingone_davinci_flow"`) {
		t.Error("Output missing resource declaration")
	}
	if !strings.Contains(result, `name        = "No Graph Flow"`) {
		t.Error("Output missing flow name")
	}
}

// TestSpecialCharactersInFlowName tests handling of special characters.
func TestSpecialCharactersInFlowName(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Test!@#$%^&*()Flow<>?:{}[]",
		"flowId": "special-chars",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [],
				"edges": []
			}
		}
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error for special characters: %v", err)
	}

	// Resource name should be sanitized
	if !strings.Contains(result, `resource "pingone_davinci_flow" "test_flow"`) {
		t.Errorf("Resource name not properly sanitized, got:\n%s", result)
	}

	// But the name attribute should preserve the original
	if !strings.Contains(result, `name        = "Test!@#$%^&*()Flow<>?:{}[]"`) {
		t.Error("Flow name not preserved in name attribute")
	}
}

// TestCompleteFlowWithAllAttributes tests a comprehensive flow with all major attributes.
func TestCompleteFlowWithAllAttributes(t *testing.T) {
	flowJSON := []byte(`{
		"name": "Complete Flow",
		"description": "A complete flow with all attributes",
		"flowId": "complete-flow-id",
		"flowStatus": "enabled",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION",
							"connectionId": "conn-123",
							"connectorId": "httpConnector",
							"capabilityName": "customHtmlMessage"
						}
					}
				],
				"edges": [
					{
						"data": {
							"id": "edge1",
							"source": "node1",
							"target": "node2"
						}
					}
				]
			}
		},
		"settings": {
			"logLevel": 2,
			"csp": "default-src 'self';"
		},
		"variables": [
			{
				"name": "testVar",
				"context": "flow",
				"fields": {
					"type": "string",
					"value": "test"
				}
			}
		]
	}`)

	result, err := Convert(flowJSON)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	t.Logf("Complete flow output:\n%s", result)

	// Verify all major sections are present
	expectedSections := []string{
		`resource "pingone_davinci_flow" "complete_flow"`,
		`environment_id = var.environment_id`,
		`name        = "Complete Flow"`,
		`description = "A complete flow with all attributes"`,
		`graph_data {`,
		`elements {`,
		`nodes = [`,
		`edges = [`,
		`settings {`,
	}

	for _, expected := range expectedSections {
		if !strings.Contains(result, expected) {
			t.Errorf("Complete flow output missing section: %s", expected)
		}
	}
}

// TestMultiFlowExport tests converting a multi-flow export (parent flow + subflows).
// When DaVinci exports include subflows, they come wrapped in a "flows" array.
// This should generate multiple separate flow resources.
func TestMultiFlowExport(t *testing.T) {
	multiFlowJSON := []byte(`{
		"flows": [
			{
				"name": "Main Flow",
				"description": "Parent flow",
				"flowId": "main-flow-id",
				"flowStatus": "enabled",
				"graphData": {
					"elements": {
						"nodes": [
							{
								"data": {
									"id": "node1",
									"nodeType": "CONNECTION",
									"connectionId": "conn-123",
									"connectorId": "httpConnector"
								}
							}
						],
						"edges": []
					}
				},
				"settings": {
					"logLevel": 4
				}
			},
			{
				"name": "Subflow One",
				"description": "First subflow",
				"flowId": "subflow-one-id",
				"flowStatus": "enabled",
				"parentFlowId": "main-flow-id",
				"graphData": {
					"elements": {
						"nodes": [
							{
								"data": {
									"id": "node2",
									"nodeType": "EVAL"
								}
							}
						],
						"edges": []
					}
				}
			},
			{
				"name": "Subflow Two",
				"description": "Second subflow",
				"flowId": "subflow-two-id",
				"flowStatus": "enabled",
				"parentFlowId": "main-flow-id",
				"graphData": {
					"elements": {
						"nodes": [],
						"edges": []
					}
				},
				"variables": [
					{
						"name": "testVar",
						"context": "flow",
						"fields": {
							"type": "string"
						}
					}
				]
			}
		],
		"companyId": "company-123",
		"customerId": "customer-456"
	}`)

	// Call ConvertMultiFlow function
	results, err := ConvertMultiFlow(multiFlowJSON)
	if err != nil {
		t.Fatalf("ConvertMultiFlow() returned error: %v", err)
	}

	// Should return 3 separate HCL resources
	if len(results) != 3 {
		t.Fatalf("ConvertMultiFlow() should return 3 flows, got %d", len(results))
	}

	// Verify first flow (Main Flow)
	mainFlow := results[0]
	expectedMainElements := []string{
		`resource "pingone_davinci_flow" "main_flow"`,
		`name        = "Main Flow"`,
		`description = "Parent flow"`,
		`graph_data {`,
		`"nodeType": "CONNECTION"`,
		`"connectionId": "conn-123"`,
		`settings {`,
		`"logLevel": 4`,
	}

	for _, expected := range expectedMainElements {
		if !strings.Contains(mainFlow, expected) {
			t.Errorf("Main flow missing expected element: %s\nGot:\n%s", expected, mainFlow)
		}
	}

	// Verify second flow (Subflow One)
	subflowOne := results[1]
	expectedSubflowOneElements := []string{
		`resource "pingone_davinci_flow" "subflow_one"`,
		`name        = "Subflow One"`,
		`description = "First subflow"`,
		`"nodeType": "EVAL"`,
	}

	for _, expected := range expectedSubflowOneElements {
		if !strings.Contains(subflowOne, expected) {
			t.Errorf("Subflow One missing expected element: %s\nGot:\n%s", expected, subflowOne)
		}
	}

	// Verify third flow (Subflow Two)
	subflowTwo := results[2]
	expectedSubflowTwoElements := []string{
		`resource "pingone_davinci_flow" "subflow_two"`,
		`name        = "Subflow Two"`,
		`description = "Second subflow"`,
		`# Flow Variables:`,
		`Name: testVar`,
		`Context: flow`,
	}

	for _, expected := range expectedSubflowTwoElements {
		if !strings.Contains(subflowTwo, expected) {
			t.Errorf("Subflow Two missing expected element: %s\nGot:\n%s", expected, subflowTwo)
		}
	}

	// Log all outputs for visual inspection
	t.Logf("Main Flow:\n%s\n", mainFlow)
	t.Logf("Subflow One:\n%s\n", subflowOne)
	t.Logf("Subflow Two:\n%s\n", subflowTwo)
}

// TestSingleFlowWrappedInFlowsArray tests that a single flow wrapped in "flows" array
// is still handled correctly (backwards compatibility).
func TestSingleFlowWrappedInFlowsArray(t *testing.T) {
	singleFlowJSON := []byte(`{
		"flows": [
			{
				"name": "Wrapped Flow",
				"description": "Single flow in array",
				"flowId": "wrapped-flow-id",
				"flowStatus": "enabled",
				"graphData": {
					"elements": {
						"nodes": [],
						"edges": []
					}
				}
			}
		]
	}`)

	results, err := ConvertMultiFlow(singleFlowJSON)
	if err != nil {
		t.Fatalf("ConvertMultiFlow() returned error: %v", err)
	}

	// Should return 1 flow
	if len(results) != 1 {
		t.Fatalf("ConvertMultiFlow() should return 1 flow, got %d", len(results))
	}

	// Verify the flow was converted correctly
	flow := results[0]
	expectedElements := []string{
		`resource "pingone_davinci_flow" "wrapped_flow"`,
		`name        = "Wrapped Flow"`,
		`description = "Single flow in array"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(flow, expected) {
			t.Errorf("Wrapped flow missing expected element: %s", expected)
		}
	}
}

// TestEmptyFlowsArray tests handling of empty flows array (edge case).
func TestEmptyFlowsArray(t *testing.T) {
	emptyJSON := []byte(`{
		"flows": []
	}`)

	results, err := ConvertMultiFlow(emptyJSON)
	if err != nil {
		t.Fatalf("ConvertMultiFlow() should not error on empty array, got: %v", err)
	}

	// Should return empty array
	if len(results) != 0 {
		t.Errorf("ConvertMultiFlow() should return 0 flows for empty array, got %d", len(results))
	}
}
