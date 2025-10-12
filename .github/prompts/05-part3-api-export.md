---
mode: agent
---

# Part 3: API Export Integration

**Status**: ⏳ IN PROGRESS

**Prerequisites**: 
- Part 2 (all phases 2.1-2.6) must be complete ✅
- All converter functions working and tested ✅
- Understanding of PingCLI authentication patterns

**Goal**: Integrate with PingOne DaVinci API to export entire environments, not just single flow files.

**Implementation Approach**: Breaking into smaller reviewable phases with separate files.

---

## Overview

Instead of converting a single flow JSON file, export complete environments including:
- All flows (with versions)
- All connector instances (connections)
- All variables (company and flow contexts)
- All applications
- All flow policies

Use **PingCLI authentication** for credential management and API access.

---

## Authentication and API Client Setup

### Use PingCLI Authentication

**CRITICAL**: Reuse PingCLI's authentication instead of building new credential handling.

Reference `pingcli/cmd/platform/` authentication patterns:
- Credential precedence: flags > env vars > config file > profile
- OAuth token acquisition and refresh
- Config file parsing
- Profile management

### Create API Client Factory

Create `internal/api/client.go`:
```go
package api

import (
    "context"
    "github.com/pingidentity/pingone-go-client/v2/pingone"
)

type Client struct {
    PingOne    *pingone.Client
    EnvironmentID string
}

func NewClient(ctx context.Context, environmentID, region, clientID, clientSecret string) (*Client, error) {
    // Initialize pingone-go-client SDK
    // Handle token acquisition
    // Return configured client
}
```

Use `pingone-go-client` SDK for all API calls.

### Command Flags

Add to command definition (following pingcli conventions):

**Export Mode**:
- `--export`: Enable API export mode (alternative to --flow-json)

**Required Flags** (or from config/env):
- `--environment-id <uuid>`: Target PingOne environment
- `--region <code>`: PingOne region (NA, EU, AP, CA)
- `--client-id <id>`: OAuth client ID
- `--client-secret <secret>`: OAuth client secret

**Optional Flags**:
- `--profile <name>`: Config profile to use (pingcli standard)

Flags should be optional if values available from config file or environment variables.

---

## Phase 3.1: Export Flow Resources

**Goal**: Export all flows from an environment.

### API Client for Flows

Create `internal/api/flows.go`:

**Functions**:
- `ListFlows(ctx context.Context) ([]Flow, error)`: Retrieve all flows
- `GetFlowWithVersions(ctx context.Context, flowID string) (*FlowDetail, error)`: Get flow details including published versions
- Handle pagination for environments with many flows

### Test Flow Export

Create `internal/api/flows_test.go`:

Test using mock HTTP client or test server:
- Successful flow retrieval
- Pagination handling
- Network errors
- Authentication failures
- Rate limiting

### Convert Exported Flows

Update `converter.Convert()` to handle API response format:
- Parse flow data from API response structure
- Generate HCL for all exported flows
- Create resource names based on flow names (sanitized for Terraform)
- Include metadata comments (flow version, last modified, etc.)

---

## Phase 3.2: Export Connector Instances

**Goal**: Export all connections (connector instances).

### API Client for Connector Instances

Create `internal/api/connector_instances.go`:

**Functions**:
- `ListConnectorInstances(ctx context.Context) ([]ConnectorInstance, error)`: Retrieve all connector instances
- Handle sensitive properties (credentials marked as TODO placeholders)

### HCL Generation for Connector Instances

Create `generateConnectorInstanceHCL()` function in converter:

Generate `pingone_davinci_connection` resource blocks:
```hcl
resource "pingone_davinci_connection" "my_connector" {
  environment_id = var.environment_id
  connector_id   = "httpConnector"  # Connector type ID
  name           = "My HTTP Connector"
  
  property {
    name  = "apiUrl"
    value = "https://api.example.com"
  }
  
  property {
    name  = "apiKey"
    value = ""  # TODO: Sensitive value masked - update with actual credential
  }
}
```

Handle connector-specific properties correctly.

Mask sensitive values (passwords, tokens, API keys, etc.) using algorithm from Phase 2.4.

### Test Connector Instance Export

Write tests for:
- API client functionality
- HCL generation for various connector types
- Sensitive value masking
- Connector with many properties

---

## Phase 3.3: Export Variables

**Goal**: Export all company and flow variables.

### API Client for Variables

Create `internal/api/variables.go`:

**Functions**:
- `ListVariables(ctx context.Context) ([]Variable, error)`: Retrieve all variables
- Handle different contexts:
  - `company`: Company-level variables
  - `flow`: Flow-level variables
  - `flowInstance`: Flow instance variables
  - `user`: User variables

### HCL Generation for Variables

Create `generateVariableHCL()` function in converter:

Generate `pingone_davinci_variable` resource blocks:
```hcl
resource "pingone_davinci_variable" "api_key" {
  environment_id = var.environment_id
  name           = "apiKey"
  context        = "company"
  type           = "secret"
  value          = ""  # TODO: Secret value masked - update with actual secret
}

resource "pingone_davinci_variable" "max_retries" {
  environment_id = var.environment_id
  name           = "maxRetries"
  context        = "company"
  type           = "number"
  value          = "3"
}
```

Handle different data types:
- `string`
- `number`
- `boolean`
- `object`
- `secret`

Mask secret variable values.

### Test Variable Export

Write tests for:
- API client functionality
- HCL generation for all contexts
- HCL generation for all data types
- Secret masking

---

## Phase 3.4: Export Applications and Flow Policies

**Goal**: Export DaVinci applications and their flow policies.

### API Client for Applications

Create `internal/api/applications.go`:

**Functions**:
- `ListApplications(ctx context.Context) ([]Application, error)`: Retrieve all DaVinci applications
- `GetApplicationFlowPolicies(ctx context.Context, appID string) ([]FlowPolicy, error)`: Get flow policies for an application

### HCL Generation for Applications

Create `generateApplicationHCL()` function in converter:

Generate `pingone_davinci_application` resource blocks:
```hcl
resource "pingone_davinci_application" "my_app" {
  environment_id = var.environment_id
  name           = "My Application"
  
  oauth {
    enabled = true
    values {
      allowed_grants                = ["authorizationCode"]
      allowed_scopes                = ["openid", "profile"]
      enabled                       = true
      enforce_signed_request_openid = false
      redirect_uris                 = ["https://example.com/callback"]
    }
  }
  
  api_keys {
    enabled = true
  }
}
```

Handle:
- OAuth configuration
- API keys
- SAML configuration
- Policy references

### HCL Generation for Flow Policies

Create `generateFlowPolicyHCL()` function in converter:

Generate `pingone_davinci_application_flow_policy` resource blocks:
```hcl
resource "pingone_davinci_application_flow_policy" "my_app_policy" {
  environment_id = var.environment_id
  application_id = pingone_davinci_application.my_app.id
  name           = "My Application Policy"
  status         = "enabled"
  
  policy_flow {
    flow_id    = pingone_davinci_flow.registration.id
    version_id = -1  # Latest published version
    weight     = 100
    
    success_nodes = ["node123", "node456"]
  }
}
```

Handle:
- Flow distributions (multiple flows)
- Success node configuration
- Trigger settings

### Test Application and Flow Policy Export

Write tests for:
- API client functionality
- HCL generation for various application configurations
- Flow policy generation
- Resource references

---

## Phase 3.5: Orchestration and Output

**Goal**: Coordinate all exports and generate complete HCL output.

### Export Orchestrator

Create `internal/exporter/exporter.go`:

```go
package exporter

type Exporter struct {
    client *api.Client
}

type ExportResult struct {
    Flows             []api.Flow
    ConnectorInstances []api.ConnectorInstance
    Variables         []api.Variable
    Applications      []api.Application
    FlowPolicies      []api.FlowPolicy
}

func (e *Exporter) Export(ctx context.Context) (*ExportResult, error) {
    // Authenticate with PingOne API
    // Export all resource types in parallel (where possible)
    // Collect all exported data
    // Return structured result
}
```

Implement parallel export where possible (flows, connectors, variables can be fetched concurrently).

### Combined HCL Generation

Update converter to accept structured export data:

```go
func (c *Converter) ConvertExport(export *exporter.ExportResult) (string, error) {
    // Generate HCL for all resources in logical order
}
```

Resource generation order (respects dependencies):
1. Variables (may be referenced by other resources)
2. Connector instances (referenced by flows)
3. Flows (may reference variables and connectors)
4. Applications (reference flows)
5. Flow policies (reference flows and applications)

Include header comments:
```hcl
# Generated by davinci-terraform-converter
# Export Date: 2024-01-15T10:30:00Z
# Environment ID: abc123-def456-ghi789
# Region: NA
#
# NOTE: Sensitive values have been masked. Update with actual credentials before applying.
```

### CLI Integration

Update command execution logic:

Detect mode:
```go
if exportMode {
    // Validate required flags
    // Initialize API client
    // Call exporter.Export()
    // Convert export to HCL
} else {
    // Original file mode logic
}
```

Validation for export mode:
- Required: environment-id, region, credentials (or from config)
- Error if --flow-json and --export both provided

Handle errors gracefully:
- Missing credentials: Helpful message about authentication setup
- Network errors: Retry guidance
- API errors: Include error code and message from API
- Rate limiting: Suggest waiting or reducing concurrency

### Integration Tests

Create `internal/exporter/exporter_test.go`:

End-to-end test:
- Use mock API responses for all resource types
- Export all resource types
- Generate complete HCL
- Validate HCL structure
- Verify resource ordering
- Check sensitive value masking

---

## Phase 3.6: Selective Export (Future Enhancement)

**Status**: Future work after basic export (3.1-3.5) working and Part 4 (dependency resolution) complete.

**Goal**: Allow users to export specific resources with optional dependency inclusion.

### Selective Export Flags

Add resource filtering flags:

**Include Flags** (export only specified resources):
- `--include-flows <comma-separated-ids-or-names>`
- `--include-applications <comma-separated-ids-or-names>`
- `--include-connections <comma-separated-ids-or-names>`
- `--include-variables <comma-separated-ids-or-names>`

**Exclude Flags** (exclude specified resources):
- `--exclude-flows <comma-separated-ids-or-names>`
- `--exclude-applications <comma-separated-ids-or-names>`
- `--exclude-connections <comma-separated-ids-or-names>`
- `--exclude-variables <comma-separated-ids-or-names>`

**Dependency Flags**:
- `--skip-dependencies`: Export only selected resources (Default false)

### HAL Link Parsing for Dependency Discovery

**CRITICAL**: PingOne API responses use HAL (Hypertext Application Language) format.

HAL response structure:
```json
{
  "id": "flow123",
  "name": "Registration Flow",
  "_links": {
    "self": { "href": "/environments/env123/flows/flow123" },
    "connectorInstances": { "href": "/environments/env123/connectorInstances?flowId=flow123" },
    "variables": { "href": "/environments/env123/variables?flowId=flow123" },
    "subflows": { "href": "/environments/env123/flows?parentFlowId=flow123" }
  }
}
```

Create `internal/api/hal.go`:

```go
package api

type HALLink struct {
    Href string `json:"href"`
}

type HALLinks map[string]HALLink

func ParseHALLinks(response map[string]interface{}) HALLinks {
    // Extract _links section
    // Parse links
    // Return structured link map
}

func DiscoverDependencies(resource interface{}) ([]Dependency, error) {
    // Parse HAL links from resource
    // Extract resource IDs and types from links
    // Return list of dependencies
}
```

Discovered relationships:
- Flow → Connector instances (via connectorInstances link)
- Flow → Variables (via variables link)
- Flow → Subflows (via subflows link)
- Application → Flow policies (via flowPolicies link)
- Flow policy → Flows (via flows link)

### Resource Filter Implementation

Create `internal/filter/filter.go`:

```go
package filter

type ResourceFilter struct {
    includeFlows       []string
    excludeFlows       []string
    includeApps        []string
    excludeApps        []string
    // ... other resource types
    withDependencies   bool
}

func (f *ResourceFilter) ShouldInclude(resourceType, resourceID, resourceName string) bool {
    // Check if resource matches include/exclude rules
    // Support ID matching (exact UUID match)
    // Support name matching (case-insensitive substring)
}

func (f *ResourceFilter) ApplyInclusions(resources interface{}) interface{} {
    // Filter resources to only included ones
}

func (f *ResourceFilter) ApplyExclusions(resources interface{}) interface{} {
    // Remove excluded resources
}
```

### Dependency Discovery Workflow

When `--skip-dependencies` is false (default):

1. User specifies: `--include-flows "Registration Flow"`
2. Export fetches "Registration Flow" resource
3. Parse HAL `_links` to find dependencies:
   - Connector instances linked via `connectorInstances` href
   - Variables linked via `variables` href
   - Subflows linked via `subflows` href
4. Fetch those dependency resources
5. Parse their HAL links for transitive dependencies
6. Continue recursively until all dependencies discovered
7. Export complete set of resources

Create `internal/exporter/dependency_discoverer.go`:

```go
func DiscoverDependencies(client *api.Client, initialResources []Resource) ([]Resource, error) {
    // Start with explicitly included resources
    // For each resource, parse HAL links
    // Fetch linked resources
    // Recursively discover dependencies
    // Build complete dependency tree
    // Return expanded resource set
}
```

### Handling Missing Dependencies

Three types of missing resources:

**1. Missing by exclusion** (user explicitly excluded):
```hcl
connection_id = ""  # TODO: Reference to "PingOne Connector" (ID: abc123) was excluded from export
```

**2. Missing by selection** (not in include filter):
```hcl
variable_id = ""  # TODO: Reference to "apiKey" (ID: xyz789) was not included in export filters
```

**3. Actually missing** (doesn't exist in environment):
```hcl
subflow_id = ""  # TODO: Reference to flow ID def456 not found in environment
```

### Update Exporter for Selective Export

Modify `Exporter.Export()`:

Two-phase export:
- **Phase 1**: Fetch specified resources (matching include/exclude filters)
- **Phase 2**: If `--skip-dependencies=false`, discover and fetch dependencies via HAL links

Track metadata:
- Resources explicitly selected vs. discovered
- Missing dependencies and reasons

### Testing Selective Export

Create `internal/filter/filter_test.go`:
- Test inclusion logic with IDs
- Test inclusion logic with names
- Test exclusion logic
- Test include + exclude combinations

Create `internal/api/hal_test.go`:
- Test HAL link parsing
- Test handling missing links
- Test invalid link formats

Create `internal/exporter/dependency_test.go`:
- Test single-level dependency discovery
- Test recursive dependency discovery
- Test `--skip-dependencies` flag
- Test circular dependency handling

**Example Test Scenarios**:

Scenario 1: Export single flow with dependencies
- Include: One flow by name
- Expect: Flow + connectors (via HAL) + variables (via HAL) + subflows (via HAL)

Scenario 2: Export application without dependencies
- Include: One application by ID
- Flag: `--skip-dependencies`
- Expect: Only application resource

Scenario 3: Export everything except test resources
- Exclude: Resources with "test" in name
- Expect: All resources except those matching "test"

---

## Success Criteria

**Phase 3.1-3.5** (Basic Export):
- ✅ Can authenticate using PingCLI credentials
- ✅ Can export all flows from an environment
- ✅ Can export all connector instances
- ✅ Can export all variables
- ✅ Can export all applications and flow policies
- ✅ Generated HCL includes all resource types in correct order
- ✅ Sensitive values are masked with TODO comments
- ✅ Resource names are sanitized and unique
- ✅ Export errors are handled gracefully

**Phase 3.6** (Selective Export - Future):
- ✅ Can filter resources by include/exclude flags
- ✅ Can discover dependencies via HAL link parsing
- ✅ Can export with or without dependencies
- ✅ Missing dependencies have clear TODO comments with reasons
- ✅ HAL link parsing handles all relationship types
- ✅ Recursive dependency discovery works correctly

---

## Phase 3.7: Acceptance Tests

**Goal**: End-to-end tests against real PingOne API to validate complete workflow.

### Test Structure

Create `tests/acceptance/` directory:
```
tests/acceptance/
  ├── README.md                    # Setup instructions, credentials
  ├── acceptance_test.go           # Main acceptance tests
  ├── fixtures/                    # Test environment setup
  └── .gitignore                   # Ignore credentials
```

### Build Tag for Isolation

```go
//go:build acceptance

package acceptance

import (
    "os"
    "testing"
)

func TestRealAPIFlowExport(t *testing.T) {
    // Only runs when: go test -tags=acceptance
}
```

### Required Test Cases

#### Test 1: Export Single Flow from Real Environment

```go
func TestExportSingleFlowFromAPI(t *testing.T) {
    // Skip if credentials not available
    clientID := os.Getenv("PINGONE_CLIENT_ID")
    if clientID == "" {
        t.Skip("Skipping acceptance test: PINGONE_CLIENT_ID not set")
    }
    
    // 1. Authenticate with PingOne
    client := createTestClient(t)
    
    // 2. Fetch a known flow from test environment
    flowID := os.Getenv("TEST_FLOW_ID")
    flow, err := client.GetFlow(context.Background(), flowID)
    require.NoError(t, err)
    
    // 3. Convert to HCL
    hcl, err := converter.ConvertFlow(flow)
    require.NoError(t, err)
    
    // 4. Validate HCL syntax
    assert.Contains(t, hcl, "resource \"pingone_davinci_flow\"")
    assert.Contains(t, hcl, "environment_id")
    
    // 5. Optional: Terraform validate
    validateTerraformHCL(t, hcl)
}
```

#### Test 2: Export All Resources from Environment

```go
func TestExportAllResourcesFromAPI(t *testing.T) {
    if os.Getenv("PINGONE_CLIENT_ID") == "" {
        t.Skip("Skipping acceptance test: credentials not set")
    }
    
    // 1. Create exporter with real client
    exporter := createTestExporter(t)
    
    // 2. Export entire environment
    hcl, err := exporter.ExportAll(context.Background())
    require.NoError(t, err)
    
    // 3. Verify all resource types present
    assert.Contains(t, hcl, "pingone_davinci_flow")
    assert.Contains(t, hcl, "pingone_davinci_connector_instance")
    assert.Contains(t, hcl, "pingone_davinci_variable")
    assert.Contains(t, hcl, "pingone_davinci_application")
    assert.Contains(t, hcl, "pingone_davinci_application_flow_policy")
    
    // 4. Verify resource ordering
    verifyResourceOrdering(t, hcl)
    
    // 5. Count resources (should match API count)
    apiCounts := exporter.GetResourceCounts()
    hclCounts := countResourcesInHCL(hcl)
    assert.Equal(t, apiCounts, hclCounts)
}
```

#### Test 3: Export with Skip Dependencies

```go
func TestExportWithSkipDependencies(t *testing.T) {
    if os.Getenv("PINGONE_CLIENT_ID") == "" {
        t.Skip("Skipping acceptance test: credentials not set")
    }
    
    exporter := createTestExporter(t)
    
    // Export with skip-dependencies flag
    hcl, err := exporter.ExportWithOptions(context.Background(), ExportOptions{
        SkipDependencies: true,
    })
    require.NoError(t, err)
    
    // Verify hardcoded IDs instead of references
    assert.NotContains(t, hcl, "pingone_davinci_connector_instance.")
    assert.Contains(t, hcl, "connection_id   = \"")  // Hardcoded ID
}
```

#### Test 4: Terraform Apply (Optional - Separate Environment)

```go
func TestTerraformApplyExportedHCL(t *testing.T) {
    if os.Getenv("ENABLE_TERRAFORM_APPLY_TEST") != "true" {
        t.Skip("Skipping expensive Terraform apply test")
    }
    
    // 1. Export from source environment
    sourceExporter := createTestExporter(t)
    hcl, err := sourceExporter.ExportAll(context.Background())
    require.NoError(t, err)
    
    // 2. Write to temporary directory
    tmpDir := t.TempDir()
    writeTestTerraformConfig(t, tmpDir, hcl)
    
    // 3. Point to different target environment
    updateProviderConfig(t, tmpDir, getTargetEnvironmentID())
    
    // 4. Run terraform apply
    err = runTerraformApply(t, tmpDir)
    require.NoError(t, err)
    
    // 5. Verify resources created in target environment
    verifyResourcesCreated(t, getTargetEnvironmentID())
    
    // 6. Cleanup: Run terraform destroy
    t.Cleanup(func() {
        runTerraformDestroy(t, tmpDir)
    })
}
```

### Test Environment Requirements

Document in `tests/acceptance/README.md`:

```markdown
# Acceptance Tests

## Prerequisites

1. **PingOne Test Environment**: Dedicated test environment with sample resources
2. **Service Account**: OAuth client with sufficient permissions
3. **Environment Variables**:
   ```bash
   export PINGONE_CLIENT_ID="your-client-id"
   export PINGONE_CLIENT_SECRET="your-client-secret"
   export PINGONE_ENVIRONMENT_ID="test-env-id"
   export PINGONE_REGION="NA"
   
   # Optional: For terraform apply tests
   export ENABLE_TERRAFORM_APPLY_TEST="true"
   export TEST_FLOW_ID="known-flow-id"
   ```

## Running Tests

```bash
# Run all acceptance tests
go test -tags=acceptance ./tests/acceptance -v

# Run specific test
go test -tags=acceptance ./tests/acceptance -run TestExportSingleFlow -v

# Skip if credentials not set (automatic)
go test -tags=acceptance ./tests/acceptance -v
# Output: SKIP: credentials not set
```

## Test Data Setup

The test environment should contain:
- At least 1 flow with nodes
- At least 1 connector instance
- At least 1 variable
- At least 1 application with flow policy

## CI/CD Integration

Add to CI pipeline (scheduled, not on every commit):
```yaml
# .github/workflows/acceptance-tests.yml
name: Acceptance Tests
on:
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2 AM
  workflow_dispatch:     # Manual trigger

jobs:
  acceptance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Run acceptance tests
        env:
          PINGONE_CLIENT_ID: ${{ secrets.ACCEPTANCE_CLIENT_ID }}
          PINGONE_CLIENT_SECRET: ${{ secrets.ACCEPTANCE_CLIENT_SECRET }}
          PINGONE_ENVIRONMENT_ID: ${{ secrets.ACCEPTANCE_ENV_ID }}
          PINGONE_REGION: NA
        run: go test -tags=acceptance ./tests/acceptance -v
```
```

### When to Run Acceptance Tests

✅ **Run acceptance tests**:
- During development of API integration (Part 3)
- Before releasing new versions
- Nightly in CI/CD against test environment
- When debugging API-related issues

❌ **Don't run acceptance tests**:
- On every commit (too slow)
- In local unit test runs (separate command)
- Without proper test environment

### Helper Functions

Create `tests/acceptance/helpers.go`:

```go
package acceptance

import (
    "testing"
    "context"
    "os"
)

func createTestClient(t *testing.T) *api.Client {
    clientID := requireEnv(t, "PINGONE_CLIENT_ID")
    clientSecret := requireEnv(t, "PINGONE_CLIENT_SECRET")
    envID := requireEnv(t, "PINGONE_ENVIRONMENT_ID")
    region := requireEnv(t, "PINGONE_REGION")
    
    client, err := api.NewClient(context.Background(), envID, region, clientID, clientSecret)
    require.NoError(t, err)
    return client
}

func requireEnv(t *testing.T, key string) string {
    value := os.Getenv(key)
    if value == "" {
        t.Fatalf("Required environment variable %s not set", key)
    }
    return value
}

func validateTerraformHCL(t *testing.T, hcl string) {
    // Write to temp dir, run terraform validate
}

func countResourcesInHCL(hcl string) map[string]int {
    // Parse HCL and count each resource type
}
```

---

**Next Step**: After Part 3 complete (including acceptance tests), proceed to Part 4 (Dependency Resolution and Terraform References).

