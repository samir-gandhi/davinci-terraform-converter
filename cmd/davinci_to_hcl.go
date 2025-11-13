// Copyright © 2025 Ping Identity Corporation

// Package cmd provides the command implementation for the DaVinci flow to HCL converter.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/spf13/pflag"
)

// Command metadata for the davinci-to-hcl subcommand
var (
	// DaVinciToHclExample provides usage examples for the command
	DaVinciToHclExample = `  # Convert a single DaVinci flow to HCL and output to stdout
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json

  # Convert a single flow to HCL and save to a file
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json --out ./output.tf

  # Convert multiple flows from separate files (comma-separated)
  pingcli tf davinci-to-hcl --flow-json ./flow1.json,./flow2.json --out ./flows.tf

  # Convert multiple flows from separate files (repeated flag)
  pingcli tf davinci-to-hcl --flow-json ./flow1.json --flow-json ./flow2.json --out ./flows.tf

  # Convert all flows from a directory
  pingcli tf davinci-to-hcl --flow-json-dir ./flows/ --out ./all-flows.tf

  # Convert without generating Terraform references (use hardcoded IDs)
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json --skip-dependencies`

	// DaVinciToHclLong provides a detailed description of the command
	DaVinciToHclLong = `Convert PingOne DaVinci flows from JSON to Terraform HCL.

Reads DaVinci flow exports (JSON format) and converts them to HCL syntax
compatible with the pingone_davinci_flow resource in the PingOne Terraform Provider.

Supports three input modes:
  1. Single file:     --flow-json file.json
  2. Multiple files:  --flow-json file1.json,file2.json (comma-separated)
                      --flow-json file1.json --flow-json file2.json (repeated flag)
  3. Directory:       --flow-json-dir /path/to/flows/ (all .json files)

When processing multiple flows, they are combined into a single HCL output with
proper separation between resources. Each flow becomes an individual Terraform
resource block.

Some flow JSON files may contain an array of flows. The tool will automatically
detect this and convert each flow into a separate Terraform HCL resource.

By default, generates Terraform references for dependencies. Use --skip-dependencies
to preserve raw UUIDs instead.`

	// DaVinciToHclShort provides a brief, one-line description of the command
	DaVinciToHclShort = "Convert DaVinci flow JSON file(s) to Terraform HCL"

	// DaVinciToHclUse defines the command's name and its arguments/flags syntax
	DaVinciToHclUse = "davinci-to-hcl --flow-json <file(s)> | --flow-json-dir <directory> [flags]"
)

// DaVinciToHclCommand is the implementation of the davinci-to-hcl subcommand.
// It encapsulates the logic for converting DaVinci flows to HCL.
type DaVinciToHclCommand struct{}

// A compile-time check to ensure DaVinciToHclCommand correctly implements the
// grpc.PingCliCommand interface.
var _ grpc.PingCliCommand = (*DaVinciToHclCommand)(nil)

// Configuration is called by the pingcli host to retrieve the command's
// metadata, such as its name, description, and usage examples.
func (c *DaVinciToHclCommand) Configuration() (*grpc.PingCliCommandConfiguration, error) {
	cmdConfig := &grpc.PingCliCommandConfiguration{
		Example: DaVinciToHclExample,
		Long:    DaVinciToHclLong,
		Short:   DaVinciToHclShort,
		Use:     DaVinciToHclUse,
	}

	return cmdConfig, nil
}

// Run is the execution entry point for the davinci-to-hcl subcommand.
// It parses flags and executes the conversion logic.
// Supports three input modes:
//  1. Single file: --flow-json file.json
//  2. Multiple files: --flow-json file1.json --flow-json file2.json OR --flow-json file1.json,file2.json
//  3. Directory: --flow-json-dir /path/to/flows/
func (c *DaVinciToHclCommand) Run(args []string, logger grpc.Logger) error {
	// Create a new FlagSet for parsing command-line flags
	flags := pflag.NewFlagSet("davinci-to-hcl", pflag.ContinueOnError)

	// Define file-based conversion flags
	flowJSONFiles := flags.StringArray("flow-json", []string{}, "Path to input DaVinci flow JSON file(s)")
	flowJSONDir := flags.String("flow-json-dir", "", "Directory containing DaVinci flow JSON files")
	out := flags.String("out", "", "Path to the output HCL file (optional, defaults to stdout)")
	skipDependencies := flags.Bool("skip-dependencies", false, "Skip generating Terraform references and use hardcoded IDs instead")

	// Parse the provided arguments
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Validate input mode: either --flow-json OR --flow-json-dir (not both)
	hasFiles := len(*flowJSONFiles) > 0
	hasDir := *flowJSONDir != ""

	if hasFiles && hasDir {
		return fmt.Errorf("cannot specify both --flow-json and --flow-json-dir flags")
	}

	if !hasFiles && !hasDir {
		return fmt.Errorf("either --flow-json or --flow-json-dir flag is required")
	}

	// Collect all flow file paths based on input mode
	var allFlowFiles []string
	var err error

	if hasDir {
		// Directory mode: discover all JSON files in directory
		allFlowFiles, err = discoverFlowFilesInDirectory(*flowJSONDir)
		if err != nil {
			return fmt.Errorf("failed to discover flow files in directory: %w", err)
		}
		if len(allFlowFiles) == 0 {
			return fmt.Errorf("no JSON files found in directory: %s", *flowJSONDir)
		}
	} else {
		// File mode: parse all --flow-json inputs (handles comma-separated and repeated flags)
		for _, fileSpec := range *flowJSONFiles {
			files := strings.Split(fileSpec, ",")
			for _, file := range files {
				trimmed := strings.TrimSpace(file)
				if trimmed == "" {
					return fmt.Errorf("empty file path in --flow-json")
				}
				allFlowFiles = append(allFlowFiles, trimmed)
			}
		}
	}

	// Execute batch conversion
	return c.runConvert(logger, allFlowFiles, *out, *skipDependencies)
}

// runConvert handles conversion from local JSON file(s).
// Processes multiple flows and combines HCL output into a single file or stdout.
func (c *DaVinciToHclCommand) runConvert(logger grpc.Logger, flowFiles []string, out string, skipDeps bool) error {
	// Validate input
	if len(flowFiles) == 0 {
		return fmt.Errorf("no flow files provided")
	}

	// Log conversion start
	message := fmt.Sprintf("Executing DaVinci flow conversion for %d file(s)", len(flowFiles))
	if out != "" {
		message += fmt.Sprintf("\nOutput will be written to: %s", out)
	} else {
		message += "\nOutput will be written to stdout"
	}
	if skipDeps {
		message += "\nDependency references will be skipped (hardcoded IDs will be used)"
	}

	if err := logger.Message(message, nil); err != nil {
		return err
	}

	// Collect all HCL output from all flows
	var allHCL strings.Builder

	// Process each flow file
	for i, flowJSON := range flowFiles {
		// Log progress
		progressMsg := fmt.Sprintf("Processing flow %d/%d: %s", i+1, len(flowFiles), flowJSON)
		if err := logger.Message(progressMsg, nil); err != nil {
			return err
		}

		// Read the flow JSON file
		flowJSONBytes, err := os.ReadFile(flowJSON)
		if err != nil {
			if logErr := logger.PluginError("Failed to read flow JSON file", map[string]string{
				"file":  flowJSON,
				"error": err.Error(),
			}); logErr != nil {
				return fmt.Errorf("failed to log error: %w", logErr)
			}
			return fmt.Errorf("failed to read flow JSON file %s: %w", flowJSON, err)
		}

		// Convert the flow JSON to HCL
		hcl, err := converter.ConvertWithOptions(flowJSONBytes, skipDeps)
		if err != nil {
			if logErr := logger.PluginError("Failed to convert flow JSON to HCL", map[string]string{
				"file":  flowJSON,
				"error": err.Error(),
			}); logErr != nil {
				return fmt.Errorf("failed to log error: %w", logErr)
			}
			return fmt.Errorf("failed to convert flow JSON to HCL for %s: %w", flowJSON, err)
		}

		// Add HCL to combined output
		allHCL.WriteString(hcl)

		// Add separator between flows (except after the last one)
		if i < len(flowFiles)-1 {
			allHCL.WriteString("\n\n")
		}
	}

	// Write combined output
	combinedHCL := allHCL.String()
	if out != "" {
		if err := os.WriteFile(out, []byte(combinedHCL), 0644); err != nil {
			if logErr := logger.PluginError("Failed to write output file", map[string]string{
				"file":  out,
				"error": err.Error(),
			}); logErr != nil {
				return fmt.Errorf("failed to log error: %w", logErr)
			}
			return fmt.Errorf("failed to write output file: %w", err)
		}
		successMsg := fmt.Sprintf("Successfully converted %d flow(s) to HCL: %s", len(flowFiles), out)
		if err := logger.Success(successMsg, nil); err != nil {
			return err
		}
	} else {
		// Output to stdout
		fmt.Println(combinedHCL)
	}

	return nil
}

// parseFlowFiles parses the --flow-json flag(s) to extract all flow file paths.
// Supports multiple files via:
//   - Comma-separated values: --flow-json file1.json,file2.json
//   - Repeated flags: --flow-json file1.json --flow-json file2.json
//   - Mixed approach: both comma-separated and repeated flags
func parseFlowFiles(args []string) ([]string, error) {
	flags := pflag.NewFlagSet("davinci-to-hcl", pflag.ContinueOnError)

	// Use StringArray to support repeated --flow-json flags
	flowJSONFiles := flags.StringArray("flow-json", []string{}, "Path to DaVinci flow JSON file(s)")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// Check if any files were specified
	if len(*flowJSONFiles) == 0 {
		return nil, fmt.Errorf("--flow-json flag is required")
	}

	// Collect all files, splitting comma-separated values
	var allFiles []string
	for _, fileSpec := range *flowJSONFiles {
		// Handle comma-separated file paths
		files := strings.Split(fileSpec, ",")
		for _, file := range files {
			trimmed := strings.TrimSpace(file)
			if trimmed == "" {
				return nil, fmt.Errorf("empty file path in --flow-json")
			}
			allFiles = append(allFiles, trimmed)
		}
	}

	return allFiles, nil
}

// parseDirectoryInput parses the --flow-json-dir flag for directory-based input.
// Returns the directory path or empty string if not specified.
// Returns error if both --flow-json and --flow-json-dir are specified.
func parseDirectoryInput(args []string) (string, error) {
	flags := pflag.NewFlagSet("davinci-to-hcl", pflag.ContinueOnError)

	flowJSONDir := flags.String("flow-json-dir", "", "Directory containing DaVinci flow JSON files")
	flowJSON := flags.StringArray("flow-json", []string{}, "Path to DaVinci flow JSON file(s)")

	if err := flags.Parse(args); err != nil {
		return "", err
	}

	// Check for conflicting flags
	if *flowJSONDir != "" && len(*flowJSON) > 0 {
		return "", fmt.Errorf("cannot specify both --flow-json and --flow-json-dir flags")
	}

	return *flowJSONDir, nil
}

// discoverFlowFilesInDirectory scans a directory for JSON files and returns their paths.
// Only returns files with .json extension (case-insensitive).
// Does not recurse into subdirectories.
func discoverFlowFilesInDirectory(dir string) ([]string, error) {
	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var jsonFiles []string
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Check for .json extension (case-insensitive)
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".json") {
			// Build full path
			fullPath := dir + string(os.PathSeparator) + name
			jsonFiles = append(jsonFiles, fullPath)
		}
	}

	return jsonFiles, nil
}

// parseArgs is a helper function to parse the args slice into a map of flags.
// This is useful for testing without having to set up the full pflag parsing.
func parseArgs(args []string) (flowJSON string, out string, skipDeps bool, err error) {
	flags := pflag.NewFlagSet("davinci-to-hcl", pflag.ContinueOnError)
	flowJSONPtr := flags.String("flow-json", "", "Path to the input DaVinci flow JSON file (required)")
	outPtr := flags.String("out", "", "Path to the output HCL file (optional, defaults to stdout)")
	skipDepsPtr := flags.Bool("skip-dependencies", false, "Skip generating Terraform references and use hardcoded IDs instead")

	if err := flags.Parse(args); err != nil {
		return "", "", false, err
	}

	if *flowJSONPtr == "" {
		return "", "", false, fmt.Errorf("--flow-json flag is required")
	}

	return *flowJSONPtr, *outPtr, *skipDepsPtr, nil
}

// hasFlag checks if a flag with the given name exists in the args slice
func hasFlag(args []string, flagName string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--"+flagName) || strings.HasPrefix(arg, "-"+flagName) {
			return true
		}
	}
	return false
}
