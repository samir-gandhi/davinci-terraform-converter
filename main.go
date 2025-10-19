// Copyright © 2025 Ping Identity Corporation

// Package main provides a CLI plugin for converting PingOne DaVinci flows
// (in JSON format) to HCL (HashiCorp Configuration Language) that is compatible
// with the PingOne Terraform Provider's DaVinci resources.
//
// This binary can operate in two modes:
// 1. Plugin mode: Launched by pingcli as a gRPC plugin
// 2. Standalone mode: Run directly from command line with flags
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/samir-gandhi/davinci-terraform-converter/cmd"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/exporter"
	"github.com/spf13/pflag"
)

// Version information - set at build time via ldflags or goreleaser
var (
	version = "dev"
	commit  = "dev"
)

// main is the entry point for the binary. It detects whether to run in
// plugin mode or standalone CLI mode based on the environment and arguments.
func main() {
	// Try to get the commit hash from the build info if it wasn't set at build time
	if commit == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					commit = setting.Value
					break
				}
			}
		}
	}

	// Check if we're being launched as a plugin
	// Plugins are invoked with specific environment variables set by go-plugin
	if os.Getenv("PLUGIN_PROTOCOL_VERSIONS") != "" {
		runAsPlugin()
		return
	}

	// Otherwise, run as standalone CLI
	runAsStandalone()
}

// runAsPlugin starts the gRPC plugin server for pingcli integration
func runAsPlugin() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: grpc.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			grpc.ENUM_PINGCLI_COMMAND_GRPC: &grpc.PingCliCommandGrpcPlugin{
				Impl: &cmd.DaVinciConvertCommand{},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// runAsStandalone provides a standalone CLI interface
func runAsStandalone() {
	flags := pflag.NewFlagSet("davinci-convert", pflag.ExitOnError)

	// File-based conversion flags
	flowJSON := flags.StringP("flow-json", "f", "", "Path to input DaVinci flow JSON file")
	out := flags.StringP("out", "o", "", "Path to output HCL file (optional, defaults to stdout)")
	outDir := flags.StringP("out-dir", "d", "", "Directory for multi-flow output (optional, for multi-flow exports)")
	skipDependencies := flags.Bool("skip-dependencies", false, "Skip generating Terraform references and use hardcoded IDs instead")

	// API export flags
	export := flags.Bool("export", false, "Enable API export mode to export all resources from an environment")
	environmentID := flags.String("environment-id", "", "PingOne environment ID for API export (or use PINGONE_ENVIRONMENT_ID env var)")
	region := flags.String("region", "", "PingOne region: NA, EU, AP, or CA (or use PINGONE_REGION env var)")
	clientID := flags.String("client-id", "", "OAuth client ID for authentication (or use PINGONE_CLIENT_ID env var)")
	clientSecret := flags.String("client-secret", "", "OAuth client secret for authentication (or use PINGONE_CLIENT_SECRET env var)")

	help := flags.BoolP("help", "h", false, "Show help message")
	showVersion := flags.BoolP("version", "v", false, "Show version information")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "DaVinci Flow to HCL Converter\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nFile-based Conversion Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Convert single flow to stdout\n")
		fmt.Fprintf(os.Stderr, "  %s --flow-json flow.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Convert single flow to file\n")
		fmt.Fprintf(os.Stderr, "  %s --flow-json flow.json --out output.tf\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Convert multi-flow export to directory\n")
		fmt.Fprintf(os.Stderr, "  %s --flow-json multiflow.json --out-dir ./flows\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nAPI Export Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Export all resources from environment\n")
		fmt.Fprintf(os.Stderr, "  %s --export --environment-id <uuid> --region NA --client-id <id> --client-secret <secret> --out environment.tf\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Export with environment variables\n")
		fmt.Fprintf(os.Stderr, "  PINGONE_ENVIRONMENT_ID=<uuid> PINGONE_CLIENT_ID=<id> PINGONE_CLIENT_SECRET=<secret> %s --export --out environment.tf\n\n", os.Args[0])
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *help {
		flags.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("davinci-convert version %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	// Determine which mode to use
	if *export {
		handleExportMode(*environmentID, *region, *clientID, *clientSecret, *out, *skipDependencies)
		return
	}

	// File-based conversion mode
	if *flowJSON == "" {
		fmt.Fprintf(os.Stderr, "Error: --flow-json flag is required for file-based conversion\n\n")
		flags.Usage()
		os.Exit(1)
	}

	// Read input file
	fileData, err := os.ReadFile(*flowJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Detect if this is a multi-flow export or single flow
	var check map[string]interface{}
	if err := json.Unmarshal(fileData, &check); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Check for "flows" array (multi-flow export)
	if _, hasFlows := check["flows"]; hasFlows {
		handleMultiFlow(fileData, *out, *outDir, *flowJSON, *skipDependencies)
	} else {
		handleSingleFlow(fileData, *out, *skipDependencies)
	}
}

// handleExportMode handles API export of all resources from an environment
func handleExportMode(environmentID, region, clientID, clientSecret, outPath string, skipDependencies bool) {
	// Get credentials from environment variables if not provided via flags
	if environmentID == "" {
		environmentID = os.Getenv("PINGONE_ENVIRONMENT_ID")
	}
	if region == "" {
		region = os.Getenv("PINGONE_REGION")
	}
	if clientID == "" {
		clientID = os.Getenv("PINGONE_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("PINGONE_CLIENT_SECRET")
	}

	// Validate required credentials
	if environmentID == "" {
		fmt.Fprintf(os.Stderr, "Error: environment ID is required - use --environment-id flag or PINGONE_ENVIRONMENT_ID env var\n")
		os.Exit(1)
	}
	if clientID == "" {
		fmt.Fprintf(os.Stderr, "Error: client ID is required - use --client-id flag or PINGONE_CLIENT_ID env var\n")
		os.Exit(1)
	}
	if clientSecret == "" {
		fmt.Fprintf(os.Stderr, "Error: client secret is required - use --client-secret flag or PINGONE_CLIENT_SECRET env var\n")
		os.Exit(1)
	}

	// Default region to NA if not specified
	if region == "" {
		region = "NA"
	}

	fmt.Fprintf(os.Stderr, "Exporting DaVinci environment: %s (Region: %s)\n", environmentID, region)

	// Create API client
	ctx := context.Background()
	client, err := api.NewClientSingleEnvironment(ctx, environmentID, region, clientID, clientSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
		os.Exit(1)
	}

	// Export all resources
	hcl, err := exporter.ExportEnvironment(ctx, client, skipDependencies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting environment: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if outPath != "" {
		if err := os.WriteFile(outPath, []byte(hcl), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully exported environment to: %s (%d bytes)\n", outPath, len(hcl))
	} else {
		// Write to stdout
		fmt.Print(hcl)
	}
}

// handleSingleFlow processes a single flow export
func handleSingleFlow(fileData []byte, outPath string, skipDependencies bool) {
	hcl, err := converter.ConvertWithOptions(fileData, skipDependencies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting flow: %v\n", err)
		os.Exit(1)
	}

	if outPath != "" {
		if err := os.WriteFile(outPath, []byte(hcl), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully converted flow to: %s\n", outPath)
	} else {
		fmt.Println(hcl)
	}
}

// handleMultiFlow processes a multi-flow export
func handleMultiFlow(fileData []byte, outPath, outDir, inputFile string, skipDependencies bool) {
	results, err := converter.ConvertMultiFlowWithOptions(fileData, skipDependencies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting multi-flow: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: No flows found in export\n")
		os.Exit(0)
	}

	// If output directory specified, write each flow to separate file
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		for i, hcl := range results {
			// Extract flow name from HCL for filename
			flowName := extractFlowName(hcl, i)
			filename := filepath.Join(outDir, fmt.Sprintf("%s.tf", flowName))

			if err := os.WriteFile(filename, []byte(hcl), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing flow %d: %v\n", i+1, err)
				os.Exit(1)
			}
			fmt.Printf("Flow %d written to: %s\n", i+1, filename)
		}
		fmt.Printf("\nSuccessfully converted %d flows to: %s\n", len(results), outDir)
	} else if outPath != "" {
		// Single output file - concatenate all flows
		combined := strings.Join(results, "\n\n")
		if err := os.WriteFile(outPath, []byte(combined), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully converted %d flows to: %s\n", len(results), outPath)
	} else {
		// Output to stdout - concatenate all flows
		combined := strings.Join(results, "\n\n")
		fmt.Println(combined)
	}
}

// extractFlowName extracts the resource name from HCL output for use as filename
func extractFlowName(hcl string, index int) string {
	// Look for: resource "pingone_davinci_flow" "name_here"
	lines := strings.Split(hcl, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "resource \"pingone_davinci_flow\"") {
			// Extract resource name
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				return parts[3]
			}
		}
	}
	// Fallback to index-based name
	return fmt.Sprintf("flow_%d", index+1)
}
