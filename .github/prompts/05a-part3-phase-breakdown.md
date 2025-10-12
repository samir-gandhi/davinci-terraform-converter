---
mode: agent
---

# Part 3: Phase Breakdown and Review Strategy

**Goal**: Break Part 3 into small, testable, reviewable pieces.

---

## Implementation Strategy

Each phase creates ONE new file with tests. Review after each phase before continuing.

---

## Phase 3.0: Authentication Setup ✅ COMPLETE

**Files Created**:
- `internal/api/client.go` (basic structure with validation)
- `internal/api/client_test.go` (13 test cases)

**Status**: Basic client validation complete. Need to add SDK integration.

**Tests Passing**: 13/13

---

## Phase 3.1a: Add SDK Authentication ✅ COMPLETE

**Goal**: Integrate pingone-go-client SDK for OAuth authentication.

**Files Updated**:
- `internal/api/client.go` - Added SDK integration with OAuth client credentials
- `go.mod` - Added pingone-go-client dependency with local replace directive

**Implementation**:
- Configured OAuth client credentials grant type
- Integrated with pingone-go-client SDK Configuration
- Handles regional endpoints (NA, EU, AP, CA)
- Automatic token acquisition via SDK

**Tests Passing**: 13/13 (all existing tests still pass with SDK)

---

## Phase 3.1b: Flow API Client ✅ COMPLETE

**Goal**: Create API client for retrieving flow data.

**Files Created**:
- `internal/api/flows.go`
- `internal/api/flows_test.go`

**Functions Implemented**:
- `ListFlows(ctx context.Context) ([]FlowSummary, error)` - Retrieves all flows from environment
- `GetFlow(ctx context.Context, flowID string) (*FlowDetail, error)` - Retrieves detailed flow data including graph
- `stringValue(*string) string` - Helper for safe pointer dereferencing

**Implementation Details**:
- Uses SDK's `DaVinciFlowsApi.GetFlows()` for listing
- Uses SDK's `DaVinciFlowsApi.GetFlowById()` for details
- Converts SDK response types to internal types (FlowSummary, FlowDetail)
- Handles embedded response structure
- Converts GraphData to map[string]interface{} for converter compatibility

**Tests Implemented**:
- Client structure validation for flow operations
- UUID validation for environment IDs
- Helper function tests (stringValue)
- Note: Full integration tests with real API require acceptance test framework (Phase 3.8)

**Tests Passing**: 16/16 (13 client + 3 flows)

---

## Phase 3.1c: Flow Export Integration

**Goal**: Connect flow API client to existing flow converter.

**Files to Create**:
- `internal/exporter/flow_exporter.go`
- `internal/exporter/flow_exporter_test.go`

**Functions to Implement**:
```go
// ExportFlows retrieves flows from API and converts to HCL
func ExportFlows(ctx context.Context, client *api.Client, skipDeps bool) (string, error)
```

**Test Approach**: Use mock API client, verify HCL output.

**Success Criteria**:
- Calls flow API correctly
- Passes flow data to converter
- Returns combined HCL for all flows
- Handles empty environments
- All tests pass

**Review Point**: Stop after flow export tested.

---

## Phase 3.2a: Connector Instance API Client

**Goal**: Create API client for connector instances.

**Files to Create**:
- `internal/api/connector_instances.go`
- `internal/api/connector_instances_test.go`

**Functions to Implement**:
```go
// ListConnectorInstances retrieves all connector instances
func (c *Client) ListConnectorInstances(ctx context.Context) ([]ConnectorInstance, error)

// GetConnectorInstance retrieves detailed connector data
func (c *Client) GetConnectorInstance(ctx context.Context, instanceID string) (*ConnectorInstanceDetail, error)
```

**Test Approach**: Use mock HTTP server with connector API responses.

**Success Criteria**:
- Can retrieve connector list
- Can retrieve individual connector details
- Handles pagination
- All tests pass

**Review Point**: Stop after connector API client tested.

---

## Phase 3.2b: Connector Instance Export Integration

**Goal**: Connect connector API to existing connector converter.

**Files to Create**:
- `internal/exporter/connector_exporter.go`
- `internal/exporter/connector_exporter_test.go`

**Functions to Implement**:
```go
// ExportConnectorInstances retrieves connectors from API and converts to HCL
func ExportConnectorInstances(ctx context.Context, client *api.Client, skipDeps bool) (string, error)
```

**Success Criteria**:
- Calls connector API correctly
- Passes data to converter
- Returns HCL with masked secrets
- All tests pass

**Review Point**: Stop after connector export tested.

---

## Phase 3.3a: Variable API Client

**Goal**: Create API client for variables.

**Files to Create**:
- `internal/api/variables.go`
- `internal/api/variables_test.go`

**Functions to Implement**:
```go
// ListVariables retrieves all variables (company and flow contexts)
func (c *Client) ListVariables(ctx context.Context) ([]Variable, error)
```

**Success Criteria**:
- Can retrieve all variable contexts
- Handles different variable types
- All tests pass

**Review Point**: Stop after variable API client tested.

---

## Phase 3.3b: Variable Export Integration

**Goal**: Connect variable API to existing variable converter.

**Files to Create**:
- `internal/exporter/variable_exporter.go`
- `internal/exporter/variable_exporter_test.go`

**Success Criteria**:
- Calls variable API correctly
- Passes data to converter
- Returns HCL with masked secrets
- All tests pass

**Review Point**: Stop after variable export tested.

---

## Phase 3.4a: Application API Client

**Goal**: Create API client for applications.

**Files to Create**:
- `internal/api/applications.go`
- `internal/api/applications_test.go`

**Functions to Implement**:
```go
// ListApplications retrieves all DaVinci applications
func (c *Client) ListApplications(ctx context.Context) ([]Application, error)

// GetApplication retrieves detailed application data
func (c *Client) GetApplication(ctx context.Context, appID string) (*ApplicationDetail, error)
```

**Success Criteria**:
- Can retrieve application list
- Can retrieve application details
- All tests pass

**Review Point**: Stop after application API client tested.

---

## Phase 3.4b: Application Export Integration

**Goal**: Connect application API to existing application converter.

**Files to Create**:
- `internal/exporter/application_exporter.go`
- `internal/exporter/application_exporter_test.go`

**Success Criteria**:
- Calls application API correctly
- Passes data to converter
- Returns HCL with OAuth/API key blocks
- All tests pass

**Review Point**: Stop after application export tested.

---

## Phase 3.5a: Flow Policy API Client

**Goal**: Create API client for flow policies.

**Files to Create**:
- `internal/api/flow_policies.go`
- `internal/api/flow_policies_test.go`

**Functions to Implement**:
```go
// ListFlowPolicies retrieves all flow policies
func (c *Client) ListFlowPolicies(ctx context.Context) ([]FlowPolicy, error)

// GetFlowPolicy retrieves detailed flow policy data
func (c *Client) GetFlowPolicy(ctx context.Context, policyID string) (*FlowPolicyDetail, error)
```

**Success Criteria**:
- Can retrieve policy list
- Can retrieve policy details
- All tests pass

**Review Point**: Stop after flow policy API client tested.

---

## Phase 3.5b: Flow Policy Export Integration

**Goal**: Connect flow policy API to existing flow policy converter.

**Files to Create**:
- `internal/exporter/flow_policy_exporter.go`
- `internal/exporter/flow_policy_exporter_test.go`

**Success Criteria**:
- Calls flow policy API correctly
- Passes data to converter
- Returns HCL with trigger/distribution blocks
- All tests pass

**Review Point**: Stop after flow policy export tested.

---

## Phase 3.6: Orchestrator

**Goal**: Coordinate all exporters to create complete environment export.

**Files to Create**:
- `internal/exporter/orchestrator.go`
- `internal/exporter/orchestrator_test.go`

**Functions to Implement**:
```go
// ExportEnvironment exports all resources in dependency order
func ExportEnvironment(ctx context.Context, client *api.Client, skipDeps bool) (string, error)
```

**Resource Export Order**:
1. Variables (no dependencies)
2. Connector Instances (no dependencies)
3. Flows (depends on connectors)
4. Applications (depends on flows)
5. Flow Policies (depends on applications and flows)

**Success Criteria**:
- Calls all exporters in correct order
- Combines HCL from all resources
- Includes provider and variable configuration
- All tests pass

**Review Point**: Stop after orchestrator tested.

---

## Phase 3.7: CLI Integration

**Goal**: Add export capability to CLI command.

**Files to Update**:
- `cmd/convert.go`

**New Flags to Add**:
```
--export                     Enable API export mode
--environment-id <uuid>      PingOne environment ID
--region <code>              PingOne region (NA/EU/AP/CA)
--client-id <id>             OAuth client ID
--client-secret <secret>     OAuth client secret
```

**Logic**:
- If `--export` provided, use API mode
- If `--flow-json` provided, use file mode (existing)
- Validate required credentials for export mode
- Call orchestrator to perform export
- Write output to file

**Success Criteria**:
- Can run export from command line
- Validates required flags
- Creates complete HCL output
- Manual testing successful

**Review Point**: Stop after CLI integration tested.

---

## Phase 3.8: Acceptance Tests (Optional)

**Goal**: Create real API integration tests (optional, requires credentials).

**Files to Create**:
- `tests/acceptance/export_test.go` with `//go:build acceptance` tag

**Test Approach**:
- Requires real PingOne environment credentials
- Uses environment variables for configuration
- Tests against real API endpoints
- Validates generated HCL can be applied

**Success Criteria**:
- Tests pass when run with valid credentials
- Tests skipped when credentials not available
- Documentation added for running acceptance tests

**Review Point**: Final review after acceptance tests.

---

## Summary

**Total Phases**: 15 reviewable checkpoints

**Total New Files**: ~28 files (14 implementation + 14 test files)

**Approach**: 
1. Implement one phase
2. Run tests
3. Review code
4. Get confirmation to continue
5. Move to next phase

**Current Position**: Phase 3.0 complete, ready for Phase 3.1a (SDK Authentication)
