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
- `-h, --help`: Show help message
- `-v, --version`: Show version information

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

```bash
$ go test ./internal/converter/... -cover
coverage: 91.3% of statements

$ go test ./internal/converter/... -v 2>&1 | grep -c "^=== RUN"
26
```

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
   - 25 comprehensive tests
3. **Environment-Specific Dependencies** (⏳ Next): Handle connection IDs, variables, and subflows
4. **Integration and Error Handling** (⏳ Future): Enhanced file I/O and validation

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
