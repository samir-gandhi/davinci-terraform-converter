package exporter

import (
	"context"
	"fmt"

	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/importgen"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/module"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

// ExportedData contains structured export data for module generation
type ExportedData struct {
	// HCL sections by resource type
	VariablesHCL    string
	ConnectorsHCL   string
	FlowsHCL        string
	ApplicationsHCL string
	FlowPoliciesHCL string

	// Metadata
	EnvironmentID   string
	Region          string
	DependencyGraph *resolver.DependencyGraph
}

// ExportEnvironmentForModule exports DaVinci resources in a structure suitable for module generation
func ExportEnvironmentForModule(ctx context.Context, client *api.Client, opts ExportOptions, logger grpc.Logger) (*ExportedData, error) {
	data := &ExportedData{
		EnvironmentID: client.EnvironmentID,
		Region:        client.Region,
	}

	// Initialize import block generator if needed
	var importGen *importgen.ImportBlockGenerator
	if opts.GenerateImports {
		importGen = importgen.NewImportBlockGenerator()
	}

	// Initialize dependency graph
	graph := resolver.NewDependencyGraph()
	data.DependencyGraph = graph

	// Track which resource types are included
	missingTracker := resolver.NewMissingDependencyTracker()
	includedTypes := []string{
		"pingone_davinci_variable",
		"pingone_davinci_connector_instance",
		"pingone_davinci_flow",
		"pingone_davinci_application",
		"pingone_davinci_application_flow_policy",
	}
	missingTracker.SetIncludedTypes(includedTypes)

	// Log export start
	if err := logger.Message("Exporting DaVinci resources for module generation...", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// Export each resource type

	// 1. Variables
	if err := logger.Message("Fetching variables...", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}
	variables, err := ExportVariablesWithImports(ctx, client, opts.SkipDependencies, graph, importGen)
	if err != nil {
		return nil, fmt.Errorf("failed to export variables: %w", err)
	}
	data.VariablesHCL = variables
	if err := logger.Message("✓ Variables exported", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// 2. Connector Instances
	if err := logger.Message("Fetching connector instances...", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}
	connectors, err := ExportConnectorInstancesWithImports(ctx, client, opts.SkipDependencies, graph, importGen)
	if err != nil {
		return nil, fmt.Errorf("failed to export connector instances: %w", err)
	}
	data.ConnectorsHCL = connectors
	if err := logger.Message("✓ Connector instances exported", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// 3. Flows
	if err := logger.Message("Fetching flows...", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}
	flows, err := ExportFlowsWithImports(ctx, client, opts.SkipDependencies, graph, importGen)
	if err != nil {
		return nil, fmt.Errorf("failed to export flows: %w", err)
	}
	data.FlowsHCL = flows
	if err := logger.Message("✓ Flows exported", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// 4. Applications
	if err := logger.Message("Fetching applications...", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}
	applications, err := ExportApplicationsWithImports(ctx, client, opts.SkipDependencies, graph, importGen)
	if err != nil {
		return nil, fmt.Errorf("failed to export applications: %w", err)
	}
	data.ApplicationsHCL = applications
	if err := logger.Message("✓ Applications exported", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// 5. Flow Policies
	if err := logger.Message("Fetching flow policies...", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}
	flowPolicies, err := ExportFlowPoliciesWithImports(ctx, client, opts.SkipDependencies, graph, importGen)
	if err != nil {
		return nil, fmt.Errorf("failed to export flow policies: %w", err)
	}
	data.FlowPoliciesHCL = flowPolicies
	if err := logger.Message("✓ Flow policies exported", nil); err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// Validate dependency graph
	if err := graph.ValidateGraph(); err != nil {
		if warnErr := logger.Warn(fmt.Sprintf("Dependency validation found issues: %v", err), nil); warnErr != nil {
			return nil, fmt.Errorf("failed to log warning: %w", warnErr)
		}
	}

	return data, nil
}

// ConvertExportedDataToModuleStructure converts ExportedData to module.ModuleStructure
func ConvertExportedDataToModuleStructure(data *ExportedData, config module.ModuleConfig) (*module.ModuleStructure, error) {
	structure := &module.ModuleStructure{
		Config: config,
		Resources: module.ModuleResources{
			FlowsHCL:        data.FlowsHCL,
			ConnectionsHCL:  data.ConnectorsHCL,
			VariablesHCL:    data.VariablesHCL,
			ApplicationsHCL: data.ApplicationsHCL,
			FlowPoliciesHCL: data.FlowPoliciesHCL,
		},
	}

	// Generate variables from dependency graph
	// For now, we'll create basic variables
	// TODO: Extract actual variables from resources once converters support it
	variables := generateVariablesFromGraph(data.DependencyGraph)
	structure.Variables = variables

	// Generate outputs from dependency graph
	outputs := generateOutputsFromGraph(data.DependencyGraph)
	structure.Outputs = outputs

	// TODO: Import blocks support - will need to parse from HCL or track separately

	return structure, nil
}

// generateVariablesFromGraph extracts variables from the dependency graph
func generateVariablesFromGraph(graph *resolver.DependencyGraph) []module.Variable {
	variables := []module.Variable{}

	// TODO: In phase 2, iterate through graph resources and extract variable-eligible attributes
	// For now, return empty list as resources don't yet expose their attributes

	return variables
}

// generateOutputsFromGraph generates output definitions from the dependency graph
func generateOutputsFromGraph(graph *resolver.DependencyGraph) []module.Output {
	outputs := []module.Output{}

	// TODO: Extract outputs from dependency graph
	// For now, return empty list - will implement in phase 2

	return outputs
}
