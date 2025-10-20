---
mode: agent
---

# Part 5: Final Integration and Error Handling

**Status**: ⏳ NOT STARTED (After Parts 2-4 complete)

**Prerequisites**:
- Part 2 (conversion logic) complete
- Part 3 (API export) complete
- Part 4 (dependency resolution) complete

**Goal**: Integrate all components into production-ready CLI tool with robust error handling.

---

## Phase 5.1: Ping CLI Plugin Integration

**Goal**: Finalize integration with Ping CLI as a plugin command.

### Current Implementation Status

The DaVinci converter is already implemented as a Ping CLI plugin:

**Implemented**:
- ✅ Plugin structure using `hashicorp/go-plugin`
- ✅ gRPC-based communication with host
- ✅ Command metadata (Use, Short, Long, Example)
- ✅ Logger integration via `grpc.Logger` interface

**Current Architecture**:
```go
// cmd/convert.go implements grpc.PingCliCommand interface

type DaVinciConvertCommand struct{}

// Configuration returns command metadata to Ping CLI host
func (c *DaVinciConvertCommand) Configuration() (*grpc.PingCliCommandConfiguration, error)

// Run executes the command with args and logger from host
func (c *DaVinciConvertCommand) Run(args []string, logger grpc.Logger) error
```

### Logger Integration Requirements

The Ping CLI logger must be used for ALL user-facing output:

**Logger Methods Available**:
```go
// grpc.Logger interface methods:
logger.Message(message string, metadata map[string]string) error  // Standard output
logger.Warn(message string, metadata map[string]string) error     // Warnings
logger.PluginError(message string, metadata map[string]string) error  // Errors
```

**DO NOT use**:
- `fmt.Println()` - bypasses plugin logging
- `fmt.Fprintf(os.Stderr, ...)` - bypasses plugin logging
- `log.Printf()` - bypasses plugin logging

**Current Issues to Fix**:

1. **Validation reports go to stderr** - Should use logger instead:
```go
// CURRENT (internal/exporter/orchestrator.go):
fmt.Fprintln(os.Stderr, report)

// SHOULD BE:
logger.Message(report, nil)
```

2. **Progress messages not logged**:
```go
// ADD to exporter:
logger.Message("Fetching flows...", nil)
logger.Message(fmt.Sprintf("✓ Found %d flows", len(flows)), nil)
```

3. **Error context lost** - Add metadata:
```go
// CURRENT:
return fmt.Errorf("failed to fetch flows: %w", err)

// ENHANCED:
logger.PluginError("Failed to fetch flows", map[string]string{
    "environment_id": environmentID,
    "error": err.Error(),
})
return err
```

### Implementation Tasks

**Task 1: Pass Logger Through Call Stack**

Update function signatures to accept logger:

```go
// Before
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool) (string, error)

// After
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool, logger grpc.Logger) (string, error)
```

Propagate through:
- `internal/exporter/orchestrator.go` - Main export function
- All exporter methods that need progress reporting
- Validation report generation

**Task 2: Replace Direct Output with Logger**

In `internal/exporter/orchestrator.go`:

```go
// Before
fmt.Fprintln(os.Stderr, "Dependency Graph Validation Report")
fmt.Fprintln(os.Stderr, strings.Repeat("=", 60))
fmt.Fprintln(os.Stderr, fmt.Sprintf("Total Resources: %d", totalResources))

// After
report := GenerateValidationReport(graph, missingTracker, todoCount)
if err := logger.Message(report, nil); err != nil {
    return "", fmt.Errorf("failed to log validation report: %w", err)
}
```

**Task 3: Add Progress Reporting**

Enhance export process with progress updates:

```go
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool, logger grpc.Logger) (string, error) {
    graph := resolver.NewDependencyGraph()
    missingTracker := resolver.NewMissingDependencyTracker()
    
    // Log start
    logger.Message("Exporting DaVinci environment...", nil)
    
    // Export variables
    logger.Message("Fetching variables...", nil)
    variables, err := client.ReadAllVariables(ctx)
    if err != nil {
        logger.PluginError("Failed to fetch variables", map[string]string{"error": err.Error()})
        return "", err
    }
    logger.Message(fmt.Sprintf("✓ Found %d variables", len(variables)), nil)
    
    // Export connectors
    logger.Message("Fetching connector instances...", nil)
    connectors, err := client.ReadAllConnectorInstances(ctx)
    if err != nil {
        logger.PluginError("Failed to fetch connectors", map[string]string{"error": err.Error()})
        return "", err
    }
    logger.Message(fmt.Sprintf("✓ Found %d connector instances", len(connectors)), nil)
    
    // ... continue for all resource types
    
    // Log validation
    logger.Message("\nValidating dependency graph...", nil)
    if err := graph.ValidateGraph(); err != nil {
        logger.Warn(fmt.Sprintf("Validation warnings: %v", err), nil)
    }
    
    // Log completion
    logger.Message("\n✓ Export complete", map[string]string{
        "resources": fmt.Sprintf("%d", totalResources),
        "todos": fmt.Sprintf("%d", todoCount),
    })
    
    return finalHCL, nil
}
```

**Task 4: Structured Error Reporting**

Use metadata for machine-readable error context:

```go
// Authentication errors
logger.PluginError("Authentication failed", map[string]string{
    "environment_id": environmentID,
    "region": region,
    "hint": "Verify client credentials and permissions",
})

// Network errors
logger.PluginError("Network timeout", map[string]string{
    "endpoint": endpoint,
    "region": region,
    "hint": "Check internet connection and region setting",
})

// API errors
logger.PluginError("API request failed", map[string]string{
    "status_code": fmt.Sprintf("%d", statusCode),
    "endpoint": endpoint,
    "error": errMsg,
})
```

**Task 5: Verbose Mode Integration**

Add verbose logging controlled by Ping CLI flags:

```go
// In cmd/convert.go, check for verbose flag
verbose := false
for i, arg := range args {
    if arg == "--verbose" || arg == "-v" {
        verbose = true
        args = append(args[:i], args[i+1:]...) // Remove flag
        break
    }
}

// Pass verbose to exporter
hcl, err := exporter.ExportEnvironment(ctx, client, skipDeps, logger, verbose)

// In exporter, log details only if verbose
if verbose {
    logger.Message(fmt.Sprintf("Processing flow: %s (ID: %s)", flow.Name, flow.FlowID), nil)
    logger.Message(fmt.Sprintf("  Nodes: %d", len(flow.GraphData.Elements.Nodes)), nil)
    logger.Message(fmt.Sprintf("  Dependencies: %d", len(dependencies)), nil)
}
```

### Testing Logger Integration

**Unit Tests**:

Create mock logger for testing:
```go
// internal/exporter/orchestrator_test.go

type mockLogger struct {
    messages []string
    warnings []string
    errors   []string
}

func (m *mockLogger) Message(msg string, metadata map[string]string) error {
    m.messages = append(m.messages, msg)
    return nil
}

func (m *mockLogger) Warn(msg string, metadata map[string]string) error {
    m.warnings = append(m.warnings, msg)
    return nil
}

func (m *mockLogger) PluginError(msg string, metadata map[string]string) error {
    m.errors = append(m.errors, msg)
    return nil
}

func TestExportEnvironment_LoggerIntegration(t *testing.T) {
    logger := &mockLogger{}
    ctx := context.Background()
    client := setupMockClient()
    
    hcl, err := ExportEnvironment(ctx, client, false, logger)
    require.NoError(t, err)
    
    // Verify progress messages logged
    assert.Contains(t, logger.messages[0], "Exporting DaVinci environment")
    assert.Contains(t, logger.messages[1], "Fetching variables")
    assert.Contains(t, logger.messages[len(logger.messages)-1], "Export complete")
    
    // Verify no direct stdout/stderr output
    // (would fail if fmt.Println used)
}
```

**Integration Tests**:

Test with actual Ping CLI plugin framework:
```go
// cmd/convert_integration_test.go

func TestPluginCommand_LoggerOutput(t *testing.T) {
    // Setup plugin client
    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig: grpc.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            grpc.ENUM_PINGCLI_COMMAND_GRPC: &grpc.PingCliCommandGrpcPlugin{},
        },
        Cmd: exec.Command("./davinci-convert"),
    })
    defer client.Kill()
    
    // Get RPC client
    rpcClient, err := client.Client()
    require.NoError(t, err)
    
    // Get plugin interface
    raw, err := rpcClient.Dispense(grpc.ENUM_PINGCLI_COMMAND_GRPC)
    require.NoError(t, err)
    
    cmd := raw.(grpc.PingCliCommand)
    
    // Capture logger output
    logOutput := &strings.Builder{}
    logger := &testLogger{writer: logOutput}
    
    // Run command
    err = cmd.Run([]string{"--flow-json", "test.json"}, logger)
    require.NoError(t, err)
    
    // Verify all output went through logger
    output := logOutput.String()
    assert.Contains(t, output, "Converting flow")
    assert.Contains(t, output, "✓ Conversion complete")
}
```

### Success Criteria

- ✅ All user-facing output uses `grpc.Logger`
- ✅ No direct `fmt.Println()` or `fmt.Fprintf(os.Stderr)` calls
- ✅ Progress messages logged for all major operations
- ✅ Error messages include structured metadata
- ✅ Validation reports logged through logger
- ✅ Verbose mode provides detailed logging
- ✅ Logger integration tested with mock logger
- ✅ Plugin works correctly in Ping CLI environment

---

## Phase 5.2: Complete CLI Integration

**Goal**: Wire all components together in command implementation.

### Support Two Modes

The command already supports both modes - needs logger integration:

**File Mode** (already implemented):
```go
func (c *DaVinciConvertCommand) runFileMode(logger grpc.Logger, flowJSON, out string, skipDeps bool) error {
    // Log start
    logger.Message("Converting DaVinci flow from file...", nil)
    
    // Read flow from file
    data, err := os.ReadFile(flowJSON)
    if err != nil {
        logger.PluginError("Failed to read flow file", map[string]string{
            "file": flowJSON,
            "error": err.Error(),
        })
        return fmt.Errorf("failed to read flow file: %w", err)
    }
    
    // Convert single flow
    hcl, err := converter.Convert(string(data))
    if err != nil {
        logger.PluginError("Conversion failed", map[string]string{"error": err.Error()})
        return fmt.Errorf("conversion failed: %w", err)
    }
    
    // Log success
    logger.Message("✓ Conversion complete", nil)
    
    // Output HCL
    return writeOutput(hcl, out, logger)
}
```

**Export Mode** (already implemented - needs logger enhancement):
```go
func (c *DaVinciConvertCommand) runExportMode(logger grpc.Logger, environmentID, region, clientID, clientSecret, out string, skipDeps bool) error {
    // Validate credentials (already implemented)
    // Log progress (already implemented)
    
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
    
    // Export all resources - NEEDS LOGGER PARAMETER
    hcl, err := exporter.ExportEnvironment(ctx, client, skipDeps, logger) // ADD logger
    if err != nil {
        logger.PluginError("Export failed", map[string]string{
            "environment_id": environmentID,
            "error": err.Error(),
        })
        return fmt.Errorf("failed to export environment: %w", err)
    }
    
    // Log completion
    logger.Message("✓ Export complete", nil)
    
    // Output HCL
    return writeOutput(hcl, out, logger)
}
```

Mode detection:
- If `--flow-json` provided → file mode
- If `--export` provided → export mode
- Error if both or neither provided

### Handle Output

Support two output methods:

**1. Single file output**:
```go
func writeOutput(hcl string) error {
    if outputPath != "" {
        // Write to file
        return os.WriteFile(outputPath, []byte(hcl), 0644)
    } else {
        // Print to stdout
        fmt.Println(hcl)
        return nil
    }
}
```

**2. Multi-file output** (optional enhancement):
```go
if outputDir != "" {
    // Write separate files per resource type
    // flows.tf, connections.tf, variables.tf, etc.
    return writeMultiFileOutput(exportResult, outputDir)
}
```

### Error Handling and User Experience

Provide clear, actionable error messages:

```go
// Missing required flags
if exportMode && environmentID == "" {
    return fmt.Errorf("--environment-id is required for export mode. Use --help for usage.")
}

// Invalid credentials
if err == api.ErrUnauthorized {
    return fmt.Errorf("authentication failed. Check your credentials:\n" +
        "  1. Verify client ID and secret\n" +
        "  2. Ensure client has DaVinci access\n" +
        "  3. Check environment ID is correct")
}

// Network errors
if err == api.ErrNetworkTimeout {
    return fmt.Errorf("network timeout. Please check:\n" +
        "  1. Internet connection\n" +
        "  2. Region is correct (use --region)\n" +
        "  3. Try again in a few moments")
}

// API rate limiting
if err == api.ErrRateLimited {
    return fmt.Errorf("API rate limit exceeded. Please wait 60 seconds and try again.")
}

// Invalid JSON
if _, ok := err.(*json.SyntaxError); ok {
    return fmt.Errorf("invalid JSON in flow file: %w\nEnsure file contains valid DaVinci flow JSON", err)
}
```

### Additional Flags

**Verbose logging**:
```go
--verbose, -v    Enable detailed logging
```

Implementation:
- Log API requests and responses (sanitize sensitive data)
- Log dependency resolution steps
- Log resource processing

**Dry run**:
```go
--dry-run    Show what would be exported without making API calls
```

Implementation:
- Validate flags and credentials
- Print summary of what would be exported
- Don't make actual API calls

---

## Phase 5.2: Comprehensive Integration Tests

**Goal**: Verify all components work together correctly.

### Command Integration Tests

Create `cmd/convert_integration_test.go`:

**File Mode Tests**:
```go
func TestConvertCommand_FileMode_ValidFlow(t *testing.T) {
    // Given: Valid flow JSON file
    // When: Run convert command with --flow-json
    // Then: HCL written to stdout or file
}

func TestConvertCommand_FileMode_MissingFile(t *testing.T) {
    // Given: Non-existent file path
    // When: Run convert command
    // Then: Error message about missing file
}

func TestConvertCommand_FileMode_InvalidJSON(t *testing.T) {
    // Given: File with malformed JSON
    // When: Run convert command
    // Then: Error message about invalid JSON
}

func TestConvertCommand_FileMode_OutputToFile(t *testing.T) {
    // Given: Valid flow and --out flag
    // When: Run convert command
    // Then: HCL written to specified file
}
```

**Export Mode Tests** (with mock API):
```go
func TestConvertCommand_ExportMode_Success(t *testing.T) {
    // Given: Mock API with complete environment
    // When: Run convert command with --export
    // Then: Complete HCL with all resource types
}

func TestConvertCommand_ExportMode_MissingCredentials(t *testing.T) {
    // Given: No credentials provided
    // When: Run convert command with --export
    // Then: Helpful error about authentication
}

func TestConvertCommand_ExportMode_APIError(t *testing.T) {
    // Given: Mock API returns error
    // When: Run convert command
    // Then: Graceful error handling
}

func TestConvertCommand_ExportMode_DependencyResolution(t *testing.T) {
    // Given: Resources with dependencies
    // When: Run convert command
    // Then: Terraform references generated (not hardcoded IDs)
}
```

**Selective Export Tests** (Phase 3.6):
```go
func TestConvertCommand_SelectiveExport_IncludeFlows(t *testing.T) {
    // Given: --include-flows flag with flow names
    // When: Run convert command
    // Then: Only specified flows exported
}

func TestConvertCommand_SelectiveExport_WithDependencies(t *testing.T) {
    // Given: --include-flows with --with-dependencies
    // When: Run convert command
    // Then: Flow and all dependencies exported (via HAL links)
}

func TestConvertCommand_SelectiveExport_NoDependencies(t *testing.T) {
    // Given: --include-flows with --no-dependencies
    // When: Run convert command
    // Then: Only flow exported, dependencies have TODO placeholders
}

func TestConvertCommand_SelectiveExport_Exclude(t *testing.T) {
    // Given: --exclude-flows flag
    // When: Run convert command
    // Then: All except excluded flows exported
}

func TestConvertCommand_SelectiveExport_InvalidFilters(t *testing.T) {
    // Given: Conflicting include/exclude flags
    // When: Run convert command
    // Then: Helpful error about invalid combination
}
```

### End-to-End Tests

Create `test/e2e_test.go`:

**Complete Export Test**:
```go
func TestEndToEnd_FullExport(t *testing.T) {
    // Setup: Mock environment with all resource types
    mockEnv := setupMockEnvironment(t)
    
    // Export: Run full export
    result, err := runExport(mockEnv)
    require.NoError(t, err)
    
    // Verify: Complete HCL generated
    assert.Contains(t, result, "resource \"pingone_davinci_flow\"")
    assert.Contains(t, result, "resource \"pingone_davinci_connection\"")
    assert.Contains(t, result, "resource \"pingone_davinci_variable\"")
    assert.Contains(t, result, "resource \"pingone_davinci_application\"")
    assert.Contains(t, result, "resource \"pingone_davinci_application_flow_policy\"")
    
    // Verify: Dependencies resolved
    assert.Contains(t, result, "pingone_davinci_connection.")  // Reference, not ID
    assert.NotContains(t, result, "abc123-def456")  // No hardcoded IDs
    
    // Verify: Terraform validation (if terraform available)
    if isTerraformInstalled() {
        validateWithTerraform(t, result)
    }
}
```

**Selective Export Test** (Phase 3.6):
```go
func TestEndToEnd_SelectiveExportWithDependencies(t *testing.T) {
    // Setup: Environment with Registration flow and dependencies
    mockEnv := setupComplexEnvironment(t)
    
    // Export: Single flow with dependencies
    result, err := runExportWithFilter(mockEnv, FilterOptions{
        IncludeFlows: []string{"Registration Flow"},
        WithDependencies: true,
    })
    require.NoError(t, err)
    
    // Verify: Flow exported
    assert.Contains(t, result, "resource \"pingone_davinci_flow\" \"registration_flow\"")
    
    // Verify: Dependencies discovered via HAL links
    assert.Contains(t, result, "resource \"pingone_davinci_connection\"")  // Used by flow
    assert.Contains(t, result, "resource \"pingone_davinci_variable\"")  // Used by flow
    
    // Verify: Unrelated resources NOT exported
    assert.NotContains(t, result, "login_flow")  // Different flow
    
    // Verify: TODO comments for excluded dependencies (if any)
    if hasExcludedDependencies(result) {
        assert.Contains(t, result, "# TODO: Reference")
    }
}
```

**Dependency Resolution Test**:
```go
func TestEndToEnd_DependencyResolution(t *testing.T) {
    // Setup: Complex dependency graph
    mockEnv := setupDependencyGraph(t)
    
    // Export
    result, err := runExport(mockEnv)
    require.NoError(t, err)
    
    // Verify: All hardcoded IDs replaced with references
    lines := strings.Split(result, "\n")
    for _, line := range lines {
        if strings.Contains(line, "_id = ") {
            // Should be reference or TODO comment
            assert.True(t, 
                strings.Contains(line, "pingone_davinci_") || 
                strings.Contains(line, "# TODO:"),
                "Found hardcoded ID in line: %s", line)
        }
    }
    
    // Verify: Resource order respects dependencies
    flowPos := findResourcePosition(result, "pingone_davinci_flow")
    connPos := findResourcePosition(result, "pingone_davinci_connection")
    assert.Less(t, connPos, flowPos, "Connections should come before flows")
}
```

---

## Phase 5.3: Documentation and Examples

**Goal**: Enable users to effectively use the tool.

### Update README

Add to `README.md`:

**File Mode Usage**:
```markdown
## Usage

### Convert Single Flow

Convert a DaVinci flow JSON file to Terraform HCL:

```bash
davinci-convert --flow-json flow.json --out flow.tf
```

Or print to stdout:

```bash
davinci-convert --flow-json flow.json
```

**Export Mode Usage**:
```markdown
### Export Entire Environment

Export all DaVinci resources from an environment:

```bash
davinci-convert --export \
  --environment-id "abc123-def456" \
  --region NA \
  --client-id "your-client-id" \
  --client-secret "your-client-secret" \
  --out davinci.tf
```

Or use config profile:

```bash
davinci-convert --export --profile production --out davinci.tf
```

**Selective Export** (Phase 3.6):
```markdown
### Selective Export

Export specific resources with dependencies:

```bash
# Export single flow with all dependencies
davinci-convert --export \
  --include-flows "Registration Flow" \
  --profile production \
  --out registration.tf

# Export multiple flows without dependencies
davinci-convert --export \
  --include-flows "flow1,flow2" \
  --no-dependencies \
  --profile production

# Export all except test resources
davinci-convert --export \
  --exclude-flows "test" \
  --exclude-applications "test" \
  --profile production
```

**Authentication Setup**:
```markdown
## Authentication

### Option 1: Command Flags

```bash
davinci-convert --export \
  --environment-id "..." \
  --region NA \
  --client-id "..." \
  --client-secret "..."
```

### Option 2: Environment Variables

```bash
export PINGONE_ENVIRONMENT_ID="..."
export PINGONE_REGION="NA"
export PINGONE_CLIENT_ID="..."
export PINGONE_CLIENT_SECRET="..."

davinci-convert --export
```

### Option 3: Config File

Create `~/.pingcli/config.yaml`:

```yaml
profiles:
  production:
    environmentID: "abc123-def456"
    region: "NA"
    clientID: "your-client-id"
    clientSecret: "your-client-secret"
```

Use profile:

```bash
davinci-convert --export --profile production
```

**Using Generated HCL**:
```markdown
## Using Generated Terraform

### 1. Initialize Terraform

```bash
terraform init
```

### 2. Update Sensitive Values

Generated HCL masks sensitive values (passwords, secrets, API keys).
Update TODO comments with actual values:

```hcl
property {
  name  = "apiKey"
  value = ""  # TODO: Sensitive value masked - update with actual credential
}
```

### 3. Apply Configuration

```bash
terraform plan
terraform apply
```

### Create Example Configurations

Create `examples/` directory:

**examples/single-flow-conversion.sh**:
```bash
#!/bin/bash
# Example: Convert single flow file

davinci-convert \
  --flow-json examples/flows/registration-flow.json \
  --out examples/outputs/registration.tf
```

**examples/full-environment-export.sh**:
```bash
#!/bin/bash
# Example: Export entire environment

davinci-convert --export \
  --environment-id "${PINGONE_ENV_ID}" \
  --region NA \
  --profile production \
  --out examples/outputs/complete-environment.tf
```

**examples/selective-export-single-flow.sh** (Phase 3.6):
```bash
#!/bin/bash
# Example: Export single flow with dependencies

davinci-convert --export \
  --include-flows "Registration Flow" \
  --with-dependencies \
  --profile production \
  --out examples/outputs/registration-with-deps.tf
```

**examples/selective-export-no-deps.sh** (Phase 3.6):
```bash
#!/bin/bash
# Example: Export application without dependencies

davinci-convert --export \
  --include-applications "My App" \
  --no-dependencies \
  --profile production \
  --out examples/outputs/app-only.tf
```

**examples/provider.tf**:
```hcl
terraform {
  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = "~> 1.0"
    }
  }
}

provider "pingone" {
  client_id      = var.client_id
  client_secret  = var.client_secret
  environment_id = var.environment_id
  region         = var.region
}
```

**examples/variables.tf**:
```hcl
variable "client_id" {
  type        = string
  description = "PingOne OAuth client ID"
}

variable "client_secret" {
  type        = string
  description = "PingOne OAuth client secret"
  sensitive   = true
}

variable "environment_id" {
  type        = string
  description = "PingOne environment ID"
}

variable "region" {
  type        = string
  description = "PingOne region"
  default     = "NA"
}
```

### Troubleshooting Guide

Add to README:

```markdown
## Troubleshooting

### "Authentication failed"

**Problem**: Invalid credentials or insufficient permissions.

**Solutions**:
1. Verify client ID and secret are correct
2. Ensure OAuth client has DaVinci scope
3. Check environment ID is correct
4. Verify region matches environment

### "Network timeout"

**Problem**: Cannot reach PingOne API.

**Solutions**:
1. Check internet connection
2. Verify region is correct (NA, EU, AP, CA)
3. Check firewall/proxy settings
4. Try again in a few moments

### "Circular dependencies detected"

**Problem**: Resources have circular references (A → B → A).

**Solutions**:
1. Review dependency error output
2. Manually break circular dependencies in DaVinci
3. Use lifecycle blocks in Terraform to manage

### "TODO comments in output"

**Problem**: Missing resource references.

**Reasons**:
- Resource was deleted but still referenced
- Selective export excluded dependency
- Resource in different environment

**Solutions**:
1. Review TODO comments for details
2. Update with correct resource references
3. Use `--with-dependencies` flag for selective exports
4. Remove orphaned references in DaVinci

### HAL Link Parsing Issues (Phase 3.6)

**Problem**: Dependencies not discovered correctly.

**Solutions**:
1. Verify API responses include `_links` sections
2. Check API version compatibility
3. Use `--verbose` flag to see HAL link parsing
4. Report issue with example API response
```

---

## Phase 5.4: Performance and Optimization

**Goal**: Ensure tool performs well with large environments.

### Concurrent API Calls

Implement parallel fetching:

```go
func (e *Exporter) Export(ctx context.Context) (*ExportResult, error) {
    result := &ExportResult{}
    
    // Fetch resource types concurrently
    var wg sync.WaitGroup
    var flowsErr, connectionsErr, variablesErr error
    
    wg.Add(3)
    
    // Fetch flows
    go func() {
        defer wg.Done()
        result.Flows, flowsErr = e.client.ListFlows(ctx)
    }()
    
    // Fetch connections
    go func() {
        defer wg.Done()
        result.ConnectorInstances, connectionsErr = e.client.ListConnectorInstances(ctx)
    }()
    
    // Fetch variables
    go func() {
        defer wg.Done()
        result.Variables, variablesErr = e.client.ListVariables(ctx)
    }()
    
    wg.Wait()
    
    // Check errors
    if flowsErr != nil {
        return nil, fmt.Errorf("failed to fetch flows: %w", flowsErr)
    }
    // ... check other errors
    
    return result, nil
}
```

Implement rate limiting:
```go
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond),
    }
}

func (r *RateLimiter) Wait(ctx context.Context) error {
    return r.limiter.Wait(ctx)
}
```

### Caching

Implement response caching for development:

```go
type CachedClient struct {
    client *api.Client
    cache  map[string][]byte
    cacheFile string
}

func (c *CachedClient) ListFlows(ctx context.Context) ([]Flow, error) {
    // Check cache first
    if cached, ok := c.cache["flows"]; ok {
        var flows []Flow
        json.Unmarshal(cached, &flows)
        return flows, nil
    }
    
    // Fetch from API
    flows, err := c.client.ListFlows(ctx)
    if err != nil {
        return nil, err
    }
    
    // Cache response
    data, _ := json.Marshal(flows)
    c.cache["flows"] = data
    
    return flows, nil
}

func (c *CachedClient) SaveCache() error {
    // Write cache to disk
}

func (c *CachedClient) LoadCache() error {
    // Load cache from disk
}
```

Usage:
```bash
# Cache responses for development
davinci-convert --export --profile dev --cache

# Use cached responses (no API calls)
davinci-convert --export --use-cache --out test.tf
```

### Progress Reporting

For large exports, show progress:

```go
func (e *Exporter) Export(ctx context.Context) (*ExportResult, error) {
    fmt.Println("Exporting DaVinci resources...")
    
    // Fetch flows
    fmt.Print("  Fetching flows... ")
    flows, err := e.client.ListFlows(ctx)
    if err != nil {
        return nil, err
    }
    fmt.Printf("✓ (%d flows)\n", len(flows))
    
    // Fetch connections
    fmt.Print("  Fetching connections... ")
    connections, err := e.client.ListConnectorInstances(ctx)
    if err != nil {
        return nil, err
    }
    fmt.Printf("✓ (%d connections)\n", len(connections))
    
    // ... other resource types
    
    fmt.Println("Export complete!")
    
    return result, nil
}
```

For very large exports, show estimated time:
```go
startTime := time.Now()
totalResources := len(flows) + len(connections) + len(variables)
processed := 0

for _, flow := range flows {
    processFlow(flow)
    processed++
    
    elapsed := time.Since(startTime)
    rate := float64(processed) / elapsed.Seconds()
    remaining := float64(totalResources - processed) / rate
    
    fmt.Printf("\rProcessing: %d/%d (ETA: %.0fs)", processed, totalResources, remaining)
}
```

---

## Success Criteria

### Phase 5.1: Ping CLI Plugin Integration
- ✅ All output uses `grpc.Logger` (no direct fmt.Println/stderr)
- ✅ Progress messages logged for all operations
- ✅ Error messages include structured metadata
- ✅ Validation reports logged through logger
- ✅ Verbose mode implemented
- ✅ Logger mock tests passing
- ✅ Plugin integration tests passing

### Phase 5.2: CLI Integration
- ✅ Both file mode and export mode functional
- ✅ Clear error messages for all common issues
- ✅ Logger integrated throughout call stack
- ✅ Command-line flag parsing working

### Phase 5.3: Integration Tests
- ✅ Integration tests verify all components work together
- ✅ End-to-end tests with mock API pass
- ✅ Plugin framework tests pass

### Phase 5.4: Documentation
- ✅ Documentation covers all usage patterns
- ✅ Examples provided for common scenarios
- ✅ Ping CLI plugin usage documented
- ✅ Troubleshooting guide complete

### Phase 5.5: Performance
- ✅ Performance is acceptable for large environments (100+ flows)
- ✅ Progress reporting works smoothly
- ✅ Concurrent API calls functional (if implemented)

---

**Next Step**: After Part 5 complete, proceed to Part 6 (Production Readiness and Release).
