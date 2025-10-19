package exporter

import (
	"context"
	"fmt"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
)

// ExportEnvironment exports all DaVinci resources from an environment in dependency order
// Returns complete Terraform configuration including provider setup and all resources
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool) (string, error) {
	var hcl strings.Builder

	// Add header comment
	hcl.WriteString("# DaVinci Environment Export\n")
	hcl.WriteString(fmt.Sprintf("# Environment ID: %s\n", client.EnvironmentID))
	hcl.WriteString(fmt.Sprintf("# Region: %s\n", client.Region))
	hcl.WriteString("#\n")
	hcl.WriteString("# Exported resources in dependency order:\n")
	hcl.WriteString("# 1. Variables (no dependencies)\n")
	hcl.WriteString("# 2. Connector Instances (no dependencies)\n")
	hcl.WriteString("# 3. Flows (depends on connectors)\n")
	hcl.WriteString("# 4. Applications (depends on flows)\n")
	hcl.WriteString("# 5. Flow Policies (depends on applications and flows)\n")
	hcl.WriteString("\n")

	// Add provider configuration
	if !skipDeps {
		hcl.WriteString(generateProviderConfig(client.Region))
		hcl.WriteString("\n")
		hcl.WriteString(generateVariableConfig())
		hcl.WriteString("\n")
	}

	// Export resources in dependency order

	// 1. Variables
	variables, err := ExportVariables(ctx, client, skipDeps)
	if err != nil {
		return "", fmt.Errorf("failed to export variables: %w", err)
	}
	hcl.WriteString(variables)
	hcl.WriteString("\n")

	// 2. Connector Instances
	connectors, err := ExportConnectorInstances(ctx, client, skipDeps)
	if err != nil {
		return "", fmt.Errorf("failed to export connector instances: %w", err)
	}
	hcl.WriteString(connectors)
	hcl.WriteString("\n")

	// 3. Flows
	flows, err := ExportFlows(ctx, client, skipDeps)
	if err != nil {
		return "", fmt.Errorf("failed to export flows: %w", err)
	}
	hcl.WriteString(flows)
	hcl.WriteString("\n")

	// 4. Applications
	applications, err := ExportApplications(ctx, client, skipDeps)
	if err != nil {
		return "", fmt.Errorf("failed to export applications: %w", err)
	}
	hcl.WriteString(applications)
	hcl.WriteString("\n")

	// 5. Flow Policies
	flowPolicies, err := ExportFlowPolicies(ctx, client, skipDeps)
	if err != nil {
		return "", fmt.Errorf("failed to export flow policies: %w", err)
	}
	hcl.WriteString(flowPolicies)

	return hcl.String(), nil
}

// generateProviderConfig generates the Terraform provider configuration block
func generateProviderConfig(region string) string {
	var hcl strings.Builder

	hcl.WriteString("terraform {\n")
	hcl.WriteString("  required_providers {\n")
	hcl.WriteString("    pingone = {\n")
	hcl.WriteString("      source  = \"pingidentity/pingone\"\n")
	hcl.WriteString("      version = \">= 1.0.0\"\n")
	hcl.WriteString("    }\n")
	hcl.WriteString("  }\n")
	hcl.WriteString("}\n")
	hcl.WriteString("\n")
	hcl.WriteString("provider \"pingone\" {\n")
	hcl.WriteString(fmt.Sprintf("  region = %q\n", region))
	hcl.WriteString("  # Configure authentication via environment variables:\n")
	hcl.WriteString("  # PINGONE_CLIENT_ID\n")
	hcl.WriteString("  # PINGONE_CLIENT_SECRET\n")
	hcl.WriteString("  # PINGONE_ENVIRONMENT_ID (for OAuth client)\n")
	hcl.WriteString("}\n")

	return hcl.String()
}

// generateVariableConfig generates the environment_id variable declaration
func generateVariableConfig() string {
	var hcl strings.Builder

	hcl.WriteString("variable \"environment_id\" {\n")
	hcl.WriteString("  description = \"PingOne environment ID for DaVinci resources\"\n")
	hcl.WriteString("  type        = string\n")
	hcl.WriteString("}\n")

	return hcl.String()
}
