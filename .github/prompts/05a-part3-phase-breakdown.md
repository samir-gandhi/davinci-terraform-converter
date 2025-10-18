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

## Phase 3.1c: Flow Export Integration ✅ COMPLETE

**Goal**: Connect flow API client to existing flow converter.

**Files Created**:
- `internal/exporter/flow_exporter.go`
- `internal/exporter/flow_exporter_test.go` (unit tests)
- `tests/acceptance/acceptance_test.go` (acceptance test framework)
- `tests/acceptance/flow_export_test.go` (5 acceptance tests)
- `tests/acceptance/helpers.go` (test utilities)

**Functions Implemented**:
- `ExportFlows(ctx context.Context, client *api.Client, skipDeps bool) (string, error)` - Exports all flows to HCL
- `ExportFlowsJSON(ctx context.Context, client *api.Client) (string, error)` - Exports flows in JSON format
- `convertFlowDetailToMap(*api.FlowDetail) (map[string]interface{}, error)` - Converts API response to converter format

**Implementation Details**:
- Retrieves all flows via `ListFlows()` API call
- Fetches detailed flow data for each flow via `GetFlow()` API call
- Converts each flow to HCL using existing converter
- Combines all flow resources with blank lines between them
- Supports skip-dependencies flag
- JSON export available for debugging

**SDK Workaround Implemented**:
- **Issue**: SDK requires Position field in flow graph nodes/edges, but API returns flows without it
- **Solution**: `GetFlow()` uses raw HTTP requests to bypass SDK's strict validation
- **Files Modified**: `internal/api/client.go` (added serviceCfg), `internal/api/flows.go` (raw HTTP implementation)
- **Documentation**: Created `SDK_POSITION_FIELD_ISSUE.md`, `WORKAROUND_RAW_HTTP.md`, `WORKAROUND_SUMMARY.md`
- **Reversion Plan**: Complete instructions in `WORKAROUND_RAW_HTTP.md` for when SDK is fixed

**Acceptance Tests Implemented** (11 tests total):
1. TestAPIClientAuthentication ✅ - OAuth2 setup validation
2. TestListFlowsFromAPI ✅ - Lists 7 flows from environment
3. TestGetSingleFlowFromAPI ✅ - Retrieves individual flow details
4. TestGetFlowWithSpecificID ✅ - Fetches by ID or first available
5. TestAPIErrorHandling ✅ - Validates 404 error responses
6. TestMultipleFlowRetrieval ✅ - Sequential flow fetching
7. TestExportFlowsFromAPI ✅ - Full HCL export (442KB)
8. TestExportFlowsWithSkipDependencies ✅ - Skip-deps flag (435KB)
9. TestExportFlowsJSON ✅ - JSON format export (509KB)
10. TestExportFlowsValidateHCLStructure ✅ - HCL syntax validation
11. TestExportSingleFlowComparison ✅ - Consistency check

**Test Environment**:
- Auth Environment: `01134525-6c1d-4475-839b-e92d12876a46`
- Target Environment: `62f10a04-6c54-40c2-a97d-80a98522ff9a` (7 flows)
- Region: NA
- Environment Variables: PINGCLI_PINGONE_WORKER_CLIENT_ID, PINGCLI_PINGONE_WORKER_CLIENT_SECRET, PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID, PINGCLI_PINGONE_EXPORT_ENVIRONMENT_ID, PINGONE_REGION

**Tests Passing**: 16/16 unit tests + 11/11 acceptance tests = 27/27 total

**Success Criteria**: ✅ ALL COMPLETE
- ✅ Calls flow API correctly
- ✅ Passes flow data to converter
- ✅ Returns combined HCL for all flows
- ✅ Handles environments with flows (7 found)
- ✅ All unit tests pass
- ✅ All acceptance tests pass against real API
- ✅ SDK workaround documented and working
- ✅ Performance validated (~2.2s for 7 flows)

**Review Point**: Phase 3.1c complete. Ready for Phase 3.2a (Connector Instance API Client).

---

## Phase 3.2a: Connector Instance API Client ✅ COMPLETE

**Goal**: Create API client for connector instances using SDK.

**Files Created**:
- `internal/api/connector_instances.go`
- `internal/api/connector_instances_test.go`

**Functions Implemented**:
- `ListConnectorInstances(ctx context.Context) ([]ConnectorInstanceSummary, error)` - Retrieves all connector instances
- `GetConnectorInstance(ctx context.Context, instanceID string) (*ConnectorInstanceDetail, error)` - Retrieves detailed connector data with properties

**Implementation Details**:
- Uses SDK's `DaVinciConnectorsApi.GetConnectorInstances()` for listing
- Uses SDK's `DaVinciConnectorsApi.GetConnectorInstanceById()` for details
- Extracts connector instances from embedded response structure
- Retrieves connector ID from relationship field
- Handles properties map for connector configuration
- No SDK validation issues encountered (unlike flows API)

**Acceptance Tests Implemented**:
- `TestListConnectorInstancesFromAPI` - List all connector instances from environment
- `TestGetSingleConnectorInstanceFromAPI` - Get single instance with properties validation
- `TestGetInvalidConnectorInstanceFromAPI` - Error handling for non-existent instances
- `TestMultipleConnectorInstanceRetrieval` - Sequential retrieval of multiple instances

**Test Results**:
- Unit tests: 19/19 passing (internal/api)
- Acceptance tests: 15/15 passing (11 flow + 4 connector instance)
- Total unit tests: 66/66 passing across all packages

**Test Environment Data**:
- 20 connector instances found in environment
- Sample instances include: Variables, PingOne Protect, various connector types
- Properties may or may not be present depending on connector type

**Key Decisions**:
- Used SDK methods directly instead of raw HTTP (no validation issues like flows API)
- Cleaner implementation compared to raw HTTP approach
- More maintainable with SDK error handling

**Review Point**: Phase 3.2a complete. Ready for Phase 3.2b (Connector Instance Export Integration).

---

## Phase 3.2b: Connector Instance Export Integration ✅ COMPLETE

**Goal**: Connect connector API to existing connector converter.

**Files Created**:
- `internal/exporter/connector_exporter.go`
- `internal/exporter/connector_exporter_test.go`
- `tests/acceptance/connector_export_test.go`

**Functions Implemented**:
- `ExportConnectorInstances(ctx context.Context, client *api.Client, skipDeps bool) (string, error)` - Exports all connector instances to HCL
- `convertInstanceDetailToJSON(*api.ConnectorInstanceDetail) ([]byte, error)` - Converts API response to converter format

**Implementation Details**:
- Retrieves all connector instances via `ListConnectorInstances()` API call
- Fetches detailed instance data for each instance via `GetConnectorInstance()` API call
- Converts each instance to JSON format expected by converter
- Converts to HCL using existing `converter.ConvertConnectorInstanceWithOptions()`
- Combines all instance resources with blank lines between them
- Supports skip-dependencies flag (uses actual environment ID vs var.environment_id)

**Acceptance Tests Implemented**:
- `TestExportConnectorInstancesFromAPI` - Export all connector instances to HCL
- `TestExportConnectorInstancesWithSkipDependenciesFromAPI` - Test skip-dependencies flag
- `TestExportConnectorInstancesValidateHCLStructure` - Validate HCL structure for each instance
- `TestExportSingleConnectorInstanceComparison` - Compare API and exported data
- `TestExportConnectorInstancesPropertiesHandling` - Test property masking and structure

**Test Results**:
- Unit tests: 23/23 passing (internal/exporter - 3 connector + 3 flow tests)
- Acceptance tests: 20/20 passing (10 flow + 5 connector API + 5 connector export)
- Total unit tests: 70/70 passing across all packages

**Test Environment Data**:
- 20 connector instances exported
- Generated HCL: 4.6KB output
- Sample instances: Variables, PingOne Protect, samAnnotationConnector, etc.
- Properties handled correctly with secret masking

**Key Discoveries**:
- Connector instance IDs can be non-UUID format (e.g., "defaultUserPool")
- Removed UUID validation from `GetConnectorInstance()` - only validates non-empty
- Updated validation tests to reflect non-UUID instance IDs
- Secrets properly masked: `clientSecret` → `"TODO: Replace with actual client secret"`

**Success Criteria Met**:
- ✅ Calls connector API correctly (ListConnectorInstances + GetConnectorInstance)
- ✅ Passes data to converter in correct JSON format
- ✅ Returns HCL with masked secrets
- ✅ All tests pass (70/70 unit tests, 20/20 acceptance tests)

**Review Point**: Phase 3.2b complete. Ready for Phase 3.3a (Variable API Client).

---

## Phase 3.3a: Variable API Client ✅ COMPLETE

**Goal**: Create API client for variables using SDK.

**Files Created**:
- `internal/api/variables.go`
- `internal/api/variables_test.go`
- `tests/acceptance/variable_api_test.go`

**Functions Implemented**:
- `ListVariables(ctx context.Context, environmentID string) ([]pingone.DaVinciVariableResponse, error)` - Retrieves all variables from environment
- `GetVariable(ctx context.Context, environmentID, variableID string) (*pingone.DaVinciVariableResponse, error)` - Retrieves specific variable by ID

**Implementation Details**:
- Uses SDK's `DaVinciVariablesApi.GetVariables()` for listing with pagination
- Uses SDK's `DaVinciVariablesApi.GetVariableById()` for details
- Handles embedded response structure with GetVariablesOk()
- Iterates through paginated results
- Validates UUID format for environment and variable IDs
- No SDK validation issues encountered

**Acceptance Tests Implemented**:
- `TestVariableAPI` - List all variables from environment with context validation
- `TestGetVariableById` - Get existing variable and test error handling for nonexistent
- `TestListVariablesEmpty` - Verify empty result handling for environment without variables

**Test Results**:
- Unit tests: 25/25 passing (internal/api - 19 previous + 6 variable tests)
- Acceptance tests: 23/23 passing (20 previous + 3 variable tests)
- Total unit tests: 76/76 passing across all packages

**Test Environment Data**:
- 4 variables found in target environment:
  - companyLogo (flowInstance context)
  - companyName (flowInstance context)
  - SampleUserContextVariable (user context)
  - SampleVariable (company context)
- 0 variables in worker environment (empty list validated)

**Key Discoveries**:
- SDK supports variables API directly with GetVariables() and GetVariableById()
- Pagination handled via iterator pattern
- Variable contexts include: company, flowInstance, user

**Success Criteria Met**:
- ✅ Can retrieve all variable contexts
- ✅ Handles different variable types (string, various contexts)
- ✅ All tests pass (76/76 unit tests, 23/23 acceptance tests)

**Review Point**: Phase 3.3a complete. Ready for Phase 3.3b (Variable Export Integration).

---

## Phase 3.3b: Variable Export Integration ✅ COMPLETE

**Goal**: Connect variable API to existing variable converter.

**Files Created**:
- `internal/exporter/variable_exporter.go`
- `internal/exporter/variable_exporter_test.go`
- `tests/acceptance/variable_export_test.go`

**Functions Implemented**:
- `ExportVariables(ctx context.Context, client *api.Client, skipDeps bool) (string, error)` - Exports all variables to HCL
- `convertVariableToJSON(variable interface{}) ([]byte, error)` - Converts SDK response to JSON format

**Implementation Details**:
- Retrieves all variables via `ListVariables()` API call
- Converts each SDK variable response to JSON format expected by converter
- Uses existing `converter.ConvertVariableWithOptions()` to generate HCL
- Combines all variable resources with blank lines between them
- Supports skip-dependencies flag

**Acceptance Tests Implemented**:
- `TestExportVariablesFromAPI` - Export all variables to HCL with preview
- `TestExportVariablesWithSkipDependencies` - Test skip-dependencies flag
- `TestExportVariablesValidateHCLStructure` - Validate HCL blocks, contexts, and data types
- `TestExportVariablesComparison` - Compare API count to HCL resource count
- `TestExportVariablesValueHandling` - Validate value, min, max, mutable fields

**Test Results**:
- Unit tests: 27/27 passing (internal/exporter - 3 connector + 3 flow + 2 variable tests)
- Acceptance tests: 28/28 passing (20 previous + 3 variable API + 5 variable export)
- Total unit tests: 78/78 passing across all packages

**Test Environment Data**:
- 16 variables exported from target environment
- Generated HCL: 4.9KB output
- Variable contexts tested: company, flowInstance, user (all 3 found)
- Data types tested: string, number, boolean, object, secret (all 5 found)
- Sample variables: companyBool, companyLogo, companyName, companyNumber, companyObject, companySecret, companyString, flowBoolean, flowNumber, flowObject, flowString, SampleUserContextVariable, SampleVariable, userbool, userNumber, userObject

**Key Validations**:
- All 16 variables from API appear in HCL output
- Resource type: `pingone_davinci_variable`
- Required fields validated: environment_id, name, context, data_type, mutable
- Optional fields present: value, min, max, display_name
- Secret data type properly handled with TODO comment

**Success Criteria Met**:
- ✅ Calls variable API correctly (ListVariables)
- ✅ Passes data to converter in correct JSON format
- ✅ Returns HCL with proper structure
- ✅ All tests pass (78/78 unit tests, 28/28 acceptance tests)
- ✅ Handles variety of contexts and data types

**Review Point**: Phase 3.3b complete. Ready for Phase 3.4a (Application API Client).

---

## Phase 3.4a: Application API Client ✅ COMPLETE

**Goal**: Create API client for DaVinci applications using SDK.

**Files Created**:
- `internal/api/applications.go`
- `internal/api/applications_test.go`
- `tests/acceptance/application_api_test.go`

**Functions Implemented**:
- `ListApplications(ctx context.Context, environmentID string) ([]pingone.DaVinciApplicationResponse, error)` - Retrieves all DaVinci applications
- `GetApplication(ctx context.Context, environmentID, applicationID string) (*pingone.DaVinciApplicationResponse, error)` - Retrieves specific application by ID

**Implementation Details**:
- Uses SDK's `DaVinciApplicationsApi.GetDavinciApplications()` for listing
- Uses SDK's `DaVinciApplicationsApi.GetDavinciApplicationById()` for details
- Handles embedded response structure with GetDavinciApplicationsOk()
- Validates UUID format for environment IDs
- Application IDs are strings (not UUIDs)
- No SDK validation issues encountered

**Acceptance Tests Implemented**:
- `TestApplicationAPI` - List all applications from environment
- `TestGetApplicationById` - Get existing application and test error handling for nonexistent
- `TestListApplicationsEmpty` - Verify empty result handling for environment without applications

**Test Results**:
- Unit tests: 30/30 passing (internal/api - 25 previous + 5 application tests)
- Acceptance tests: 31/31 passing (28 previous + 3 application tests)
- Total unit tests: 83/83 passing across all packages

**Test Environment Data**:
- 10 applications found in target environment
- Sample applications: applicationFFlowPolicyMultiWeight, applicationEFlowPolicyWVersion, applicationDP1FlowPolicyOnly, applicationCConfigured, applicationBDefault, applicationAMinimal, DaVinci API Protect Sample Application (multiple instances)
- 0 applications in worker environment (empty list validated)
- Applications can have API key configuration
- Applications can have OAuth configuration
- Both configurations can exist on same application

**Key Discoveries**:
- SDK supports applications API directly with GetDavinciApplications() and GetDavinciApplicationById()
- Application IDs are strings (not UUID format like environment IDs)
- Applications can have multiple authentication configurations (API key and/or OAuth)

**Success Criteria Met**:
- ✅ Can retrieve application list
- ✅ Can retrieve application details with API key and OAuth configurations
- ✅ All tests pass (83/83 unit tests, 31/31 acceptance tests)

**Review Point**: Phase 3.4a complete. Ready for Phase 3.4b (Application Export Integration).

---

## Phase 3.4b: Application Export Integration ✅ COMPLETE

**Goal**: Connect application API to existing application converter.

**Files Created**:
- `internal/exporter/application_exporter.go`
- `internal/exporter/application_exporter_test.go`
- `tests/acceptance/application_export_test.go`

**Functions Implemented**:
- `ExportApplications(ctx context.Context, client *api.Client, skipDeps bool) (string, error)` - Exports all applications to HCL
- `convertApplicationToJSON(application interface{}) ([]byte, error)` - Converts SDK response to JSON format

**Implementation Details**:
- Retrieves all applications via `ListApplications()` API call
- Converts each SDK application response to JSON format expected by converter
- Uses existing `converter.ConvertApplicationWithOptions()` to generate HCL
- Combines all application resources with blank lines between them
- Supports skip-dependencies flag

**Acceptance Tests Implemented**:
- `TestExportApplicationsFromAPI` - Export all applications to HCL with preview
- `TestExportApplicationsWithSkipDependencies` - Test skip-dependencies flag
- `TestExportApplicationsValidateHCLStructure` - Validate HCL blocks, API key, OAuth configurations
- `TestExportApplicationsComparison` - Compare API count to HCL resource count
- `TestExportApplicationsAuthConfigHandling` - Validate authentication configuration handling

**Test Results**:
- Unit tests: 29/29 passing (internal/exporter - 3 connector + 3 flow + 2 variable + 2 application tests)
- Acceptance tests: 36/36 passing (31 previous + 5 application export)
- Total unit tests: 85/85 passing across all packages

**Test Environment Data**:
- 10 applications exported from target environment
- Generated HCL: 3.7KB output
- Sample applications: applicationFFlowPolicyMultiWeight, applicationEFlowPolicyWVersion, applicationDP1FlowPolicyOnly, applicationCConfigured, applicationBDefault, applicationAMinimal, DaVinci API Protect Sample Application (multiple instances)
- All applications found in HCL output
- Authentication methods validated: API key and OAuth configurations

**Key Validations**:
- All 10 applications from API appear in HCL output
- Resource type: `pingone_davinci_application`
- Required fields validated: environment_id, name
- Optional fields present: api_key (with enabled field), oauth (with grant_types, scopes)
- API key enabled field: `api_key = { enabled = true }`
- OAuth fields: grant_types, scopes, client_id, client_secret (when present)

**Success Criteria Met**:
- ✅ Calls application API correctly (ListApplications)
- ✅ Passes data to converter in correct JSON format
- ✅ Returns HCL with proper structure
- ✅ Handles OAuth and API key blocks
- ✅ All tests pass (85/85 unit tests, 36/36 acceptance tests)

**Review Point**: Phase 3.4b complete. Ready for Phase 3.4c (Terraform Validation).

---

## Phase 3.4c: Terraform Validation Testing ⚠️ IN PROGRESS

**Goal**: Validate that all Phase 3 exported resources produce valid Terraform HCL that passes terraform init and terraform validate.

**Files Created**:
- `tests/acceptance/terraform_validation_test.go` (5 comprehensive terraform validation tests)

**Test Functions Implemented**:
- `TestTerraformValidateVariablesFromAPI` ✅ PASSING - Validates variables HCL (16 variables, 5.5KB)
- `TestTerraformValidateConnectorInstancesFromAPI` ❌ FAILING - environment_id validation error
- `TestTerraformValidateApplicationsFromAPI` ❌ FAILING - Syntax errors
- `TestTerraformValidateFlowsFromAPI` ❌ FAILING - Single quote syntax errors
- `TestTerraformValidateAllResourcesFromAPI` ❌ FAILING - Combined validation

**Test Implementation Details**:
- Creates temporary directory for each test
- Exports resources from API to HCL
- Writes provider.tf with pingidentity/pingone provider configuration
- Writes variables.tf with environment_id and region variables
- Runs `terraform init` to download provider
- Runs `terraform validate` to check HCL correctness
- Provider: Uses development override at /Users/samirgandhi/go/bin

**Bugs Discovered and Fixed**:

**Variables Converter** (✅ FIXED):
- Wrong attribute names: Changed `boolean` → `bool`, `number` → `float32`, `object` → `json_object` to match provider schema
- Empty value blocks: Fixed `value = {}` being written when no value present
- Secret masking: API returns `"******"` for secrets, now detected and treated as empty value
- Mutable validation: Fixed to set `mutable = true` when no value is written (provider requirement)
- Files modified: `internal/converter/variable_converter.go`, `internal/converter/variable_converter_test.go`
- Result: All variables terraform validation PASSING

**Connector Instance Converter** (❌ NEEDS FIX):
- Issue: `environment_id` validation error - "Must be a valid UUID, got: " (empty or malformed)
- Likely cause: Not properly setting environment_id in HCL output
- Status: Needs investigation and fix

**Application Converter** (❌ NEEDS FIX):
- Issue: Syntax errors causing terraform init failure
- Error details: Unknown - need to examine actual HCL output
- Status: Needs investigation and fix

**Flow Converter** (❌ NEEDS FIX):
- Issue: Single quote syntax errors - "Invalid character - Single quotes are not valid. Use double quotes"
- Likely cause: Improper string escaping in HCL generation
- Status: Needs investigation and fix

**Test Results**:
- Terraform validation tests: 1/5 passing (Variables only)
- Variables: ✅ "Success! The configuration is valid, but there were some validation warnings"
- Other resources: ❌ Schema violations and syntax errors

**Success Criteria**:
- ✅ Variables: terraform init and validate passing
- ❌ Connector instances: environment_id validation error
- ❌ Applications: syntax errors
- ❌ Flows: quote escaping issues
- ❌ All resources: combined validation failing

**Review Point**: Phase 3.4c partially complete. Variables validated successfully. Other converters need fixes before continuing to Phase 3.5a.

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

**Phases Complete**: 10/15 (Phase 3.0, 3.1a, 3.1b, 3.1c, 3.2a, 3.2b, 3.3a, 3.3b, 3.4a, 3.4b)

**Phase In Progress**: Phase 3.4c (Terraform Validation Testing) - Variables passing, other resources need fixes

**Total New Files**: ~28 files planned (14 implementation + 14 test files)
- **Files Created So Far**: 24 files (10 implementation + 14 test files)

**Approach**: 
1. Implement one phase
2. Run tests
3. Review code
4. Get confirmation to continue
5. Move to next phase

**Current Position**: Phase 3.4c in progress - terraform validation testing. Variables converter fully validated and passing. Connector instance, application, and flow converters have schema violations that need fixing.

**Key Achievements**:
- ✅ OAuth2 authentication with PingOne SDK
- ✅ Dual-environment support (auth env vs target env)
- ✅ Flow listing and retrieval from API (8 flows)
- ✅ Flow export to HCL with 442KB output from 8 flows
- ✅ SDK workaround for Position and Version field validation issues (flows API)
- ✅ Connector instance listing and retrieval from API (20 instances found)
- ✅ Connector instance export to HCL with 4.6KB output from 20 instances
- ✅ SDK direct usage for connector instances (no validation issues beyond non-UUID IDs)
- ✅ Variable listing and retrieval from API (16 variables found)
- ✅ Variable export to HCL with 4.9KB output from 16 variables
- ✅ Variable variety tested: 3 contexts (company, flowInstance, user) and 5 data types (string, number, boolean, object, secret)
- ✅ Application listing and retrieval from API (10 applications found)
- ✅ Application export to HCL with 3.7KB output from 10 applications
- ✅ Comprehensive acceptance test framework with real API calls (36 tests)
- ✅ JSON export format for debugging
- ✅ Skip-dependencies support
- ✅ Secret masking in connector properties and variable secrets
- ✅ Terraform validation test framework created (5 tests)
- ✅ Variables terraform validation PASSING (fixed attribute names, empty blocks, mutable logic)

**Test Status**:
- Unit tests: 85/85 passing across all packages
- Acceptance tests: 36/36 passing (without terraform validation)
- Terraform validation: 1/5 passing (Variables only - Connectors/Applications/Flows need fixes)

