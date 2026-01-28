package converter

import (
	"strings"
	"testing"
)

func TestFix20_MapStructure(t *testing.T) {
	flowData := map[string]interface{}{
		"name": "Map Structure Test Flow",
		"graphData": map[string]interface{}{
			"elements": map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"data": map[string]interface{}{
							"id":       "node-1",
							"nodeType": "CONNECTION",
						},
					},
					map[string]interface{}{
						"data": map[string]interface{}{
							"id":       "node-2",
							"nodeType": "CONNECTION",
						},
					},
				},
				"edges": []interface{}{
					map[string]interface{}{
						"data": map[string]interface{}{
							"id":     "edge-1",
							"source": "node-1",
							"target": "node-2",
						},
					},
				},
			},
		},
	}

	hcl, err := ConvertFlowToHCL(flowData, "var.environment_id", true, nil)
	if err != nil {
		t.Fatalf("ConvertFlowToHCL error: %v", err)
	}

	// Verify nodes map structure
	// Should see: nodes = {
	if !strings.Contains(hcl, "nodes = {") {
		t.Errorf("Expected 'nodes = {', got HCL:\n%s", hcl)
	}
	// Should verify map keys
	if !strings.Contains(hcl, "\"node-1\" = {") {
		t.Errorf("Expected '\"node-1\" = {', got HCL:\n%s", hcl)
	}
	if !strings.Contains(hcl, "\"node-2\" = {") {
		t.Errorf("Expected '\"node-2\" = {', got HCL:\n%s", hcl)
	}

	// Verify edges map structure
	if !strings.Contains(hcl, "edges = {") {
		t.Errorf("Expected 'edges = {', got HCL:\n%s", hcl)
	}
	if !strings.Contains(hcl, "\"edge-1\" = {") {
		t.Errorf("Expected '\"edge-1\" = {', got HCL:\n%s", hcl)
	}

	// Should NOT contain list syntax
	if strings.Contains(hcl, "nodes = [") {
		t.Errorf("HCL still contains 'nodes = [' list syntax")
	}
	if strings.Contains(hcl, "edges = [") {
		t.Errorf("HCL still contains 'edges = [' list syntax")
	}
}
