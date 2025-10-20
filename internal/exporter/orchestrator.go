package exporter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
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

	// Initialize dependency graph and missing dependency tracker
	graph := resolver.NewDependencyGraph()
	missingTracker := resolver.NewMissingDependencyTracker()

	// Track which resource types are included in this export
	includedTypes := []string{
		"pingone_davinci_variable",
		"pingone_davinci_connector_instance",
		"pingone_davinci_flow",
		"pingone_davinci_application",
		"pingone_davinci_application_flow_policy",
	}
	missingTracker.SetIncludedTypes(includedTypes)

	// Export resources in dependency order, building the graph as we go

	// 1. Variables
	variables, err := ExportVariables(ctx, client, skipDeps, graph)
	if err != nil {
		return "", fmt.Errorf("failed to export variables: %w", err)
	}
	hcl.WriteString(variables)
	hcl.WriteString("\n")

	// 2. Connector Instances
	connectors, err := ExportConnectorInstances(ctx, client, skipDeps, graph)
	if err != nil {
		return "", fmt.Errorf("failed to export connector instances: %w", err)
	}
	hcl.WriteString(connectors)
	hcl.WriteString("\n")

	// 3. Flows
	flows, err := ExportFlows(ctx, client, skipDeps, graph)
	if err != nil {
		return "", fmt.Errorf("failed to export flows: %w", err)
	}
	hcl.WriteString(flows)
	hcl.WriteString("\n")

	// 4. Applications
	applications, err := ExportApplications(ctx, client, skipDeps, graph)
	if err != nil {
		return "", fmt.Errorf("failed to export applications: %w", err)
	}
	hcl.WriteString(applications)
	hcl.WriteString("\n")

	// 5. Flow Policies
	flowPolicies, err := ExportFlowPolicies(ctx, client, skipDeps, graph)
	if err != nil {
		return "", fmt.Errorf("failed to export flow policies: %w", err)
	}
	hcl.WriteString(flowPolicies)

	// Get the final HCL output
	finalHCL := hcl.String()

	// Validate dependency graph and print reports to stderr
	if err := graph.ValidateGraph(); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠ Warning: Dependency validation found issues:\n%v\n\n", err)
	}

	// Count TODO comments in generated HCL
	todoCount := strings.Count(finalHCL, "# TODO:")

	// Print validation report to stderr with TODO count
	report := graph.GenerateValidationReport()
	// Insert TODO count after Total Dependencies line
	if todoCount > 0 {
		lines := strings.Split(report, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "Total Dependencies:") {
				// Insert TODO count line after dependencies line
				newLines := append(lines[:i+1], append([]string{fmt.Sprintf("TODO Comments: %d", todoCount)}, lines[i+1:]...)...)
				report = strings.Join(newLines, "\n")
				break
			}
		}
	}
	fmt.Fprintln(os.Stderr, report)

	// Print missing dependencies summary to stderr if any
	if len(missingTracker.GetMissing()) > 0 {
		fmt.Fprintln(os.Stderr, missingTracker.GenerateSummaryReport())
	}

	return finalHCL, nil
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
