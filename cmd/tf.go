package cmd

import (
	"fmt"

	"github.com/pingidentity/pingcli/shared/grpc"
)

var (
	// Parent command metadata
	TfExample = `  # Convert a DaVinci flow JSON file to Terraform HCL
  pingcli tf davinci-to-hcl --flow-json ./my-flow.json --out ./output.tf

  # Export PingOne DaVinci resources to Terraform HCL
  pingcli tf export --services pingone-davinci --environment-id <uuid> --out ./environment.tf

  # Get help for subcommands
  pingcli tf davinci-to-hcl --help
  pingcli tf export --help`

	TfLong = `Terraform utilities for Ping Identity resources.

Provides tools to convert configurations and export resources to Terraform HCL format 
compatible with the PingOne and PingFederate Terraform Providers.

Available subcommands:
  davinci-to-hcl - Convert a DaVinci flow JSON file to HCL
  export         - Export Ping Identity resources from live environments to HCL

Supported services for export:
  pingone-davinci - PingOne DaVinci flows, variables, connections, apps, policies`

	TfShort = "Terraform utilities for Ping Identity"

	TfUse = "tf [subcommand]"
)

// TfCommand is the parent command that routes to subcommands
type TfCommand struct{}

// Ensure TfCommand implements grpc.PingCliCommand
var _ grpc.PingCliCommand = (*TfCommand)(nil)

// Configuration returns the parent command metadata
func (c *TfCommand) Configuration() (*grpc.PingCliCommandConfiguration, error) {
	return &grpc.PingCliCommandConfiguration{
		Use:     TfUse,
		Short:   TfShort,
		Long:    TfLong,
		Example: TfExample,
	}, nil
}

// Run routes to the appropriate subcommand
func (c *TfCommand) Run(args []string, logger grpc.Logger) error {
	// Check if subcommand provided
	if len(args) == 0 {
		return fmt.Errorf("subcommand required. Use 'pingcli tf --help' for usage")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "davinci-to-hcl":
		cmd := &DaVinciToHclCommand{}
		return cmd.Run(subArgs, logger)

	case "export":
		cmd := &ExportCommand{}
		return cmd.Run(subArgs, logger)

	case "--help", "-h", "help":
		// Show help text
		config, _ := c.Configuration()
		helpText := fmt.Sprintf(`%s

Usage:
  %s

%s

Examples:
%s

Use "pingcli tf [subcommand] --help" for more information about a subcommand.`,
			config.Short, config.Use, config.Long, config.Example)
		return logger.Message(helpText, nil)

	default:
		return fmt.Errorf("unknown subcommand: %s\nUse 'pingcli tf --help' for usage", subcommand)
	}
}
