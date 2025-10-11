package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

// FlowPolicyResponse represents the DaVinci API response for an application flow policy
type FlowPolicyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Environment struct {
		ID string `json:"id"`
	} `json:"environment"`
	Application struct {
		ID string `json:"id"`
	} `json:"application"`
	Status            string             `json:"status"`
	FlowDistributions []FlowDistribution `json:"flowDistributions,omitempty"`
	Trigger           *FlowPolicyTrigger `json:"trigger,omitempty"`
}

// FlowDistribution represents a flow distribution in a flow policy
type FlowDistribution struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Weight  int    `json:"weight"`
}

// FlowPolicyTrigger represents the trigger configuration for a flow policy
type FlowPolicyTrigger struct {
	Type          string                 `json:"type,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

// ConvertFlowPolicy converts a DaVinci application flow policy JSON to Terraform HCL
func ConvertFlowPolicy(policyJSON []byte) (string, error) {
	return ConvertFlowPolicyWithOptions(policyJSON, false)
}

// ConvertFlowPolicyWithOptions converts a flow policy with optional skip-dependencies flag
func ConvertFlowPolicyWithOptions(policyJSON []byte, skipDependencies bool) (string, error) {
	var policy FlowPolicyResponse
	if err := json.Unmarshal(policyJSON, &policy); err != nil {
		return "", fmt.Errorf("failed to parse flow policy JSON: %w", err)
	}

	if policy.Name == "" {
		return "", fmt.Errorf("flow policy name is required")
	}

	if policy.Application.ID == "" {
		return "", fmt.Errorf("application.id is required")
	}

	return generateFlowPolicyHCL(policy, skipDependencies), nil
}

// generateFlowPolicyHCL generates the Terraform HCL for a flow policy
func generateFlowPolicyHCL(policy FlowPolicyResponse, skipDependencies bool) string {
	var hcl strings.Builder

	// Resource name using pingcli format
	resourceName := utils.SanitizeResourceName(policy.Name)
	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_application_flow_policy\" \"%s\" {\n", resourceName))

	// Environment ID
	if skipDependencies {
		hcl.WriteString(fmt.Sprintf("  environment_id = \"%s\"\n", policy.Environment.ID))
	} else {
		hcl.WriteString("  environment_id = var.environment_id\n")
	}

	hcl.WriteString("\n")

	// Required attributes
	hcl.WriteString(fmt.Sprintf("  name           = \"%s\"\n", policy.Name))
	hcl.WriteString(fmt.Sprintf("  status         = \"%s\"\n", policy.Status))

	hcl.WriteString("\n")

	// Application reference
	// Note: Always use hardcoded ID, no Terraform reference for application
	// This will be resolved in Part 4 if needed
	hcl.WriteString(fmt.Sprintf("  da_vinci_application_id = \"%s\"  # TODO: Replace with application reference\n", policy.Application.ID))

	// Flow distributions (set of objects)
	if len(policy.FlowDistributions) > 0 {
		hcl.WriteString("\n")
		writeFlowDistributionsBlock(&hcl, policy.FlowDistributions, skipDependencies)
	}

	// Trigger configuration (optional)
	if policy.Trigger != nil {
		hcl.WriteString("\n")
		writeFlowPolicyTriggerBlock(&hcl, policy.Trigger)
	}

	hcl.WriteString("}\n")

	return hcl.String()
}

// writeFlowDistributionsBlock writes the flow_distributions set
func writeFlowDistributionsBlock(hcl *strings.Builder, distributions []FlowDistribution, skipDependencies bool) {
	hcl.WriteString("  flow_distributions = [\n")

	for _, dist := range distributions {
		hcl.WriteString("    {\n")

		// Flow ID reference
		if skipDependencies {
			hcl.WriteString(fmt.Sprintf("      flow_id      = \"%s\"\n", dist.ID))
		} else {
			// TODO: Part 4 - resolve flow reference
			hcl.WriteString(fmt.Sprintf("      flow_id      = \"%s\"  # TODO: Replace with flow reference\n", dist.ID))
		}

		hcl.WriteString(fmt.Sprintf("      version      = %d\n", dist.Version))
		hcl.WriteString(fmt.Sprintf("      weight       = %d\n", dist.Weight))
		hcl.WriteString("    },\n")
	}

	hcl.WriteString("  ]\n")
}

// writeFlowPolicyTriggerBlock writes the trigger configuration block for flow policies
func writeFlowPolicyTriggerBlock(hcl *strings.Builder, trigger *FlowPolicyTrigger) {
	hcl.WriteString("  trigger = {\n")

	if trigger.Type != "" {
		hcl.WriteString(fmt.Sprintf("    type = \"%s\"\n", trigger.Type))
	}

	if len(trigger.Configuration) > 0 {
		hcl.WriteString("\n")
		hcl.WriteString("    configuration = {\n")

		// Write configuration key-value pairs
		for key, value := range trigger.Configuration {
			switch v := value.(type) {
			case string:
				hcl.WriteString(fmt.Sprintf("      %s = \"%s\"\n", key, v))
			case bool:
				hcl.WriteString(fmt.Sprintf("      %s = %t\n", key, v))
			case float64:
				if v == float64(int64(v)) {
					hcl.WriteString(fmt.Sprintf("      %s = %d\n", key, int64(v)))
				} else {
					hcl.WriteString(fmt.Sprintf("      %s = %f\n", key, v))
				}
			default:
				hcl.WriteString(fmt.Sprintf("      %s = \"%v\"\n", key, v))
			}
		}

		hcl.WriteString("    }\n")
	}

	hcl.WriteString("  }\n")
}
