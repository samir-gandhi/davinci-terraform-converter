// Copyright © 2025 Ping Identity Corporation

// Package cmd provides the command implementation for the DaVinci flow to HCL converter.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/exporter"
	"github.com/spf13/pflag"
)

// Command metadata for the DaVinci convert command
var (
	// Example provides usage examples for the command
	Example = `  # Convert a DaVinci flow to HCL and output to stdout
  pingcli davinci convert --flow-json ./my-flow.json

  # Convert a DaVinci flow to HCL and save to a file
  pingcli davinci convert --flow-json ./my-flow.json --out ./output.tf

  # Convert without generating Terraform references (use hardcoded IDs)
  pingcli davinci convert --flow-json ./my-flow.json --skip-dependencies

  # Export all DaVinci resources from an environment via API
  pingcli davinci convert --export \
    --environment-id <uuid> \
    --region NA \
    --client-id <id> \
    --client-secret <secret> \
    --out ./environment.tf

  # Export with skip-dependencies (use raw UUIDs instead of var.environment_id)
  pingcli davinci convert --export \
    --environment-id <uuid> \
    --skip-dependencies`

	// Long provides a detailed description of the command
	Long = `Convert a PingOne DaVinci flow from its native JSON format to HCL (HashiCorp Configuration Language).

The resulting HCL will be compatible with the DaVinci resources in the PingOne Terraform Provider.
Environment-specific values (such as connection IDs, variables, and subflows) will be converted 
to placeholder references that can be replaced with Terraform resource references.

Export Mode:
  When using --export, the command will authenticate to PingOne and export all DaVinci resources
  from the specified environment. This includes:
  - Variables
  - Connector Instances
  - Flows
  - Applications
  - Flow Policies

  Resources are exported in dependency order and include Terraform provider configuration.

  Authentication can be provided via flags or environment variables:
  - PINGONE_CLIENT_ID
  - PINGONE_CLIENT_SECRET
  - PINGONE_ENVIRONMENT_ID (or --environment-id flag)
  - PINGONE_REGION (or --region flag)`

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

	// Define file-based conversion flags
	flowJSON := flags.String("flow-json", "", "Path to the input DaVinci flow JSON file")
	out := flags.String("out", "", "Path to the output HCL file (optional, defaults to stdout)")
	skipDependencies := flags.Bool("skip-dependencies", false, "Skip generating Terraform references and use hardcoded IDs instead")

	// Define API export flags
	export := flags.Bool("export", false, "Enable API export mode to export all resources from an environment")
	environmentID := flags.String("environment-id", "", "PingOne environment ID for API export (or use PINGONE_ENVIRONMENT_ID env var)")
	region := flags.String("region", "", "PingOne region: NA, EU, AP, or CA (or use PINGONE_REGION env var)")
	clientID := flags.String("client-id", "", "OAuth client ID for authentication (or use PINGONE_CLIENT_ID env var)")
	clientSecret := flags.String("client-secret", "", "OAuth client secret for authentication (or use PINGONE_CLIENT_SECRET env var)")

	// Parse the provided arguments
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Determine which mode to use
	if *export {
		return c.runExportMode(logger, *environmentID, *region, *clientID, *clientSecret, *out, *skipDependencies)
	} else {
		return c.runFileMode(logger, *flowJSON, *out, *skipDependencies)
	}
}

// runFileMode handles conversion from a local JSON file
func (c *DaVinciConvertCommand) runFileMode(logger grpc.Logger, flowJSON, out string, skipDeps bool) error {
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

// runExportMode handles API export of all resources from an environment
func (c *DaVinciConvertCommand) runExportMode(logger grpc.Logger, environmentID, region, clientID, clientSecret, out string, skipDeps bool) error {
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
		return fmt.Errorf("environment ID is required: use --environment-id flag or PINGONE_ENVIRONMENT_ID env var")
	}
	if clientID == "" {
		return fmt.Errorf("client ID is required: use --client-id flag or PINGONE_CLIENT_ID env var")
	}
	if clientSecret == "" {
		return fmt.Errorf("client secret is required: use --client-secret flag or PINGONE_CLIENT_SECRET env var")
	}

	// Default region to NA if not specified
	if region == "" {
		region = "NA"
	}

	// Log export start
	if err := logger.Message(fmt.Sprintf("Exporting DaVinci environment: %s (Region: %s)", environmentID, region), nil); err != nil {
		return err
	}

	// Create API client
	ctx := context.Background()
	client, err := api.NewClientSingleEnvironment(ctx, environmentID, region, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Export all resources
	hcl, err := exporter.ExportEnvironment(ctx, client, skipDeps, logger)
	if err != nil {
		return fmt.Errorf("failed to export environment: %w", err)
	}

	// Write output
	if out != "" {
		if err := os.WriteFile(out, []byte(hcl), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if err := logger.Message(fmt.Sprintf("Successfully exported environment to: %s (%d bytes)", out, len(hcl)), nil); err != nil {
			return err
		}
	} else {
		// Write to stdout
		fmt.Print(hcl)
	}

	return nil
}

// parseArgs is a helper function to parse the args slice into a map of flags.
// This is useful for testing without having to set up the full pflag parsing.
func parseArgs(args []string) (flowJSON string, out string, skipDeps bool, err error) {
	flags := pflag.NewFlagSet("convert", pflag.ContinueOnError)
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
