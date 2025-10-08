// Copyright © 2025 Ping Identity Corporation

// Package cmd provides the command implementation for the DaVinci flow to HCL converter.
package cmd

import (
	"fmt"
	"strings"

	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/spf13/pflag"
)

// Command metadata for the DaVinci convert command
var (
	// Example provides usage examples for the command
	Example = `  # Convert a DaVinci flow to HCL and output to stdout
  pingcli davinci convert --flow-json ./my-flow.json

  # Convert a DaVinci flow to HCL and save to a file
  pingcli davinci convert --flow-json ./my-flow.json --out ./output.tf`

	// Long provides a detailed description of the command
	Long = `Convert a PingOne DaVinci flow from its native JSON format to HCL (HashiCorp Configuration Language).

The resulting HCL will be compatible with the DaVinci resources in the PingOne Terraform Provider.
Environment-specific values (such as connection IDs, variables, and subflows) will be converted 
to placeholder references that can be replaced with Terraform resource references.`

	// Short provides a brief, one-line description of the command
	Short = "Convert a DaVinci flow from JSON to HCL"

	// Use defines the command's name and its arguments/flags syntax
	Use = "convert"
)

// DaVinciConvertCommand is the implementation of the grpc.PingCliCommand interface.
// It encapsulates the logic for converting DaVinci flows to HCL.
type DaVinciConvertCommand struct{}

// A compile-time check to ensure DaVinciConvertCommand correctly implements the
// grpc.PingCliCommand interface.
var _ grpc.PingCliCommand = (*DaVinciConvertCommand)(nil)

// Configuration is called by the pingcli host to retrieve the command's
// metadata, such as its name, description, and usage examples.
func (c *DaVinciConvertCommand) Configuration() (*grpc.PingCliCommandConfiguration, error) {
	cmdConfig := &grpc.PingCliCommandConfiguration{
		Example: Example,
		Long:    Long,
		Short:   Short,
		Use:     Use,
	}

	return cmdConfig, nil
}

// Run is the execution entry point for the plugin command.
// It parses flags and executes the conversion logic.
func (c *DaVinciConvertCommand) Run(args []string, logger grpc.Logger) error {
	// Create a new FlagSet for parsing command-line flags
	flags := pflag.NewFlagSet("convert", pflag.ContinueOnError)

	// Define flags
	flowJSON := flags.String("flow-json", "", "Path to the input DaVinci flow JSON file (required)")
	out := flags.String("out", "", "Path to the output HCL file (optional, defaults to stdout)")

	// Parse the provided arguments
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *flowJSON == "" {
		return fmt.Errorf("--flow-json flag is required")
	}

	// For now, just print a message to confirm the command structure is working
	message := fmt.Sprintf("Executing DaVinci flow conversion for file: %s", *flowJSON)
	if *out != "" {
		message += fmt.Sprintf("\nOutput will be written to: %s", *out)
	} else {
		message += "\nOutput will be written to stdout"
	}

	if err := logger.Message(message, nil); err != nil {
		return err
	}

	return nil
}

// parseArgs is a helper function to parse the args slice into a map of flags.
// This is useful for testing without having to set up the full pflag parsing.
func parseArgs(args []string) (flowJSON string, out string, err error) {
	flags := pflag.NewFlagSet("convert", pflag.ContinueOnError)
	flowJSONPtr := flags.String("flow-json", "", "Path to the input DaVinci flow JSON file (required)")
	outPtr := flags.String("out", "", "Path to the output HCL file (optional, defaults to stdout)")

	if err := flags.Parse(args); err != nil {
		return "", "", err
	}

	if *flowJSONPtr == "" {
		return "", "", fmt.Errorf("--flow-json flag is required")
	}

	return *flowJSONPtr, *outPtr, nil
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
