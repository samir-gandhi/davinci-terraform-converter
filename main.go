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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/samir-gandhi/davinci-terraform-converter/cmd"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/spf13/pflag"
)

// main is the entry point for the binary. It detects whether to run in
// plugin mode or standalone CLI mode based on the environment and arguments.
func main() {
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

	flowJSON := flags.StringP("flow-json", "f", "", "Path to input DaVinci flow JSON file (required)")
	out := flags.StringP("out", "o", "", "Path to output HCL file (optional, defaults to stdout)")
	outDir := flags.StringP("out-dir", "d", "", "Directory for multi-flow output (optional, for multi-flow exports)")
	help := flags.BoolP("help", "h", false, "Show help message")
	version := flags.BoolP("version", "v", false, "Show version information")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "DaVinci Flow to HCL Converter\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Convert single flow to stdout\n")
		fmt.Fprintf(os.Stderr, "  %s --flow-json flow.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Convert single flow to file\n")
		fmt.Fprintf(os.Stderr, "  %s --flow-json flow.json --out output.tf\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Convert multi-flow export to directory\n")
		fmt.Fprintf(os.Stderr, "  %s --flow-json multiflow.json --out-dir ./flows\n\n", os.Args[0])
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *help {
		flags.Usage()
		os.Exit(0)
	}

	if *version {
		fmt.Println("davinci-convert v0.1.0")
		os.Exit(0)
	}

	if *flowJSON == "" {
		fmt.Fprintf(os.Stderr, "Error: --flow-json flag is required\n\n")
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
		handleMultiFlow(fileData, *out, *outDir, *flowJSON)
	} else {
		handleSingleFlow(fileData, *out)
	}
}

// handleSingleFlow processes a single flow export
func handleSingleFlow(fileData []byte, outPath string) {
	hcl, err := converter.Convert(fileData)
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
func handleMultiFlow(fileData []byte, outPath, outDir, inputFile string) {
	results, err := converter.ConvertMultiFlow(fileData)
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
