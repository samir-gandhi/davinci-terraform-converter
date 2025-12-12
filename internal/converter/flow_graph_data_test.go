package converter

import (
	"strings"
	"testing"
)

// Phase 4a: Validate graph_data pan, renderer, and boolean flags serialization
func TestGraphData_PanRendererAndFlags(t *testing.T) {
	flowData := map[string]interface{}{
		"name": "test-flow",
		"graphData": map[string]interface{}{
			"elements": map[string]interface{}{
				"nodes": []interface{}{},
				"edges": []interface{}{},
			},
			"pan": map[string]interface{}{
				"x": 10.5,
				"y": -20.25,
			},
			"zoom":                2.0,
			"minZoom":             0.5,
			"maxZoom":             3.25,
			"zoomingEnabled":      true,
			"panningEnabled":      false,
			"userZoomingEnabled":  true,
			"userPanningEnabled":  false,
			"boxSelectionEnabled": true,
			"renderer": map[string]interface{}{
				"name":    "canvas",
				"version": "1",
			},
		},
	}

	hcl, err := ConvertFlowToHCL(flowData, "var.pingone_environment_id", false, nil)
	if err != nil {
		t.Fatalf("ConvertFlowToHCL returned error: %v", err)
	}

	checks := []string{
		"graph_data = {",
		"pan = {",
		"x = 10.5",
		"y = -20.25",
		"zoom                  = 2",
		"min_zoom              = 0.5",
		"max_zoom              = 3.25",
		"zooming_enabled       = true",
		"panning_enabled       = false",
		"user_zooming_enabled  = true",
		"user_panning_enabled  = false",
		"box_selection_enabled = true",
		"renderer = jsonencode(",
		"\"name\"",
		"\"canvas\"",
	}

	for _, c := range checks {
		if !strings.Contains(hcl, c) {
			t.Errorf("HCL missing expected fragment: %s\nHCL:\n%s", c, hcl)
		}
	}
}
