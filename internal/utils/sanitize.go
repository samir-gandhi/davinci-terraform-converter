// Copyright © 2025 Ping Identity Corporation

package utils

import (
	"fmt"
	"regexp"
)

// SanitizeResourceName converts a resource name to a valid Terraform resource name
// using the same logic as pingcli's ImportBlock.Sanitize() method.
// This ensures consistency between the converter and pingcli export functionality.
//
// The sanitization process:
// 1. Hexadecimal encodes special characters (anything not alphanumeric, underscore, or hyphen)
// 2. Prefixes the name with "pingcli__"
//
// Examples:
//   - "Customer" -> "pingcli__Customer"
//   - "Customer HTML Form (PF)" -> "pingcli__Customer-0020-HTML-0020-Form-0020--0028-PF-0029-"
//   - "Customer@HTML#Form$PF%" -> "pingcli__Customer-0040-HTML-0023-Form-0024-PF-0025-"
func SanitizeResourceName(name string) string {
	// Hexadecimal encode special characters
	name = regexp.MustCompile(`[^0-9A-Za-z_\-]`).ReplaceAllStringFunc(name, func(s string) string {
		return fmt.Sprintf("-%04X-", s)
	})
	// Prefix resource names with pingcli__
	return "pingcli__" + name
}

// CamelCaseToWords converts a camelCase or PascalCase string to space-separated words
// Examples:
//   - "clientSecret" -> "client secret"
//   - "apiKey" -> "api key"
//   - "envId" -> "env id"
func CamelCaseToWords(s string) string {
	// Insert space before uppercase letters (except at start)
	result := regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(s, "$1 $2")
	return result
}
