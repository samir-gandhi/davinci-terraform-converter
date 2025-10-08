# DaVinci Terraform Converter

A CLI plugin for `pingcli` that converts PingOne DaVinci flows from their native JSON format to HCL (HashiCorp Configuration Language) compatible with the PingOne Terraform Provider.

## Overview

This tool ingests a DaVinci flow JSON file and generates HCL code for the `pingone_davinci_flow` resource. It intelligently handles environment-specific values (like connection IDs, variables, and subflows) by converting them into placeholder references that can be replaced with Terraform resource references.

## Installation

### Prerequisites

- Go 1.25.1 or higher
- `pingcli` (https://github.com/pingidentity/pingcli)

### Building from Source

```bash
go build -o davinci-convert .
```

## Usage

The plugin is designed to be used as a `pingcli` plugin command:

```bash
# Convert a DaVinci flow to HCL and output to stdout
pingcli davinci convert --flow-json ./my-flow.json

# Convert a DaVinci flow to HCL and save to a file
pingcli davinci convert --flow-json ./my-flow.json --out ./output.tf
```

### Flags

- `--flow-json` (required): Path to the input DaVinci flow JSON file
- `--out` (optional): Path to the output HCL file. If not provided, output goes to stdout

## Development

### Project Structure

```
.
├── cmd/                    # Command implementation
│   ├── convert.go         # Main command logic
│   └── convert_test.go    # Command tests
├── internal/              # Internal packages
│   └── converter/         # Core conversion logic (to be implemented)
├── main.go                # Plugin entry point
├── go.mod                 # Go module dependencies
└── README.md              # This file
```

### Running Tests

```bash
go test ./...
```

### Running with Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture

This plugin follows the TDD (Test-Driven Development) approach and is built in phases:

1. **Project Scaffolding** (✅ Complete): Basic command structure and flag parsing
2. **Core Conversion Logic**: Convert simple flows to HCL
3. **Environment-Specific Dependencies**: Handle connection IDs, variables, and subflows
4. **Integration and Error Handling**: Robust file I/O and error handling

## References

- [PingCLI](https://github.com/pingidentity/pingcli)
- [PingOne Go SDK](https://github.com/pingidentity/pingone-go-client)
- [PingOne Terraform Provider](https://github.com/pingidentity/terraform-provider-pingone)

## License

Copyright © 2025 Ping Identity Corporation

See LICENSE file for details.
