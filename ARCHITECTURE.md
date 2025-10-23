# Architecture

This document describes the architecture of the DaVinci Terraform Converter.

## System Overview

The DaVinci Terraform Converter operates in two modes:

1. **Plugin Mode**: Integrates with Ping CLI as a gRPC plugin
2. **Standalone Mode**: Runs as an independent CLI binary

```
┌─────────────────────────────────────────────────────────────┐
│                        Entry Point                          │
│                         (main.go)                           │
│                                                             │
│  Detects: PLUGIN_PROTOCOL_VERSIONS environment variable    │
│    ├─ Set: Run as Plugin                                   │
│    └─ Not Set: Run as Standalone CLI                       │
└─────────────────────────────────────────────────────────────┘
                          │
           ┌──────────────┴──────────────┐
           │                             │
    [Plugin Mode]               [Standalone Mode]
           │                             │
           ▼                             ▼
┌──────────────────────┐      ┌──────────────────────┐
│  HashiCorp go-plugin │      │   Direct Execution   │
│   gRPC Server        │      │   Cobra-like CLI     │
│                      │      │                      │
│  ┌────────────────┐ │      │  ┌────────────────┐ │
│  │  TfCommand     │ │      │  │  TfCommand     │ │
│  │  (Parent)      │ │      │  │  (Parent)      │ │
│  └────────────────┘ │      │  └────────────────┘ │
└──────────────────────┘      └──────────────────┘
           │                             │
           └──────────────┬──────────────┘
                          │
                          ▼
            ┌─────────────────────────┐
            │  Command Router         │
            │  (cmd/tf.go)            │
            └─────────────────────────┘
                      │
         ┌────────────┴────────────┐
         │                         │
         ▼                         ▼
┌──────────────────┐      ┌──────────────────┐
│  DaVinci to HCL  │      │   Export         │
│  (File Conv)     │      │   (API Export)   │
└──────────────────┘      └──────────────────┘
```

## Command Structure

### Parent Command: `tf`

Located in `cmd/tf.go`, this is the entry point for both plugin and standalone modes.

**Responsibilities:**
- Route to appropriate subcommand
- Provide help text
- Handle unknown subcommands

**Usage:**
```bash
# Plugin mode
pingcli tf davinci-to-hcl --flow-json input.json
pingcli tf export --services pingone-davinci

# Standalone mode
./davinci-convert davinci-to-hcl --flow-json input.json
./davinci-convert export --services pingone-davinci
```

### Subcommand: `davinci-to-hcl`

Located in `cmd/davinci_to_hcl.go`, converts DaVinci flow JSON files to HCL.

**Purpose:** File-based conversion for standalone flow JSON files

**Flags:**
- `--flow-json` (required): Path to input JSON file
- `--out` (optional): Output file path (defaults to stdout)
- `--skip-dependencies` (optional): Use hardcoded IDs instead of references

**Process:**
1. Read flow JSON from file
2. Convert to HCL via `converter.ConvertWithOptions()`
3. Write to file or stdout

**Note:** Standalone flow JSON files don't have environment context, so `environment_id = var.environment_id` is always used (even with `--skip-dependencies`).

### Subcommand: `export`

Located in `cmd/export.go`, exports resources from PingOne DaVinci API.

**Purpose:** API-based export of complete environments

**Flags:**
- `--services` (optional): Services to export (defaults to all: `["pingone-davinci"]`)
- `--pingone-worker-environment-id` (required): Worker app environment
- `--pingone-export-environment-id` (optional): Target environment (defaults to worker)
- `--pingone-worker-client-id` (required): OAuth client ID
- `--pingone-worker-client-secret` (required): OAuth client secret
- `--pingone-region-code` (optional): Region (NA, EU, AP, CA, AU) (defaults to NA)
- `--out` (optional): Output file path (defaults to stdout)
- `--skip-dependencies` (optional): Use hardcoded UUIDs

**Process:**
1. Authenticate with worker environment
2. Fetch resources from export environment
3. Build dependency graph
4. Generate HCL with references
5. Write output

## Two-Environment Model

The export command supports a two-environment architecture for security and flexibility:

```
┌──────────────────────────────┐
│  Worker Environment          │
│  (Authentication)            │
│                              │
│  ┌────────────────────────┐ │
│  │  OAuth2 Worker App     │ │
│  │  - Client ID           │ │
│  │  - Client Secret       │ │
│  │  - Grants token        │ │
│  └────────────────────────┘ │
└──────────────────────────────┘
            │
            │ Authenticates
            ▼
┌──────────────────────────────┐
│  PingOne API                 │
└──────────────────────────────┘
            │
            │ Fetches resources
            ▼
┌──────────────────────────────┐
│  Export Environment          │
│  (Target Resources)          │
│                              │
│  ┌────────────────────────┐ │
│  │  DaVinci Resources     │ │
│  │  - Flows               │ │
│  │  - Connections         │ │
│  │  - Variables           │ │
│  │  - Applications        │ │
│  │  - Policies            │ │
│  └────────────────────────┘ │
└──────────────────────────────┘
```

**Benefits:**
- Worker app credentials isolated from exported resources
- Can export from different environments using same worker app
- Supports dev/staging/prod workflows
- Defaults to single-environment for simplicity

**Implementation:**
```go
// Two-environment export
client, err := api.NewClient(ctx, 
    workerEnvironmentID,    // For authentication
    exportEnvironmentID,    // For resource fetch
    regionCode, 
    clientID, 
    clientSecret)
```

## Internal Architecture

### API Client (`internal/api/`)

Wraps the PingOne Go SDK for DaVinci operations.

**Components:**
- `client.go`: API client initialization and configuration
- `flows.go`: Flow operations (list, get)
- Additional resource operations

**Key Features:**
- OAuth2 authentication
- Two-environment support
- Region-aware endpoints
- Pagination handling
- Raw HTTP workaround for SDK issues

### Converter (`internal/converter/`)

Transforms API responses to Terraform HCL.

**Converters:**
- `flow_converter.go`: DaVinci flows
- `variable_converter.go`: Variables
- `connector_instance_converter.go`: Connector instances
- `application_converter.go`: Applications
- `flow_policy_converter.go`: Flow policies

**Process:**
1. Parse API JSON response
2. Extract resource attributes
3. Generate HCL structure
4. Apply dependency references (if enabled)
5. Return formatted HCL string

### Exporter (`internal/exporter/`)

Orchestrates the export process.

**Components:**
- `orchestrator.go`: Main export coordination
- `*_exporter.go`: Resource-specific exporters

**Process:**
1. Fetch all resource types from API
2. Build dependency graph
3. Convert each resource to HCL
4. Generate import blocks (optional, with `--generate-imports`)
5. Order resources by dependencies
6. Validate and generate report
7. Combine into single HCL output

### Import Block Generator (`internal/importgen/`)

Generates Terraform import blocks for automatic state management.

**Purpose:**
Eliminates manual `terraform import` commands by generating import blocks alongside resource definitions. Import blocks are generated by default; use `--skip-imports` flag to disable. Requires Terraform 1.5+.

**Components:**
- `import_generator.go`: Core import block generation logic

**Import ID Formats by Resource Type:**

| Resource Type | Import ID Format | Example |
|--------------|------------------|---------|
| Variable | `<env_id>/<var_id>` | `abc123.../var456...` |
| Connector Instance | `<env_id>/<instance_id>` | `abc123.../conn456...` |
| Flow | `<env_id>/<flow_id>` | `abc123.../flow456...` |
| Application | `<env_id>/<app_id>` | `abc123.../app456...` |
| Flow Policy | `<env_id>/<policy_id>` | `abc123.../pol456...` |
| Flow Policy Assignment | `<env_id>/<app_id>/<policy_id>` | `abc123.../app456.../pol789...` |

**Output Format:**
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
```

**Integration:**
- Called by exporters when `ExportOptions.GenerateImports = true` (default)
- Import blocks placed immediately before resource definitions
- Uses actual environment ID (not `var.environment_id`)
- Special handling for flow policy assignments (3-part ID with metadata)
- Disable with `--skip-imports` flag for Terraform < 1.5 or manual import workflows

### Resolver (`internal/resolver/`)

Manages dependencies between resources.

**Components:**
- `dependency_graph.go`: Graph structure and operations
- `missing_tracker.go`: Tracks missing dependencies
- `reference_generator.go`: Generates Terraform references

**Dependency Types:**
- Flow → Connector Instance
- Flow → Variable
- Flow → Flow (subflows)
- Policy → Flow
- Policy → Application

**Reference Generation:**
```hcl
# Without skip-dependencies
connection_id = pingone_davinci_connector_instance.httpconnector_abc123.id

# With skip-dependencies  
connection_id = "abc123-def456-..."
```

## Plugin System

### gRPC Communication

Uses HashiCorp's go-plugin framework:

```
┌──────────────────┐        gRPC         ┌──────────────────┐
│   Ping CLI       │◄──────────────────►│   Plugin Binary  │
│   (Host)         │                     │   (davinci-      │
│                  │                     │    convert)      │
│  - Spawns plugin │                     │                  │
│  - Sends args    │                     │  - Implements    │
│  - Logs output   │                     │    interface     │
│  - Kills on exit │                     │  - Returns HCL   │
└──────────────────┘                     └──────────────────┘
```

### Logger Integration

All output must go through the gRPC logger in plugin mode:

```go
// ✅ CORRECT
logger.Message("Exporting flows...", nil)
logger.Warn("Validation issues found", metadata)
logger.PluginError("Export failed", metadata)

// ❌ INCORRECT (bypasses plugin)
fmt.Println("Exporting flows...")
fmt.Fprintf(os.Stderr, "Error: ...")
```

**Standalone Mode Exception:**
Stdout output (`fmt.Print(hcl)`) is allowed in standalone mode for piping output. In plugin mode, users should use `--out` flag.

### Command Interface

```go
type PingCliCommand interface {
    // Configuration returns command metadata
    Configuration() (*PingCliCommandConfiguration, error)
    
    // Run executes the command
    Run(args []string, logger Logger) error
}
```

## Service Architecture

Designed for extensibility to other Ping Identity products:

```
pingcli tf export --services <service>
                              │
                ┌─────────────┴─────────────┐
                │                           │
         pingone-davinci              (future services)
                │                           │
        ┌───────┴───────┐              ┌────────────┐
        │               │              │            │
    Flows      Connections         pingone-sso  pingfederate
    Variables  Applications            │             │
    Policies                       Users         Config
                                   Groups        Adapters
                                   Apps          Policies
```

**Current:**
- `pingone-davinci`: Complete implementation

**Future:**
- `pingone-sso`: PingOne SSO resources
- `pingone-authorize`: Authorize policies
- `pingfederate`: PingFederate configuration

**Validation:**
```go
// Service validation in cmd/export.go
supportedServices := []string{"pingone-davinci"}
for _, svc := range services {
    if !contains(supportedServices, svc) {
        return fmt.Errorf("unsupported service: %s", svc)
    }
}
```

## Data Flow

### Export Flow

```
1. User Command
   └─> pingcli tf export --services pingone-davinci

2. Authentication
   └─> OAuth2 token from worker environment

3. Resource Fetch (Parallel)
   ├─> List Variables
   ├─> List Connector Instances  
   ├─> List Flows
   ├─> List Applications
   └─> List Flow Policies

4. Dependency Analysis
   └─> Build graph, track references

5. Conversion
   ├─> Variables (no deps) →  HCL
   ├─> Connectors (no deps) → HCL
   ├─> Flows (refs connectors) → HCL
   ├─> Applications (no deps) → HCL
   └─> Policies (refs flows/apps) → HCL

6. Output
   └─> Combined HCL file or stdout
```

### Dependency Resolution Flow

```
1. Resource Discovery
   └─> Identify all resources and IDs

2. Graph Building
   ├─> Add resources as nodes
   └─> Extract dependencies as edges

3. Reference Generation
   ├─> Sanitize resource names
   ├─> Check graph for target
   ├─> Generate Terraform reference
   └─> Or generate TODO placeholder

4. Validation
   ├─> Check for circular deps
   ├─> Verify all refs resolvable
   └─> Generate validation report
```

## Error Handling

### Authentication Errors

```go
if err == api.ErrUnauthorized {
    logger.PluginError("Authentication failed", map[string]string{
        "environment_id": envID,
        "hint": "Verify credentials and permissions",
    })
}
```

### API Errors

```go
if apiErr, ok := err.(*api.APIError); ok {
    logger.PluginError("API request failed", map[string]string{
        "status_code": apiErr.StatusCode,
        "endpoint": apiErr.Endpoint,
        "error": apiErr.Message,
    })
}
```

### Validation Errors

```go
if validationErr := graph.ValidateGraph(); validationErr != nil {
    logger.Warn("Dependency validation issues", map[string]string{
        "circular_deps": fmt.Sprintf("%d", len(circularDeps)),
        "missing_refs": fmt.Sprintf("%d", missingCount),
    })
}
```

## Configuration

### Environment Variables

Priority: Command-line flags > Environment variables > Defaults

```bash
# Worker authentication
export PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID="..."
export PINGCLI_PINGONE_WORKER_CLIENT_ID="..."
export PINGCLI_PINGONE_WORKER_CLIENT_SECRET="..."
export PINGCLI_PINGONE_REGION_CODE="NA"

# Export target (optional)
export PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID="..."
```

### Defaults

- Region: `NA` (North America)
- Export environment: Defaults to worker environment
- Services: `["pingone-davinci"]` (all available)
- Output: stdout (unless `--out` specified)

## Testing Strategy

### Unit Tests

- Each converter tested independently
- Mock API responses
- Dependency graph validation
- Reference generation accuracy

### Integration Tests

- Real API calls (with test credentials)
- End-to-end export process
- Two-environment scenarios
- Skip-dependencies mode

### Plugin Tests

- gRPC communication
- Logger integration
- Command routing
- Metadata validation

## Performance Considerations

### Current Implementation

- Sequential API calls
- Single-threaded conversion
- In-memory graph building

**Typical Performance:**
- Small environment (10 resources): < 5 seconds
- Medium environment (50 resources): < 15 seconds
- Large environment (100+ resources): < 30 seconds

### Future Optimizations

See [.github/prompts/SCALABILITY_ROADMAP.md](.github/prompts/SCALABILITY_ROADMAP.md):

- Concurrent API fetching
- Streaming conversion
- Caching for development
- Rate limiting
- Progress estimation

## Security

### Credential Handling

- Environment variables prefixed with `PINGCLI_`
- No credentials in logs (sanitized)
- Client secret marked as sensitive
- No credentials in generated HCL

### Sensitive Data Masking

Connector properties with sensitive data are masked:

```hcl
property {
  name  = "password"
  value = ""  # TODO: Sensitive value masked
}
```

Users must manually populate after generation.

## References

- Plugin Framework: https://github.com/hashicorp/go-plugin
- PingOne SDK: https://github.com/pingidentity/pingone-go-client
- Terraform Provider: https://github.com/pingidentity/terraform-provider-pingone
