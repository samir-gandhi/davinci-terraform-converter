# DaVinci Terraform Converter

A CLI tool for converting PingOne DaVinci resources to HCL (HashiCorp Configuration Language) compatible with the PingOne Terraform Provider.

**Dual Mode Operation:**
- **Standalone CLI**: Run directly from command line as `davinci-convert`
- **PingCLI Plugin**: Integrate with `pingcli tf` commands

## Overview

This tool provides two primary workflows:

1. **File Conversion** (`davinci-to-hcl`): Convert DaVinci flow JSON files to HCL
2. **API Export** (`export`): Export complete DaVinci environments from PingOne API to HCL

**Features:**

- ✅ Complete environment export from PingOne DaVinci API
- ✅ Export flows, connector instances, variables, applications, and flow policies
- ✅ Automatic dependency resolution with Terraform references
- ✅ **Terraform import blocks** for automatic state import (Terraform 1.5+)
- ✅ Two-environment authentication model (worker app + target environment)
- ✅ Service-based architecture (extensible to future Ping products)
- ✅ Skip-dependencies mode for standalone resource testing
- ✅ OAuth2 authentication with PingOne SDK
- ✅ Resource name sanitization
- ✅ Secret masking in connector properties
- ✅ Comprehensive validation and error reporting

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

### Command Structure

The tool provides two primary commands:

**1. `davinci-to-hcl` - Convert DaVinci flow JSON files to HCL**

Use this for file-based conversion of exported flow JSON.

**2. `export` - Export resources from PingOne DaVinci API to HCL**

Use this for complete environment exports with automatic dependency resolution.

### Standalone CLI Mode

#### File Conversion (davinci-to-hcl)

```bash
# Convert flow JSON to HCL (stdout)
davinci-convert davinci-to-hcl --flow-json flow.json

# Convert and save to file
davinci-convert davinci-to-hcl --flow-json flow.json --out output.tf

# Skip dependencies (use hardcoded IDs)
davinci-convert davinci-to-hcl --flow-json flow.json --skip-dependencies
```

#### API Export (export)

```bash
# Export all services (defaults to all available: pingone-davinci)
davinci-convert export \
  --pingone-worker-environment-id "abc123..." \
  --pingone-export-environment-id "def456..." \
  --pingone-worker-client-id "client-id" \
  --pingone-worker-client-secret "client-secret" \
  --pingone-region-code "NA" \
  --out environment.tf

# Export without import blocks (for Terraform < 1.5)
davinci-convert export \
  --pingone-worker-environment-id "abc123..." \
  --pingone-export-environment-id "def456..." \
  --pingone-worker-client-id "client-id" \
  --pingone-worker-client-secret "client-secret" \
  --out environment.tf \
  --skip-imports

# Export with skip-dependencies (hardcoded UUIDs)
davinci-convert export \
  --pingone-worker-environment-id "abc123..." \
  --pingone-export-environment-id "def456..." \
  --pingone-worker-client-id "client-id" \
  --pingone-worker-client-secret "client-secret" \
  --skip-dependencies

# Use environment variables (see Configuration section)
davinci-convert export --out environment.tf

# Skip import blocks if not using Terraform 1.5+
davinci-convert export --out environment.tf --skip-imports
```

### Terraform Import Blocks

By default, all exports generate Terraform import blocks alongside resource definitions, enabling automatic import of existing DaVinci resources into Terraform state. Requires Terraform 1.5+.

**Default Workflow:**

```bash
# Export (import blocks generated automatically)
davinci-convert export --out davinci.tf

# Import all resources automatically
terraform init
terraform apply  # Imports all resources in single operation
```

**Without import blocks** (use `--skip-imports`):

```bash
# Export without import blocks (Terraform < 1.5 or manual import preferred)
davinci-convert export --out davinci.tf --skip-imports

# Manual import (required for each resource)
terraform import pingone_davinci_variable.var1 "env-id/var-id"
terraform import pingone_davinci_flow.flow1 "env-id/flow-id"
# ... repeat 50+ times for complete environment
```

**Benefits:**

- Default behavior eliminates 50+ manual import commands
- No manual ID mapping or copy/paste errors
- Import blocks placed immediately before resource definitions
- Works with all 6 DaVinci resource types
- Supports both dependency and skip-dependencies modes
- Use `--skip-imports` for Terraform < 1.5 or manual import workflows

**Example output:**

```hcl
import {
  to = pingone_davinci_variable.companyName
  id = "abc123-def456.../var-id-789..."
}

resource "pingone_davinci_variable" "companyName" {
  environment_id = var.environment_id
  name          = "companyName"
  type          = "string"
  value         = "Acme Corp"
}

import {
  to = pingone_davinci_flow.registrationFlow
  id = "abc123-def456.../flow-id-xyz..."
}

resource "pingone_davinci_flow" "registrationFlow" {
  environment_id = var.environment_id
  name          = "registrationFlow"
  # ... flow configuration
}
```

See `examples/05-import-blocks-usage.sh` for complete demonstration.

### PingCLI Plugin Mode

When used with `pingcli`, commands are namespaced under `tf`:

```bash
# Convert DaVinci flow JSON to HCL
pingcli tf davinci-to-hcl --flow-json flow.json --out output.tf

# Export environment from API
pingcli tf export \
  --pingone-worker-environment-id "abc123..." \
  --pingone-export-environment-id "def456..." \
  --pingone-worker-client-id "client-id" \
  --pingone-worker-client-secret "client-secret" \
  --out environment.tf
```

### Configuration

#### Environment Variables

Configuration follows Ping CLI standards with `PINGCLI_` prefix:

```bash
# Worker environment (for authentication)
export PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID="abc123-def456-..."
export PINGCLI_PINGONE_WORKER_CLIENT_ID="your-client-id"
export PINGCLI_PINGONE_WORKER_CLIENT_SECRET="your-client-secret"

# Export environment (target resources)
export PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID="target-env-id"

# Region (NA, EU, AP, CA, AU)
export PINGCLI_PINGONE_REGION_CODE="NA"
```

**Priority:** Command-line flags > Environment variables > Defaults

#### Two-Environment Model

The export command uses a two-environment architecture:

- **Worker Environment**: Contains OAuth2 worker app for authentication
- **Export Environment**: Target environment containing resources to export

**Benefits:**
- Isolate credentials from exported resources
- Export from multiple environments with same worker app
- Support dev/staging/prod workflows
- Single-environment convenience (export environment defaults to worker environment)

**Example:**

```bash
# Two environments (recommended for production)
davinci-convert export \
  --pingone-worker-environment-id "auth-env-id" \
  --pingone-export-environment-id "prod-env-id" \
  ...

# Single environment (development convenience)
davinci-convert export \
  --pingone-worker-environment-id "dev-env-id" \
  ...
```

### Flags Reference

#### davinci-to-hcl Command

| Flag | Required | Description |
|------|----------|-------------|
| `--flow-json` | Yes | Path to input DaVinci flow JSON file |
| `--out` | No | Output file path (defaults to stdout) |
| `--skip-dependencies` | No | Use hardcoded IDs instead of variable references |

#### export Command

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--services` | No | `["pingone-davinci"]` | Services to export (currently only pingone-davinci) |
| `--pingone-worker-environment-id` | Yes* | - | Worker environment ID for authentication |
| `--pingone-export-environment-id` | No | Worker env | Target environment ID for resource export |
| `--pingone-worker-client-id` | Yes* | - | OAuth2 client ID |
| `--pingone-worker-client-secret` | Yes* | - | OAuth2 client secret |
| `--pingone-region-code` | No | `NA` | Region: NA, EU, AP, CA, AU |
| `--out` | No | stdout | Output file path |
| `--skip-dependencies` | No | false | Use hardcoded UUIDs instead of Terraform references |

\* Required unless set via environment variables

### Skip Dependencies Mode

The `--skip-dependencies` flag controls how resource dependencies are handled:

**Without `--skip-dependencies` (default):**
```hcl
resource "pingone_davinci_flow" "my_flow" {
  environment_id = var.environment_id
  connection_id  = pingone_davinci_connector_instance.httpconnector_abc123.id
}
```

**With `--skip-dependencies`:**
```hcl
resource "pingone_davinci_flow" "my_flow" {
  environment_id = "62f10a04-6c54-40c2-a97d-80a98522ff9a"
  connection_id  = "abc123-def456-ghi789-..."
}
```


**When to use:**

- Testing individual resource imports
- Standalone resource files without full environment
- Quick prototyping

**Note:** API exports include full environment context, so skip-dependencies produces fully standalone HCL. File conversions (`davinci-to-hcl`) of standalone JSON files may still require `var.environment_id` if the JSON lacks environment metadata.

## Output Examples

### Export Command Output

A typical full environment export produces comprehensive HCL:

```hcl
# Variables (standalone, no dependencies)
resource "pingone_davinci_variable" "companyname" {
  environment_id = var.environment_id
  name           = "companyName"
  description    = "Company branding variable"
  type           = "string"
  value          = "Ping Identity"
}

# Connector Instances (standalone, may have masked secrets)
resource "pingone_davinci_connector_instance" "httpconnector_abc123" {
  environment_id = var.environment_id
  connector_id   = "httpConnector"
  name           = "HTTP Connector"

  property {
    name  = "oauth2"
    value = jsonencode({
      "properties": {
        "providerName": {
          "value": "generic"
        }
      }
    })
  }

  property {
    name  = "password"
    value = ""  # TODO: Sensitive value masked
  }
}

# Flows (references connectors and variables)
resource "pingone_davinci_flow" "signin_flow" {
  environment_id = var.environment_id
  name           = "Sign-In Flow"
  description    = "Primary user authentication flow"

  connection_link {
    id   = pingone_davinci_connector_instance.httpconnector_abc123.id
    name = "HTTP Connector"
  }

  graph_data = jsonencode({
    # ... flow graph data ...
  })
}

# Applications (standalone)
resource "pingone_davinci_application" "web_app" {
  environment_id = var.environment_id
  name           = "Web Application"
  oauth {
    enabled = true
  }
}

# Flow Policies (references flows and applications)
resource "pingone_davinci_application_flow_policy_assignment" "web_app_policy" {
  environment_id = var.environment_id
  application_id = pingone_davinci_application.web_app.id
  flow_policy_id = pingone_davinci_flow_policy.signin_policy.id
  priority       = 1
}
```

### davinci-to-hcl Output

Flow JSON file conversion:

```hcl
resource "pingone_davinci_flow" "simple_demo_flow" {
  environment_id = var.environment_id
  name           = "Simple Demo Flow"
  description    = "A simple flow for demonstration"

  graph_data = jsonencode({
    "elements": {
      "nodes": [
        {
          "data": {
            "id": "httpNode",
            "nodeType": "CONNECTION",
            "connectionId": "conn-abc-123",
            "connectorId": "httpConnector",
            "capabilityName": "customHtmlMessage",
            "properties": {
              "message": {
                "value": "Welcome to the flow!"
              }
            }
          }
        }
      ],
      "edges": [
        {
          "data": {
            "id": "edge1",
            "source": "httpNode",
            "target": "evalNode"
          }
        }
      ]
    }
  })

  settings = jsonencode({
    "csp": "default-src 'self';",
    "logLevel": 4
  })
}
```

## Real-World Usage

### Complete Environment Export

Export all DaVinci resources from a production environment:

```bash
# Set credentials
export PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID="worker-env-id"
export PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID="prod-env-id"
export PINGCLI_PINGONE_WORKER_CLIENT_ID="client-id"
export PINGCLI_PINGONE_WORKER_CLIENT_SECRET="client-secret"
export PINGCLI_PINGONE_REGION_CODE="NA"

# Export (import blocks generated by default)
davinci-convert export --out production-environment.tf

# Review output
# Example output: 62 resources, 62 import blocks, 10,534 lines HCL
# - 5 variables
# - 20 connector instances
# - 30 flows
# - 3 applications
# - 4 flow policies

# Import all resources into Terraform state
cd terraform-workspace/
terraform init
terraform apply  # Automatically imports all 62 resources
```

### Selective Export with Skip Dependencies

Export for testing without full environment setup:

```bash
davinci-convert export \
  --pingone-worker-environment-id "dev-env" \
  --skip-dependencies \
  --out standalone-test.tf

# Produces fully standalone HCL with hardcoded UUIDs
# Can be applied immediately without variables.tf
```

### Incremental Development Workflow

1. **Export baseline:**
   ```bash
   davinci-convert export --out baseline.tf
   ```

2. **Make changes in DaVinci UI**

3. **Export updated state:**
   ```bash
   davinci-convert export --out updated.tf
   ```

4. **Compare:**
   ```bash
   diff baseline.tf updated.tf
   ```

5. **Apply to Terraform:**
   ```bash
   terraform import pingone_davinci_flow.new_flow <flow-id>
   terraform plan
   ```

## Troubleshooting

### Authentication Issues

**Problem:** `Authentication failed`

**Solutions:**
- Verify worker environment ID is correct
- Verify OAuth2 client credentials
- Ensure client has `PingOne API` scope
- Verify region code matches environment region

**Check credentials:**
```bash
echo $PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID
echo $PINGCLI_PINGONE_WORKER_CLIENT_ID
```

### Missing Dependencies

**Problem:** TODO placeholders in generated HCL

**Cause:** Resources reference external dependencies not exported

**Solutions:**
- Export from complete environment (not partial)
- Manually add missing resources
- Use `--skip-dependencies` for testing

### API Rate Limiting

**Problem:** Export fails with rate limit errors

**Solutions:**
- Reduce concurrency (future feature)
- Add delays between requests
- Contact Ping support for rate limit increase

### Sensitive Data Warnings

**Problem:** Connector properties show `# TODO: Sensitive value masked`

**Cause:** Passwords and secrets are masked for security

**Solution:** Manually populate sensitive values after generation:
```hcl
property {
  name  = "password"
  value = var.http_connector_password  # Use Terraform variable
}
```

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - Implementation status and completed work
- [docs/KNOWN_LIMITATIONS.md](docs/KNOWN_LIMITATIONS.md) - Current limitations and workarounds
- [docs/SKIP_DEPENDENCIES.md](docs/SKIP_DEPENDENCIES.md) - Skip-dependencies mode details

## Examples

Example flow files in `.github/prompts/`:

- `simple-demo-flow.json` - Simple single flow
- `PingOne_Sign On with Sessions_multiflow.json` - Real production multi-flow export

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for development environment setup, testing guidelines, and contribution process.

### Quick Start

```bash
# Install dependencies
make deps

# Run tests
make test

# Build binary
make build

# Run linters
make lint

# Full development check
make devcheck
```

## References

- [PingCLI](https://github.com/pingidentity/pingcli)
- [PingOne Go SDK](https://github.com/pingidentity/pingone-go-client)
- [PingOne Terraform Provider](https://github.com/pingidentity/terraform-provider-pingone)
- [DaVinci Documentation](https://docs.pingidentity.com/davinci/)
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)

## License

Copyright © 2025 Ping Identity Corporation

See LICENSE file for details.


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

Example flow files and scripts are provided in `examples/`:

- `01-convert-single-flow.sh` - Convert standalone flow JSON to HCL
- `02-convert-multiflow.sh` - Convert complex multi-flow JSON to HCL
- `03-export-environment.sh` - Export complete environment from API
- `04-export-skip-dependencies.sh` - Export standalone resources
- `05-import-blocks-usage.sh` - Terraform import blocks demonstration

Example flow JSON files in `.github/prompts/`:

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
