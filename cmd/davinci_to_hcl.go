// Copyright © 2025 Ping Identity Corporation

// Package cmd provides the command implementation for the DaVinci flow to HCL converter.
package cmd

import (
	"fmt"
	"strings"

	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/spf13/pflag"
)

// Command metadata for the davinci-to-hcl subcommand
var (
	// DaVinciToHclExample provides usage examples for the command
	DaVinciToHclExample = `  # Convert a DaVinci flow to HCL and output to stdout
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json

  # Convert a DaVinci flow to HCL and save to a file
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json --out ./output.tf

  # Convert without generating Terraform references (use hardcoded IDs)
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json --skip-dependencies`

	// DaVinciToHclLong provides a detailed description of the command
	DaVinciToHclLong = `Convert a single PingOne DaVinci flow from JSON to Terraform HCL.

Reads a DaVinci flow export (JSON format) and converts it to HCL syntax
compatible with the pingone_davinci_flow resource in the PingOne Terraform Provider.

By default, generates Terraform references for dependencies. Use --skip-dependencies
to preserve raw UUIDs instead.`

	// DaVinciToHclShort provides a brief, one-line description of the command
	DaVinciToHclShort = "Convert a DaVinci flow JSON file to Terraform HCL"

	// DaVinciToHclUse defines the command's name and its arguments/flags syntax
	DaVinciToHclUse = "davinci-to-hcl --flow-json <file> [flags]"
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
func (c *DaVinciToHclCommand) Run(args []string, logger grpc.Logger) error {
	// Create a new FlagSet for parsing command-line flags
	flags := pflag.NewFlagSet("davinci-to-hcl", pflag.ContinueOnError)

	// Define file-based conversion flags
	flowJSON := flags.String("flow-json", "", "Path to the input DaVinci flow JSON file")
	out := flags.String("out", "", "Path to the output HCL file (optional, defaults to stdout)")
	skipDependencies := flags.Bool("skip-dependencies", false, "Skip generating Terraform references and use hardcoded IDs instead")

	// Parse the provided arguments
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Execute file-based conversion
	return c.runConvert(logger, *flowJSON, *out, *skipDependencies)
}

// runConvert handles conversion from a local JSON file
func (c *DaVinciToHclCommand) runConvert(logger grpc.Logger, flowJSON, out string, skipDeps bool) error {
	// Validate required flags
	if flowJSON == "" {
		return fmt.Errorf("--flow-json flag is required for file-based conversion")
	}

	// For now, just print a message to confirm the command structure is working
	message := fmt.Sprintf("Executing DaVinci flow conversion for file: %s", flowJSON)
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

	return nil
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
