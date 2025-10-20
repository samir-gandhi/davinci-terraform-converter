# Plugin Restructuring Plan: Multi-Command Support

**Date**: October 20, 2025  
**Goal**: Restructure as a proper Ping CLI plugin with parent command `tf` and subcommands `convert` and `export`

## Current Problem

The current implementation tries to be a single command (`davinci convert`) with dual modes (file vs export), which doesn't fit well with Ping CLI's plugin architecture. Also the current implementation is narrowly focused on davinci. DaVinci will be the focus initially, but this plugin should leave room for growth into all supported ping identity terraform resources - this means all resources in pingone and pingfederate and room for growth in the future. Keep this room for growth in mind while building new structures. 

**Current Structure**:
```
pingcli davinci convert --flow-json file.json        # File mode
pingcli davinci convert --export --environment-id ... # Export mode
```

**Issues**:
- Conflicting flags between modes
- Confusing user experience (which flags work with which mode?)
- Doesn't follow Ping CLI plugin patterns
- Single command trying to do two different things

## Proposed Solution

Create a parent command `tf` with two subcommands following Ping CLI patterns:

```
pingcli tf davinci-to-hcl --flow-json file.json       # Convert mode
pingcli tf export --environment-id <id> ...           # Export mode
```

This follows the standard Ping CLI pattern where plugins register parent commands with subcommands.

## Architecture Overview

### Plugin Structure

```
davinci-terraform-converter/
├── main.go                    # Plugin entry point (serves parent command)
├── cmd/
│   ├── tf.go                 # NEW: Parent "tf" command (dispatcher)
│   ├── convert.go            # Convert subcommand (file mode)
│   └── export.go             # NEW: Export subcommand (environment mode)
├── internal/
│   ├── api/                  # Existing API client
│   ├── converter/            # Existing conversion logic
│   ├── exporter/             # Existing export logic
│   └── resolver/             # Existing dependency resolution
```

### Command Hierarchy

```
tf (parent)
├── convert (subcommand)
└── export (subcommand)
```

**Parent Command** (`tf`):
- Registered with Ping CLI as `tf`
- No direct execution logic
- Routes to subcommands based on args[0]
- Provides help text for the command group

**Subcommand: convert**:
- Converts DaVinci flow JSON to Terraform HCL
- File-based operation
- Flags: `--flow-json`, `--out`, `--skip-dependencies`

**Subcommand: export**:
- Exports entire environment from API
- API-based operation
- Flags: `--environment-id`, `--region`, `--client-id`, `--client-secret`, `--out`, `--skip-dependencies`

## Implementation Plan

### Phase 1: Create Parent Command Structure

**File**: `cmd/tf.go` (NEW)

```go
package cmd

import (
    "fmt"
    "github.com/pingidentity/pingcli/shared/grpc"
)

var (
    // Parent command metadata
    TfExample = `  # Convert a DaVinci flow JSON file to Terraform HCL
  pingcli tf convert --flow-json ./my-flow.json --out ./output.tf

  # Export an entire DaVinci environment to Terraform HCL
  pingcli tf export --environment-id <uuid> --out ./environment.tf

  # Get help for subcommands
  pingcli tf convert --help
  pingcli tf export --help`

    TfLong = `Terraform utilities for PingOne DaVinci resources.

Provides tools to convert DaVinci flows and export entire environments 
to Terraform HCL format compatible with the PingOne Terraform Provider.

Available subcommands:
  convert  - Convert a single DaVinci flow JSON file to HCL
  export   - Export all DaVinci resources from an environment to HCL`

    TfShort = "Terraform utilities for DaVinci"

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
    case "convert":
        cmd := &ConvertCommand{}
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
```

### Phase 2: Update Convert Command

**File**: `cmd/convert.go` (UPDATE)

Keep existing logic but update metadata:

```go
package cmd

var (
    ConvertExample = `  # Convert a DaVinci flow to HCL and output to stdout
  pingcli tf convert --flow-json ./my-flow.json

  # Convert and save to a file
  pingcli tf convert --flow-json ./my-flow.json --out ./output.tf

  # Convert without generating Terraform references (use hardcoded IDs)
  pingcli tf convert --flow-json ./my-flow.json --skip-dependencies`

    ConvertLong = `Convert a single PingOne DaVinci flow from JSON to Terraform HCL.

Reads a DaVinci flow export (JSON format) and converts it to HCL syntax
compatible with the pingone_davinci_flow resource in the PingOne Terraform Provider.

By default, generates Terraform references for dependencies. Use --skip-dependencies
to preserve raw UUIDs instead.`

    ConvertShort = "Convert a DaVinci flow JSON file to Terraform HCL"

    ConvertUse = "convert --flow-json <file> [flags]"
)

type ConvertCommand struct{}

// Remove --export flag handling
// Keep only: --flow-json, --out, --skip-dependencies
func (c *ConvertCommand) Run(args []string, logger grpc.Logger) error {
    // Parse flags
    flags := pflag.NewFlagSet("convert", pflag.ContinueOnError)
    
    var flowJSON string
    var out string
    var skipDeps bool
    
    flags.StringVar(&flowJSON, "flow-json", "", "Path to DaVinci flow JSON file (required)")
    flags.StringVarP(&out, "out", "o", "", "Output file path (default: stdout)")
    flags.BoolVar(&skipDeps, "skip-dependencies", false, "Skip dependency resolution")
    
    if err := flags.Parse(args); err != nil {
        return err
    }
    
    // Validate required flags
    if flowJSON == "" {
        return fmt.Errorf("--flow-json is required")
    }
    
    // Execute conversion (existing logic)
    return c.runConvert(logger, flowJSON, out, skipDeps)
}

// Rename runFileMode to runConvert for clarity
func (c *ConvertCommand) runConvert(logger grpc.Logger, flowJSON, out string, skipDeps bool) error {
    // Existing conversion logic
    // ...
}
```

### Phase 3: Create Export Command

**File**: `cmd/export.go` (NEW)

```go
package cmd

import (
    "context"
    "fmt"
    "os"
    
    "github.com/pingidentity/pingcli/shared/grpc"
    "github.com/samir-gandhi/davinci-terraform-converter/internal/api"
    "github.com/samir-gandhi/davinci-terraform-converter/internal/exporter"
    "github.com/spf13/pflag"
)

var (
    ExportExample = `  # Export entire DaVinci environment
  pingcli tf export --environment-id <uuid> --out ./environment.tf

  # Export with explicit credentials
  pingcli tf export \
    --environment-id <uuid> \
    --region NA \
    --client-id <id> \
    --client-secret <secret> \
    --out ./davinci.tf

  # Export without Terraform dependencies (raw UUIDs)
  pingcli tf export --environment-id <uuid> --skip-dependencies

  # Use environment variables for credentials
  export PINGONE_ENVIRONMENT_ID="..."
  export PINGONE_CLIENT_ID="..."
  export PINGONE_CLIENT_SECRET="..."
  pingcli tf export --out ./environment.tf`

    ExportLong = `Export all DaVinci resources from a PingOne environment to Terraform HCL.

Connects to the PingOne DaVinci API and exports:
  • Variables
  • Connector Instances
  • Flows
  • Applications
  • Flow Policies

The generated HCL includes proper Terraform resource references and dependency ordering.

Authentication can be provided via flags or environment variables:
  PINGONE_ENVIRONMENT_ID
  PINGONE_CLIENT_ID
  PINGONE_CLIENT_SECRET
  PINGONE_REGION (default: NA)`

    ExportShort = "Export DaVinci environment to Terraform HCL"

    ExportUse = "export --environment-id <uuid> [flags]"
)

type ExportCommand struct{}

var _ grpc.PingCliCommand = (*ExportCommand)(nil)

func (c *ExportCommand) Configuration() (*grpc.PingCliCommandConfiguration, error) {
    return &grpc.PingCliCommandConfiguration{
        Use:     ExportUse,
        Short:   ExportShort,
        Long:    ExportLong,
        Example: ExportExample,
    }, nil
}

func (c *ExportCommand) Run(args []string, logger grpc.Logger) error {
    // Parse flags
    flags := pflag.NewFlagSet("export", pflag.ContinueOnError)
    
    var environmentID string
    var region string
    var clientID string
    var clientSecret string
    var out string
    var skipDeps bool
    
    flags.StringVar(&environmentID, "environment-id", "", "PingOne environment ID")
    flags.StringVar(&region, "region", "", "PingOne region (NA, EU, AP, CA)")
    flags.StringVar(&clientID, "client-id", "", "OAuth client ID")
    flags.StringVar(&clientSecret, "client-secret", "", "OAuth client secret")
    flags.StringVarP(&out, "out", "o", "", "Output file path (default: stdout)")
    flags.BoolVar(&skipDeps, "skip-dependencies", false, "Skip dependency resolution")
    
    if err := flags.Parse(args); err != nil {
        return err
    }
    
    // Execute export
    return c.runExport(logger, environmentID, region, clientID, clientSecret, out, skipDeps)
}

func (c *ExportCommand) runExport(logger grpc.Logger, environmentID, region, clientID, clientSecret, out string, skipDeps bool) error {
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
        logger.PluginError("Failed to create API client", map[string]string{
            "environment_id": environmentID,
            "region": region,
            "error": err.Error(),
        })
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
            logger.PluginError("Failed to write output file", map[string]string{
                "file": out,
                "error": err.Error(),
            })
            return fmt.Errorf("failed to write output file: %w", err)
        }
        if err := logger.Message(fmt.Sprintf("✓ Successfully exported to: %s (%d bytes)", out, len(hcl)), nil); err != nil {
            return err
        }
    } else {
        // Write to stdout
        fmt.Println(hcl)
    }

    return nil
}
```

### Phase 4: Update Main Entry Point

**File**: `main.go` (UPDATE)

```go
// Change plugin registration to use TfCommand instead of DaVinciConvertCommand

func runAsPlugin() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: grpc.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            grpc.ENUM_PINGCLI_COMMAND_GRPC: &grpc.PingCliCommandGrpcPlugin{
                Impl: &cmd.TfCommand{},  // Changed from DaVinciConvertCommand
            },
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}

// Update standalone mode to parse subcommands
func runStandalone() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Error: subcommand required")
        fmt.Fprintln(os.Stderr, "Usage: davinci-convert [convert|export] [flags]")
        os.Exit(1)
    }

    subcommand := os.Args[1]
    args := os.Args[2:]
    logger := &simpleLogger{}

    tfCmd := &cmd.TfCommand{}
    
    // Pass subcommand as first arg
    if err := tfCmd.Run(append([]string{subcommand}, args...), logger); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Phase 5: Update Tests

**Files to update**:
- `cmd/convert_test.go` - Update to test ConvertCommand directly
- `cmd/export_test.go` - NEW: Create tests for ExportCommand
- `cmd/tf_test.go` - NEW: Test parent command routing

**Example test structure**:

```go
// cmd/tf_test.go
func TestTfCommand_Routing(t *testing.T) {
    tests := []struct{
        name string
        args []string
        expectError bool
    }{
        {"no args", []string{}, true},
        {"convert subcommand", []string{"convert", "--help"}, false},
        {"export subcommand", []string{"export", "--help"}, false},
        {"unknown subcommand", []string{"invalid"}, true},
    }
    // ...
}

// cmd/convert_test.go
func TestConvertCommand_Run(t *testing.T) {
    // Test convert command directly
}

// cmd/export_test.go
func TestExportCommand_Run(t *testing.T) {
    // Test export command directly
}
```

### Phase 6: Update Documentation

**Files to update**:
- `README.md` - Update all examples to use `pingcli tf convert` and `pingcli tf export`
- `.github/prompts/07-part5-integration.md` - Update Phase 5 documentation
- All example scripts in `examples/` directory

**Before**:
```bash
pingcli davinci convert --flow-json flow.json
pingcli davinci convert --export --environment-id <uuid>
```

**After**:
```bash
pingcli tf convert --flow-json flow.json
pingcli tf export --environment-id <uuid>
```

## Benefits of New Structure

### 1. **Clear Separation of Concerns**
- Convert command: file operations only
- Export command: API operations only
- No confusing flag combinations

### 2. **Better User Experience**
```bash
# Clear, intuitive commands
pingcli tf convert --flow-json file.json
pingcli tf export --environment-id <uuid>

# Subcommand-specific help
pingcli tf convert --help
pingcli tf export --help
```

### 3. **Follows Ping CLI Patterns**
Other Ping CLI plugins use parent/subcommand structure:
```bash
pingcli platform export ...
pingcli platform import ...
```

### 4. **Easier to Extend**
Future subcommands can be added easily:
```bash
pingcli tf validate     # Validate generated HCL
pingcli tf plan         # Terraform plan wrapper
pingcli tf apply        # Terraform apply wrapper
```

### 5. **Plugin Discovery**
Ping CLI will show:
```bash
$ pingcli --help
Available commands:
  ...
  tf          Terraform utilities for DaVinci
  ...

$ pingcli tf --help
Terraform utilities for PingOne DaVinci resources.

Available subcommands:
  convert  - Convert a single DaVinci flow JSON file to HCL
  export   - Export all DaVinci resources from an environment to HCL
```

## Migration Path

### Step 1: Create New Structure (No Breaking Changes)
- Add `cmd/tf.go`
- Add `cmd/export.go`
- Update `cmd/convert.go` to be subcommand-aware
- Update `main.go` to use TfCommand

### Step 2: Update Tests
- Create unit tests for all commands
- Update integration tests
- Verify both plugin and standalone modes work

### Step 3: Update Documentation
- Update README with new command structure
- Update all examples
- Add migration guide for existing users

### Step 4: Deprecation (Optional)
If you want to maintain backward compatibility temporarily:
- Keep old `davinci convert` command as alias
- Show deprecation warning
- Direct users to new commands

### Step 5: Release
- Tag as v2.0.0 (breaking change)
- Update Ping CLI plugin registry
- Announce new command structure

## File Checklist

### New Files
- [ ] `cmd/tf.go` - Parent command
- [ ] `cmd/export.go` - Export subcommand
- [ ] `cmd/tf_test.go` - Parent command tests
- [ ] `cmd/export_test.go` - Export subcommand tests

### Modified Files
- [ ] `cmd/convert.go` - Update to be subcommand
- [ ] `main.go` - Use TfCommand, update standalone mode
- [ ] `README.md` - Update all examples
- [ ] `.github/prompts/07-part5-integration.md` - Update documentation
- [ ] All example scripts in `examples/`

### Test Files
- [ ] `cmd/convert_test.go` - Update for subcommand
- [ ] Integration tests - Update command invocations

## Timeline Estimate

- **Phase 1-2**: Create parent + update convert - 2 hours
- **Phase 3**: Create export command - 1 hour
- **Phase 4**: Update main.go - 30 minutes
- **Phase 5**: Update all tests - 2 hours
- **Phase 6**: Documentation updates - 1 hour

**Total**: ~6-7 hours

## Success Criteria

- [ ] `pingcli tf convert --flow-json file.json` works
- [ ] `pingcli tf export --environment-id <uuid>` works
- [ ] `pingcli tf --help` shows subcommands
- [ ] `pingcli tf convert --help` shows convert-specific help
- [ ] `pingcli tf export --help` shows export-specific help
- [ ] All tests pass
- [ ] Plugin works in Ping CLI
- [ ] Standalone mode works
- [ ] Documentation complete and accurate
- [ ] No breaking changes to internal APIs

## Next Steps

1. **Review this plan** - Confirm approach is correct
2. **Create feature branch** - `feature/multi-command-restructure`
3. **Implement Phase 1** - Create parent command
4. **Implement Phases 2-3** - Update convert, create export
5. **Implement Phase 4** - Update main.go
6. **Implement Phase 5** - Update all tests
7. **Implement Phase 6** - Update documentation
8. **Test thoroughly** - Both plugin and standalone modes
9. **Release** - Tag new version

Ready to proceed with implementation?
