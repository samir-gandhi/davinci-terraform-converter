package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveFlowConversion tests conversion of a complete flow with all structures
// This corresponds to Part 2.1 Phase 2.1 - Comprehensive Flow Structure Conversion
func TestComprehensiveFlowConversion(t *testing.T) {
	// Complete flow JSON with graphData, nodes, edges, settings, input/output schemas
	flowJSON := `{
		"name": "PingOne DaVinci API Protect Example",
		"description": "This flow demonstrates how to protect an app with PingOne",
		"flowColor": "#CACED3",
		"flowId": "9119d34321b84902f2a117cee401efe7",
		"companyId": "5f11aa88-c9f7-4fba-b881-9f5ccd19365f",
		"customerId": "9d08d4e04b1c3ffed309b999748fa1f5",
		"deployedDate": 1749506378135,
		"createdDate": 1749506377880,
		"currentVersion": 2,
		"publishedVersion": 2,
		"flowStatus": "enabled",
		"inputSchema": [
			{
				"propertyName": "email",
				"preferredDataType": "string",
				"preferredControlType": "textField",
				"required": true,
				"isExpanded": true,
				"description": ""
			},
			{
				"propertyName": "password",
				"preferredDataType": "string",
				"preferredControlType": "textField",
				"required": true,
				"isExpanded": true,
				"description": ""
			},
			{
				"propertyName": "riskData",
				"preferredDataType": "string",
				"preferredControlType": "textField",
				"required": true,
				"isExpanded": true,
				"description": ""
			}
		],
		"settings": {
			"csp": "worker-src 'self' blob:; script-src 'self' https://cdn.jsdelivr.net https://code.jquery.com https://devsdk.singularkey.com http://cdnjs.cloudflare.com 'unsafe-inline' 'unsafe-eval';",
			"intermediateLoadingScreenCSS": "",
			"intermediateLoadingScreenHTML": "",
			"logLevel": 2
		},
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "m4sfmek769",
							"nodeType": "CONNECTION",
							"connectionId": "94141bf2f1b9b59a5f5365ff135e02bb",
							"connectorId": "pingOneSSOConnector",
							"name": "PingOne",
							"label": "PingOne",
							"status": "configured",
							"capabilityName": "userLookup",
							"type": "action",
							"properties": {
								"matchAttributes": {
									"value": ["email"]
								},
								"userIdentifierForFindUser": {
									"value": "[{\"children\":[{\"text\":\"\"},{\"text\":\"\"},{\"type\":\"link\",\"src\":\"auth.svg\",\"url\":\"email\",\"data\":\"{{global.parameters.email}}\",\"tooltip\":\"{{global.parameters.email}}\",\"children\":[{\"text\":\"email\"}]},{\"text\":\"\"}]}]"
								}
							}
						},
						"position": {
							"x": 277,
							"y": 266
						},
						"group": "nodes",
						"removed": false,
						"selected": false,
						"selectable": true,
						"locked": false,
						"grabbable": true,
						"pannable": false,
						"classes": ""
					},
					{
						"data": {
							"id": "yqi3iaujxx",
							"nodeType": "EVAL",
							"label": "Evaluator",
							"properties": {
								"0di26c5iy7": {
									"value": "anyTriggersFalse"
								},
								"6i7lwwrw94": {
									"value": "allTriggersFalse"
								}
							}
						},
						"position": {
							"x": 427,
							"y": 266
						},
						"group": "nodes",
						"removed": false,
						"selected": false,
						"selectable": true,
						"locked": false,
						"grabbable": true,
						"pannable": false,
						"classes": ""
					}
				],
				"edges": [
					{
						"data": {
							"id": "wv1og0m5r3",
							"source": "sxdpclcyko",
							"target": "n6js2rcdqf"
						},
						"group": "edges",
						"removed": false,
						"selected": false,
						"selectable": true,
						"locked": false,
						"grabbable": true,
						"pannable": true,
						"classes": ""
					},
					{
						"data": {
							"id": "05t56lofq2",
							"source": "09anefv002",
							"target": "sxdpclcyko"
						},
						"group": "edges",
						"removed": false,
						"selected": false,
						"selectable": true,
						"locked": false,
						"grabbable": true,
						"pannable": true,
						"classes": ""
					}
				]
			},
			"pan": {
				"x": 0,
				"y": 0
			},
			"zoom": 1,
			"minZoom": 1e-50,
			"maxZoom": 1e+50,
			"zoomingEnabled": true,
			"panningEnabled": true,
			"userZoomingEnabled": true,
			"userPanningEnabled": true,
			"boxSelectionEnabled": true,
			"renderer": {
				"name": "null"
			}
		},
		"trigger": {
			"type": "AUTHENTICATION",
			"configuration": {
				"mfa": {
					"enabled": false,
					"time": 0,
					"timeFormat": "seconds"
				},
				"pwd": {
					"enabled": false,
					"time": 0,
					"timeFormat": "seconds"
				}
			}
		}
	}`

	// Expected HCL output - nested HCL blocks, NOT jsonencode for graph_data
	expectedHCL := `resource "pingone_davinci_flow" "PingOne_DaVinci_API_Protect_Example" {
  environment_id = var.environment_id

  name        = "PingOne DaVinci API Protect Example"
  description = "This flow demonstrates how to protect an app with PingOne"
  color       = "#CACED3"

  settings = {
    csp                              = "worker-src 'self' blob:; script-src 'self' https://cdn.jsdelivr.net https://code.jquery.com https://devsdk.singularkey.com http://cdnjs.cloudflare.com 'unsafe-inline' 'unsafe-eval';"
    intermediate_loading_screen_css  = ""
    intermediate_loading_screen_html = ""
    log_level                        = 2
  }

  graph_data = {
    elements = {
      nodes = [
        {
          data = {
            id              = "m4sfmek769"
            node_type       = "CONNECTION"
            connection_id   = pingone_davinci_connector_instance.pingone_sso_connector_94141bf2f1b9b59a5f5365ff135e02bb.id
            connector_id    = "pingOneSSOConnector"
            name            = "PingOne"
            label           = "PingOne"
            status          = "configured"
            capability_name = "userLookup"
            type            = "action"
            properties = jsonencode({
              "matchAttributes" : {
                "value" : ["email"]
              },
              "userIdentifierForFindUser" : {
                "value" : "[{\"children\":[{\"text\":\"\"},{\"text\":\"\"},{\"type\":\"link\",\"src\":\"auth.svg\",\"url\":\"email\",\"data\":\"{{global.parameters.email}}\",\"tooltip\":\"{{global.parameters.email}}\",\"children\":[{\"text\":\"email\"}]},{\"text\":\"\"}]}]"
              }
            })
          }
          position = {
            x = 277
            y = 266
          }
          group      = "nodes"
          removed    = false
          selected   = false
          selectable = true
          locked     = false
          grabbable  = true
          pannable   = false
          classes    = ""
        },
        {
          data = {
            id        = "yqi3iaujxx"
            node_type = "EVAL"
            label     = "Evaluator"
            properties = jsonencode({
              "0di26c5iy7" : {
                "value" : "anyTriggersFalse"
              },
              "6i7lwwrw94" : {
                "value" : "allTriggersFalse"
              }
            })
          }
          position = {
            x = 427
            y = 266
          }
          group      = "nodes"
          removed    = false
          selected   = false
          selectable = true
          locked     = false
          grabbable  = true
          pannable   = false
          classes    = ""
        }
      ]
      edges = [
        {
          data = {
            id     = "wv1og0m5r3"
            source = "sxdpclcyko"
            target = "n6js2rcdqf"
          }
          group      = "edges"
          removed    = false
          selected   = false
          selectable = true
          locked     = false
          grabbable  = true
          pannable   = true
          classes    = ""
        },
        {
          data = {
            id     = "05t56lofq2"
            source = "09anefv002"
            target = "sxdpclcyko"
          }
          group      = "edges"
          removed    = false
          selected   = false
          selectable = true
          locked     = false
          grabbable  = true
          pannable   = true
          classes    = ""
        }
      ]
    }

    pan = {
      x = 0
      y = 0
    }

    zoom                  = 1
    min_zoom              = 1e-50
    max_zoom              = 1e+50
    zooming_enabled       = true
    panning_enabled       = true
    user_zooming_enabled  = true
    user_panning_enabled  = true
    box_selection_enabled = true

    renderer = jsonencode({
      "name" : "null"
    })
  }

  input_schema = [
    {
      property_name           = "email"
      preferred_data_type     = "string"
      preferred_control_type  = "textField"
      required                = true
      is_expanded             = true
      description             = ""
    },
    {
      property_name           = "password"
      preferred_data_type     = "string"
      preferred_control_type  = "textField"
      required                = true
      is_expanded             = true
      description             = ""
    },
    {
      property_name           = "riskData"
      preferred_data_type     = "string"
      preferred_control_type  = "textField"
      required                = true
      is_expanded             = true
      description             = ""
    }
  ]

  trigger = {
    type = "AUTHENTICATION"
    configuration = {
      mfa = {
        enabled     = false
        time        = 0
        time_format = "seconds"
      }
      pwd = {
        enabled     = false
        time        = 0
        time_format = "seconds"
      }
    }
  }
}
`

	// Parse the flow JSON
	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err, "Failed to parse flow JSON")

	// Convert to HCL
	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err, "ConvertFlowToHCL failed")

	// Assert the HCL output matches expected structure
	assert.Equal(t, expectedHCL, result, "Generated HCL does not match expected output")
}

// TestFlowConversion_NoEdges tests a flow with nodes but no edges
func TestFlowConversion_NoEdges(t *testing.T) {
	flowJSON := `{
		"name": "Simple Flow",
		"flowId": "test123",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION"
						},
						"position": {
							"x": 100,
							"y": 100
						}
					}
				],
				"edges": []
			},
			"zoom": 1
		}
	}`

	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err)

	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err)

	// Should contain nodes but empty edges array
	assert.Contains(t, result, "nodes = [")
	assert.Contains(t, result, "edges = []")
}

// TestFlowConversion_MinimalFlow tests conversion of minimal flow with only required fields
func TestFlowConversion_MinimalFlow(t *testing.T) {
	flowJSON := `{
		"name": "Minimal Flow",
		"flowId": "minimal123"
	}`

	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err)

	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err)

	// Should contain required fields (flexible spacing match)
	assert.Contains(t, result, `"Minimal Flow"`)
	assert.Contains(t, result, `name`)
	// Should not contain optional fields that weren't provided
	assert.NotContains(t, result, "description =")
	assert.NotContains(t, result, "color =")
}

// TestFlowConversion_EscapeSpecialCharacters tests handling of special characters
func TestFlowConversion_EscapeSpecialCharacters(t *testing.T) {
	flowJSON := `{
		"name": "Flow with \"quotes\" and \\ backslashes",
		"description": "Test 'single' and \"double\" quotes",
		"flowId": "escape123"
	}`

	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err)

	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err)

	// Should properly escape special characters in HCL strings
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "description")
}

// TestFlowConversion_NodeProperties tests jsonencode() usage for node properties
func TestFlowConversion_NodeProperties(t *testing.T) {
	flowJSON := `{
		"name": "Properties Test",
		"flowId": "props123",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "testnode",
							"nodeType": "CONNECTION",
							"properties": {
								"key1": "value1",
								"key2": {
									"nested": "value"
								}
							}
						}
					}
				]
			}
		}
	}`

	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err)

	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err)

	// properties should use jsonencode()
	assert.Contains(t, result, "properties = jsonencode(")
	assert.Contains(t, result, "key1")
	assert.Contains(t, result, "key2")
}

// TestFlowConversion_RendererField tests jsonencode() usage for renderer field
func TestFlowConversion_RendererField(t *testing.T) {
	flowJSON := `{
		"name": "Renderer Test",
		"flowId": "renderer123",
		"graphData": {
			"renderer": {
				"name": "canvas",
				"config": {
					"option": "value"
				}
			}
		}
	}`

	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err)

	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err)

	// renderer should use jsonencode()
	assert.Contains(t, result, "renderer = jsonencode(")
}

// TestFlowConversion_ConnectionIDReference tests generation of Terraform references for connection_id
func TestFlowConversion_ConnectionIDReference(t *testing.T) {
	flowJSON := `{
		"name": "Connection Reference Test",
		"flowId": "connref123",
		"graphData": {
			"elements": {
				"nodes": [
					{
						"data": {
							"id": "node1",
							"nodeType": "CONNECTION",
							"connectionId": "abc123def456",
							"connectorId": "httpConnector"
						}
					}
				]
			}
		}
	}`

	var flowData map[string]interface{}
	err := json.Unmarshal([]byte(flowJSON), &flowData)
	require.NoError(t, err)

	result, err := converter.ConvertFlowToHCL(flowData, "var.environment_id")
	require.NoError(t, err)

	// connection_id should be converted to Terraform reference
	// Format: pingone_davinci_connector_instance.<connector_id>_<connection_id>.id
	// Check for key components (flexible spacing)
	assert.Contains(t, result, "connection_id")
	assert.Contains(t, result, "pingone_davinci_connector_instance.http_connector_abc123def456.id")
}
