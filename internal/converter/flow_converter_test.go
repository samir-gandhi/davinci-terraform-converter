package converter

import (
	"strings"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{
			name:     "simple camelCase",
			input:    "httpConnector",
			expected: "httpconnector",
		},
		{
			name:     "single word lowercase",
			input:    "connector",
			expected: "connector",
		},
		{
			name:     "single word uppercase",
			input:    "Connector",
			expected: "connector",
		},

		// Consecutive capitals (kept together, just lowercased)
		{
			name:     "acronym at start",
			input:    "SSO",
			expected: "sso",
		},
		{
			name:     "acronym at end",
			input:    "pingOneSSO",
			expected: "pingonesso",
		},
		{
			name:     "acronym in middle",
			input:    "pingOneSSOConnector",
			expected: "pingonessoconnector",
		},
		{
			name:     "multiple acronyms",
			input:    "HTTPSSOAPIConnector",
			expected: "httpssoapiconnector",
		},
		{
			name:     "PingOne with acronym",
			input:    "PingOneAPIConnector",
			expected: "pingoneapiconnector",
		},

		// Real-world connector IDs (matches pingcli/dvtf-pingctl output)
		{
			name:     "PingOne SSO Connector",
			input:    "pingOneSSOConnector",
			expected: "pingonessoconnector",
		},
		{
			name:     "HTTP Connector",
			input:    "httpConnector",
			expected: "httpconnector",
		},
		{
			name:     "annotation connector",
			input:    "annotationConnector",
			expected: "annotationconnector",
		},
		{
			name:     "strings connector",
			input:    "stringsConnector",
			expected: "stringsconnector",
		},
		{
			name:     "variables connector",
			input:    "variablesConnector",
			expected: "variablesconnector",
		},

		// Special characters (removed, not typical in connector IDs)
		{
			name:     "with hyphens",
			input:    "ping-one-connector",
			expected: "pingoneconnector",
		},
		{
			name:     "with spaces",
			input:    "ping one connector",
			expected: "pingoneconnector",
		},
		{
			name:     "with dots",
			input:    "ping.one.connector",
			expected: "pingoneconnector",
		},
		{
			name:     "mixed special chars",
			input:    "ping-one.connector SSO",
			expected: "pingoneconnectorsso",
		},

		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "all uppercase",
			input:    "HTTP",
			expected: "http",
		},
		{
			name:     "all lowercase",
			input:    "connector",
			expected: "connector",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "single capital",
			input:    "A",
			expected: "a",
		},
		{
			name:     "underscores preserved",
			input:    "ping_one_connector",
			expected: "ping_one_connector",
		},

		// Complex real-world examples
		{
			name:     "PingOne MFA connector",
			input:    "pingOneMFAConnector",
			expected: "pingonemfaconnector",
		},
		{
			name:     "Azure AD connector",
			input:    "azureADConnector",
			expected: "azureadconnector",
		},
		{
			name:     "SAML connector",
			input:    "SAMLConnector",
			expected: "samlconnector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertFlowToHCL_EmitsColor(t *testing.T) {
	flow := map[string]interface{}{
		"name":        "Example Flow",
		"description": "Desc",
		// API provides flow color as flowColor
		"flowColor": "#ABCDEF",
	}

	hcl, err := ConvertFlowToHCL(flow, "var.environment_id", true, nil)
	if err != nil {
		t.Fatalf("ConvertFlowToHCL error: %v", err)
	}

	if !strings.Contains(hcl, "color       = \"#ABCDEF\"") {
		t.Fatalf("expected color to be emitted, got: %s", hcl)
	}
}

func TestConvertFlowToHCL_EmitsInputSchema(t *testing.T) {
	inputSchema := []interface{}{
		map[string]interface{}{
			"description":          "desc1",
			"isExpanded":           true,
			"preferredControlType": "textField",
			"preferredDataType":    "boolean",
			"propertyName":         "checkRequired",
			"required":             true,
		},
		map[string]interface{}{
			"isExpanded":           true,
			"preferredControlType": "textField",
			"preferredDataType":    "string",
			"propertyName":         "pingOneUserId",
			"required":             true,
			"description":          "",
		},
	}

	flow := map[string]interface{}{
		"name":        "Example Flow",
		"flowColor":   "#ABCDEF",
		"inputSchema": inputSchema,
	}

	hcl, err := ConvertFlowToHCL(flow, "var.environment_id", true, nil)
	if err != nil {
		t.Fatalf("ConvertFlowToHCL error: %v", err)
	}

	// Ensure input_schema block exists
	if !strings.Contains(hcl, "input_schema = [") {
		t.Fatalf("expected input_schema block, got: %s", hcl)
	}
}
