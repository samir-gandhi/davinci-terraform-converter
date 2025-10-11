// Copyright © 2025 Ping Identity Corporation
package utils_test

import (
	"testing"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

func TestSanitizeResourceName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple alphanumeric",
			input:    "Customer",
			expected: "pingcli__Customer",
		},
		{
			name:     "Alphanumeric with capitals",
			input:    "CustomerHTMLFormPF",
			expected: "pingcli__CustomerHTMLFormPF",
		},
		{
			name:     "Spaces and parentheses",
			input:    "Customer HTML Form (PF)",
			expected: "pingcli__Customer-0020-HTML-0020-Form-0020--0028-PF-0029-",
		},
		{
			name:     "Special characters",
			input:    "Customer@HTML#Form$PF%",
			expected: "pingcli__Customer-0040-HTML-0023-Form-0024-PF-0025-",
		},
		{
			name:     "Flow with spaces",
			input:    "My Registration Flow",
			expected: "pingcli__My-0020-Registration-0020-Flow",
		},
		{
			name:     "Flow with special chars",
			input:    "Login & Signup",
			expected: "pingcli__Login-0020--0026--0020-Signup",
		},
		{
			name:     "Underscore preserved",
			input:    "flow_name",
			expected: "pingcli__flow_name",
		},
		{
			name:     "Hyphen preserved",
			input:    "flow-name",
			expected: "pingcli__flow-name",
		},
		{
			name:     "Mixed case with numbers",
			input:    "Flow123Test",
			expected: "pingcli__Flow123Test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.SanitizeResourceName(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeResourceName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
