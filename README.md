# DaVinci Terraform Converter

A CLI tool for converting PingOne DaVinci flows from their native JSON format to HCL (HashiCorp Configuration Language) compatible with the PingOne Terraform Provider.

**Dual Mode Operation:**
- **Standalone CLI**: Run directly from command line
- **PingCLI Plugin**: Integrate with `pingcli` tool

## Overview

This tool ingests a DaVinci flow JSON file (single flow or multi-flow export) and generates HCL code for the `pingone_davinci_flow` resource. It preserves all flow structure including nodes, edges, settings, and variables.

**Features:**
- ✅ Single flow conversion
- ✅ Multi-flow export support (parent flow + subflows)
- ✅ Complete attribute preservation (nodes, edges, settings, variables)
- ✅ Automatic resource name sanitization
- ✅ Multiple output modes (stdout, file, directory)
- ✅ **API Export Mode**: Export flows and connector instances directly from PingOne DaVinci
- ✅ OAuth2 authentication with PingOne SDK
- ✅ Skip-dependencies flag for standalone resource testing
- ✅ Secret masking in connector properties

## Installation

### Prerequisites

- Go 1.25.1 or higher
- (Optional) `pingcli` for plugin mode

### Building from Source

```bash
# Build only
make build

# Build and install to GOBIN (typically ~/go/bin)
make install

# Run all checks and build
make all
```

After installation, the binary will be available as `davinci-terraform-converter` in your `$GOBIN` or `$GOPATH/bin` directory.

## Usage

### Version Information

Check the installed version and git commit:

```bash
davinci-terraform-converter --version
# Output: davinci-convert version dev (commit: 4fb29ab...)
```

### Standalone CLI Mode

The binary can be run directly without `pingcli`:

```bash
# Convert single flow to stdout
davinci-terraform-converter --flow-json flow.json

# Convert single flow to file
davinci-terraform-converter --flow-json flow.json --out output.tf

# Convert multi-flow export to separate files in directory
davinci-terraform-converter --flow-json multiflow-export.json --out-dir ./flows

# Convert multi-flow export to single combined file
davinci-terraform-converter --flow-json multiflow-export.json --out combined.tf
```

**Note:** If you built with `make build` only (not installed), use `./davinci-convert` instead.

### PingCLI Plugin Mode

When launched by `pingcli`, it operates as a plugin:

```bash
# Convert a DaVinci flow to HCL and output to stdout
pingcli davinci convert --flow-json ./my-flow.json

# Convert a DaVinci flow to HCL and save to a file
pingcli davinci convert --flow-json ./my-flow.json --out ./output.tf
```

### Flags

- `-f, --flow-json` (required): Path to input DaVinci flow JSON file
- `-o, --out` (optional): Path to output HCL file (defaults to stdout)
- `-d, --out-dir` (optional): Directory for multi-flow output (creates separate .tf files)
- `--skip-dependencies` (optional): Export with hardcoded environment IDs instead of `var.environment_id` references. Useful for testing individual resources before full environment export is available.
- `-h, --help`: Show help message
- `-v, --version`: Show version information

### Skip Dependencies Flag

The `--skip-dependencies` flag is useful when exporting individual resources (flows, connector instances, etc.) without a complete environment configuration:

```bash
# Export flow with hardcoded environment ID
davinci-terraform-converter --flow-json flow.json --skip-dependencies

# This generates:
# resource "pingone_davinci_flow" "my_flow" {
#   environment_id = "62f10a04-6c54-40c2-a97d-80a98522ff9a"  # Actual ID
#   name = "My Flow"
# }

# Without --skip-dependencies (default):
# resource "pingone_davinci_flow" "my_flow" {
#   environment_id = var.environment_id  # Variable reference
#   name = "My Flow"
# }
```

**When to use:**
- Testing individual resource imports
- Exporting resources to different environments
- Before full environment export capability is available

**Note:** When full environment export is supported, `var.environment_id` (default behavior) will be preferred for portability.

## API Export Mode (In Development)

The tool now supports direct export from PingOne DaVinci environments via API:

### Prerequisites

Set up PingOne OAuth2 credentials as environment variables:

```bash
export PINGCLI_PINGONE_WORKER_CLIENT_ID="your-client-id"
export PINGCLI_PINGONE_WORKER_CLIENT_SECRET="your-client-secret"
export PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID="auth-environment-id"
export PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID="target-environment-id"
export PINGONE_REGION="NA"  # or EU, AP, CA
```

### Current API Export Capabilities

**Flows:**
- Export all flows from a DaVinci environment
- Generates 442KB HCL from 8 flows (real environment test)
- Includes flow graph data, nodes, edges, and settings

**Connector Instances:**
- Export all connector instances from a DaVinci environment
- Generates 4.6KB HCL from 20 instances (real environment test)
- Includes connector properties with secret masking

**Coming Soon:**
- Variables export
- Applications export
- Flow policies export
- Complete environment export

### Running Acceptance Tests

Acceptance tests validate against real PingOne API:

```bash
# Run acceptance tests (requires environment variables)
go test -tags=acceptance ./tests/acceptance -v

# Skip acceptance tests (runs unit tests only)
go test ./...
```

## Input Formats

### Single Flow Export

```json
```

## Input Formats

### Single Flow Export

```json
{
  "name": "My Flow",
  "flowId": "abc123",
  "graphData": { ... },
  "settings": { ... }
}
```

### Multi-Flow Export (Parent + Subflows)

```json
{
  "flows": [
    {
      "name": "Parent Flow",
      "flowId": "parent-id",
      "graphData": { ... }
    },
    {
      "name": "Subflow",
      "flowId": "subflow-id",
      "parentFlowId": "parent-id",
      "graphData": { ... }
    }
  ],
  "companyId": "...",
  "customerId": "..."
}
```

## Output Example

**Input:** Simple flow with nodes and settings

**Output:**
```hcl
resource "pingone_davinci_flow" "simple_demo_flow" {
  environment_id = var.environment_id

  name        = "Simple Demo Flow"
  description = "A simple flow for demonstration"

  graph_data {
    elements {
      nodes = [
        {
          "data": {
            "capabilityName": "customHtmlMessage",
            "connectionId": "conn-abc-123",
            "connectorId": "httpConnector",
            "id": "httpNode",
            "nodeType": "CONNECTION",
            "properties": {
              "message": {
                "value": "Welcome to the flow!"
              }
            }
          }
        }
      ]

      edges = [
        {
          "data": {
            "id": "edge1",
            "source": "httpNode",
            "target": "evalNode"
          }
        }
      ]
    }
  }

  settings {
    {
      "csp": "default-src 'self';",
      "logLevel": 4
    }
  }
}
```

## Development

### Project Structure

```
.
├── cmd/                    # Command implementation (plugin interface)
│   ├── convert.go         # Command implementation
│   └── convert_test.go    # Command tests
├── internal/              # Internal packages
│   └── converter/         # Core conversion logic
│       ├── converter.go   # Main conversion functions
│       ├── converter_test.go  # Unit tests (26 tests)
│       └── real_file_test.go  # Integration test
├── .github/
│   └── prompts/           # Example flow files
├── main.go                # Dual-mode entry point (plugin + CLI)
├── go.mod                 # Go module dependencies
├── Makefile               # Build automation
└── README.md              # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run with coverage report
make test-coverage

# Run specific test
go test ./internal/converter/... -v -run TestMultiFlowExport

# Format code
make fmt

# Run linters
make lint
```

### Available Make Targets

```bash
make help  # Display all available targets
```

- `build` - Build the plugin binary
- `install` - Build and install to GOBIN
- `test` - Run all tests
- `test-verbose` - Run tests with verbose output
- `test-coverage` - Generate coverage report
- `clean` - Clean build artifacts
- `fmt` - Format Go code
- `vet` - Run go vet
- `lint` - Run all linting tools
- `deps` - Download and tidy dependencies
- `all` - Run all checks and build

### Test Coverage

Current test coverage as of Phase 3.2b:

**Unit Tests:**
- Total: 70 tests passing across all packages
- `internal/api`: 19 tests (client, flows, connector instances)
- `internal/converter`: 35 tests (flows, connectors, applications, variables)
- `internal/exporter`: 6 tests (flow export, connector export)
- `internal/utils`: 9 tests (resource name sanitization)
- `cmd`: 1 test (root command)

**Acceptance Tests:**
- Total: 20 tests with real API calls
- Flow API: 6 tests (list, get, error handling, multi-retrieval)
- Connector Instance API: 4 tests (list, get, error handling, multi-retrieval)
- Flow Export: 5 tests (export, skip-deps, HCL structure, comparison, JSON format)
- Connector Instance Export: 5 tests (export, skip-deps, HCL structure, comparison, properties)

**Coverage by Module:**
```bash
$ go test ./internal/converter/... -cover
coverage: 91.3% of statements
```

**Real Environment Testing:**
- 8 flows exported (442KB HCL)
- 20 connector instances exported (4.6KB HCL)
- Authentication with PingOne OAuth2
- Dual-environment support (auth + target)

**Test Approach:**
- Unit tests for logic validation
- Acceptance tests for API integration (requires credentials)
- Skip acceptance tests when `PINGCLI_PINGONE_*` environment variables not set

### Building

```bash
# Build binary
make build

# Clean, test, and build
make all

# Generate coverage report
make test-coverage
```

## Architecture

This tool follows TDD (Test-Driven Development) and is built in phases:

1. **Project Scaffolding** (✅ Complete): Basic command structure and flag parsing
2. **Core Conversion Logic** (✅ Complete): Convert flows to HCL with full attribute support
   - Single flow conversion
   - Multi-flow export support
   - GraphData (nodes + edges)
   - Settings blocks
   - Variables documentation
   - Connector instances
   - Applications
3. **API Export Integration** (⏳ In Progress - Phase 6/15 Complete):
   - ✅ OAuth2 authentication with PingOne SDK
   - ✅ Dual-environment support (auth + target)
   - ✅ Flow listing and retrieval from API
   - ✅ Flow export to HCL (442KB from 8 flows)
   - ✅ Connector instance listing and retrieval
   - ✅ Connector instance export to HCL (4.6KB from 20 instances)
   - ⏳ Variables API and export (Phase 3.3)
   - ⏳ Applications API and export (Phase 3.4)
   - ⏳ Flow policies API and export (Phase 3.5)
   - ⏳ Complete environment export (Phase 3.6)
   - ⏳ CLI integration (Phase 3.7)
4. **Error Handling & Polish** (⏳ Future): Enhanced validation and user experience

**Current Capabilities:**
- Export flows from PingOne DaVinci environment via API
- Export connector instances from PingOne DaVinci environment via API
- Convert to Terraform HCL with `--skip-dependencies` support
- 70 unit tests + 20 acceptance tests with real API integration

## Examples

Example flow files are provided in `.github/prompts/`:

- `simple-demo-flow.json` - Simple single flow
- `PingOne_Sign On with Sessions_multiflow.json` - Real production multi-flow export

## References

- [PingCLI](https://github.com/pingidentity/pingcli)
- [PingOne Go SDK](https://github.com/pingidentity/pingone-go-client)
- [PingOne Terraform Provider](https://github.com/pingidentity/terraform-provider-pingone)
- [DaVinci Documentation](https://docs.pingidentity.com/davinci/)

## License

Copyright © 2025 Ping Identity Corporation

See LICENSE file for details.
