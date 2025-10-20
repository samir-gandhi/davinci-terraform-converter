# Ping CLI Logger Integration - Implementation Checklist

**Goal**: Replace all direct stdout/stderr output with Ping CLI's gRPC logger

## Current Status

✅ **Already Implemented**:
- Plugin structure using `hashicorp/go-plugin`
- Command metadata (Use, Short, Long, Example)
- Basic logger usage in `cmd/convert.go`
- File mode and export mode structure

❌ **Needs Implementation**:
- Pass logger through orchestrator and exporters
- Replace stderr output with logger calls
- Add progress reporting
- Enhanced error reporting with metadata
- Verbose mode support

## Implementation Tasks

### Task 1: Update Function Signatures ⏳

**Files to modify**:

1. `internal/exporter/orchestrator.go`:
```go
// Before
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool) (string, error)

// After
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool, logger grpc.Logger) (string, error)
```

2. Individual exporters (if they need progress reporting):
```go
// Add logger parameter to any exporter that needs to log progress
func exportFlows(ctx context.Context, client *api.Client, logger grpc.Logger) ([]string, error)
```

### Task 2: Replace Direct Output ⏳

**In `internal/exporter/orchestrator.go`**:

Search for and replace:
```bash
# Find all direct output
grep -n "fmt.Fprintln(os.Stderr" internal/exporter/orchestrator.go
grep -n "fmt.Println" internal/exporter/orchestrator.go
```

Replace patterns:
```go
// Before
fmt.Fprintln(os.Stderr, report)

// After
if err := logger.Message(report, nil); err != nil {
    return "", fmt.Errorf("failed to log report: %w", err)
}
```

**Specific locations**:
- Line ~103: Validation report output
- Line ~115: Missing dependency summary output
- Line ~125: TODO count output

### Task 3: Add Progress Messages ⏳

**In `internal/exporter/orchestrator.go`** add progress for each resource type:

```go
// Variables
logger.Message("Fetching variables...", nil)
variables, err := client.ReadAllVariables(ctx)
if err != nil {
    logger.PluginError("Failed to fetch variables", map[string]string{"error": err.Error()})
    return "", err
}
logger.Message(fmt.Sprintf("✓ Found %d variables", len(variables)), nil)

// Connector instances
logger.Message("Fetching connector instances...", nil)
connectors, err := client.ReadAllConnectorInstances(ctx)
if err != nil {
    logger.PluginError("Failed to fetch connectors", map[string]string{"error": err.Error()})
    return "", err
}
logger.Message(fmt.Sprintf("✓ Found %d connector instances", len(connectors)), nil)

// Flows
logger.Message("Fetching flows...", nil)
flows, err := client.ListFlows(ctx)
if err != nil {
    logger.PluginError("Failed to fetch flows", map[string]string{"error": err.Error()})
    return "", err
}
logger.Message(fmt.Sprintf("✓ Found %d flows", len(flows)), nil)

// Applications
logger.Message("Fetching applications...", nil)
apps, err := client.ReadAllDavinciApplications(ctx)
if err != nil {
    logger.PluginError("Failed to fetch applications", map[string]string{"error": err.Error()})
    return "", err
}
logger.Message(fmt.Sprintf("✓ Found %d applications", len(apps)), nil)

// Flow policies
logger.Message("Fetching flow policies...", nil)
// ... similar pattern
```

### Task 4: Update Command Layer ⏳

**In `cmd/convert.go`**:

Update the call to ExportEnvironment:
```go
// Around line 197
// Before
hcl, err := exporter.ExportEnvironment(ctx, client, skipDeps)

// After
hcl, err := exporter.ExportEnvironment(ctx, client, skipDeps, logger)
```

### Task 5: Add Import Statement ⏳

**In `internal/exporter/orchestrator.go`**:

Add to imports:
```go
import (
    // ... existing imports
    "github.com/pingidentity/pingcli/shared/grpc"
)
```

### Task 6: Implement Verbose Mode ⏳

**In `cmd/convert.go`**:

Parse verbose flag and pass to exporter:
```go
func (c *DaVinciConvertCommand) Run(args []string, logger grpc.Logger) error {
    verbose := false
    
    // Parse flags
    flags := pflag.NewFlagSet("davinci convert", pflag.ContinueOnError)
    // ... existing flag definitions
    flags.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
    
    // ... rest of implementation
    
    // Pass verbose to export
    hcl, err := exporter.ExportEnvironment(ctx, client, skipDeps, logger, verbose)
}
```

**In `internal/exporter/orchestrator.go`**:

Add verbose parameter and log details:
```go
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool, logger grpc.Logger, verbose bool) (string, error) {
    // ... existing code
    
    // When processing resources
    for _, flow := range flows {
        if verbose {
            logger.Message(fmt.Sprintf("Processing flow: %s (ID: %s)", flow.Name, flow.FlowID), map[string]string{
                "nodes": fmt.Sprintf("%d", nodeCount),
                "dependencies": fmt.Sprintf("%d", depCount),
            })
        }
        // ... convert flow
    }
}
```

## Testing Checklist

### Unit Tests ⏳

Create `internal/exporter/orchestrator_logger_test.go`:

```go
package exporter

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

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
    client := setupMockClient(t) // Need to implement
    
    hcl, err := ExportEnvironment(ctx, client, false, logger, false)
    require.NoError(t, err)
    
    // Verify progress messages logged
    assert.Contains(t, logger.messages[0], "Fetching variables")
    assert.Contains(t, logger.messages[1], "✓ Found")
    
    // Verify validation report logged
    foundReport := false
    for _, msg := range logger.messages {
        if strings.Contains(msg, "Validation Report") {
            foundReport = true
            break
        }
    }
    assert.True(t, foundReport, "Validation report should be logged")
}

func TestExportEnvironment_VerboseMode(t *testing.T) {
    logger := &mockLogger{}
    ctx := context.Background()
    client := setupMockClient(t)
    
    hcl, err := ExportEnvironment(ctx, client, false, logger, true)
    require.NoError(t, err)
    
    // Verify verbose messages logged
    verboseCount := 0
    for _, msg := range logger.messages {
        if strings.Contains(msg, "Processing") {
            verboseCount++
        }
    }
    assert.Greater(t, verboseCount, 0, "Verbose mode should log processing details")
}
```

### Manual Testing ⏳

Build and test with Ping CLI:

```bash
# Build the plugin
cd /path/to/davinci-terraform-converter
go build -o davinci-convert .

# Test with Ping CLI (if pingcli is installed)
pingcli davinci convert --flow-json test-flow.json

# Verify:
# - No output to stderr directly
# - All messages appear through Ping CLI's output
# - Progress messages display correctly
# - Error messages are clear and structured
```

## Verification Checklist

Before marking complete:

- [ ] All `fmt.Fprintln(os.Stderr, ...)` removed from orchestrator.go
- [ ] All `fmt.Println(...)` removed (except in main.go if needed)
- [ ] Logger passed to ExportEnvironment function
- [ ] Progress messages added for all major operations
- [ ] Error messages include metadata maps
- [ ] Validation report goes through logger
- [ ] Missing dependency summary goes through logger
- [ ] Verbose mode implemented and functional
- [ ] Mock logger tests passing
- [ ] Manual test with pingcli successful
- [ ] No breaking changes to existing tests

## Estimated Effort

- **Task 1-2**: 30 minutes - Update signatures and replace direct output
- **Task 3**: 45 minutes - Add progress messages throughout
- **Task 4**: 15 minutes - Update command layer
- **Task 5**: 5 minutes - Add import
- **Task 6**: 1 hour - Implement and test verbose mode
- **Testing**: 1 hour - Write tests and manual verification

**Total**: ~3-4 hours

## Files Modified

**Production Code**:
- `cmd/convert.go` - Update ExportEnvironment call, add verbose flag
- `internal/exporter/orchestrator.go` - Add logger parameter, replace output, add progress

**Test Code**:
- `internal/exporter/orchestrator_logger_test.go` - New file with logger tests

**Documentation**:
- `README.md` - Update usage examples showing Ping CLI integration
- `.github/prompts/07-part5-integration.md` - Already updated

## Success Metrics

✅ **Complete when**:
1. No direct stdout/stderr writes in orchestrator
2. All user-facing output through grpc.Logger
3. Progress messages visible during export
4. Error messages include helpful metadata
5. Verbose mode provides detailed logging
6. Tests pass with mock logger
7. Manual test with pingcli works correctly
