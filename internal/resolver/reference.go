package resolver

import (
	"fmt"
)

// GenerateTerraformReference creates a Terraform reference string for a dependency
// Format: resource_type.resource_name.attribute
func GenerateTerraformReference(graph *DependencyGraph, resourceType, resourceID, attribute string) (string, error) {
	name, err := graph.GetReferenceName(resourceType, resourceID)
	if err != nil {
		return "", err
	}
	
	// Map internal resource types to Terraform resource types
	terraformType := mapToTerraformResourceType(resourceType)
	
	return fmt.Sprintf("%s.%s.%s", terraformType, name, attribute), nil
}

// GenerateTODOPlaceholder creates a TODO comment for a missing dependency
func GenerateTODOPlaceholder(resourceType, resourceID string, err error) string {
	return fmt.Sprintf(`"" # TODO: Reference to %s %s not found - %v`, resourceType, resourceID, err)
}

// mapToTerraformResourceType converts internal resource types to Terraform provider resource types
func mapToTerraformResourceType(internalType string) string {
	mapping := map[string]string{
		"flow":               "pingone_davinci_flow",
		"flow_policy":        "pingone_davinci_flow_policy",
		"connector_instance": "pingone_davinci_connector",
		"variable":           "pingone_davinci_variable",
		"application":        "pingone_davinci_application",
	}
	
	if terraformType, exists := mapping[internalType]; exists {
		return terraformType
	}
	
	// Fallback: prepend pingone_davinci_
	return "pingone_davinci_" + internalType
}
