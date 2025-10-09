---
mode: agent
---
Used to define the main command name throughout the prompt.

commandName: "pingcli davinci convert"
Default path for the input flow, can be changed when using the prompt.

flowJsonPath: "path/to/flow.json"
Default path for the output HCL, can be changed when using the prompt.
outputPath: "path/to/output.hcl"

# Project Brief: PingOne DaVinci Flow to HCL Converter CLI Plugin

## Your Role

You are an expert Go software engineer with extensive experience in building CLI tools, working with REST APIs, and developing Terraform providers. Your task is to build a CLI plugin for the pingcli tool. You will follow Test-Driven Development (TDD) principles to ensure the code is robust, maintainable, and ready for enterprise use. Break up the parts into manageable sections, and ensure each part is completed before moving to the next.

## Project Overview

The goal is to create a comprehensive CLI command that exports a complete PingOne DaVinci environment configuration and converts it into HCL (HashiCorp Configuration Language). This HCL will be compatible with the DaVinci resources in the PingOne Terraform Provider. The tool will:

1. Export all DaVinci resources from an environment via API calls (flows, applications, connector instances, variables, flow policies, etc.)
2. Convert the exported JSON to valid HCL with proper Terraform resource syntax
3. Intelligently link dependencies between resources using Terraform references instead of hardcoded IDs
4. Validate the generated HCL against the Terraform provider schema

Key Technologies & Reference Materials

You will need to be familiar with the following:

    [Ping CLI Core](../../../../pingidentity/pingcli/cmd/) (pingcli): The target framework for this plugin. You will create a new command within this ecosystem. (https://github.com/pingidentity/pingcli)

    [Ping CLI Plugin System](../../../../pingidentity/pingcli/examples/plugin/): Understand how to create new commands for pingcli using the plugin architecture.

    [PingOne Go SDK](../../../../pingidentity/pingone-go-client/pingone): This is the primary library for interacting with the PingOne API and contains the necessary Go structs for unmarshaling DaVinci flow JSON. (https://github.com/pingidentity/pingone-go-client)

    [PingOne Terraform Provider](../../../../pingidentity/terraform-provider-pingone/internal/service/davinci/): The source code for the DaVinci resources should be used as a reference to understand how the Go SDK structs are mapped to HCL.

    [DaVinci OpenAPI Specification](./davinci-openapi.yaml): This is the source of truth for the structure of the DaVinci Flow JSON payload.
    
    [Sample DaVinci Flow JSON](./davinci-api-protect-reg-authn-flow.json): Use this to understand the real-world structure and data you will be working with.

    [Legacy DaVinci CLI Tool](../../../../patrickcping/dvtf-pingctl/): You will be looking to understand how the legacy tool handled similar conversions, especially regarding environment-specific values.

#{part1}
Part 1: Project Scaffolding and Command Structure (COMPLETE)

Your first task is to set up the project structure and create the basic CLI command.

    Initialize Project: Create a new Go module for the project.

    Project Layout: Structure the project using standard Go conventions (e.g., a cmd directory for the main command, internal for the core logic, and pkg if any code is intended to be shared).

    Create Cobra Command:

        Using the cobra library (which pingcli uses), scaffold a new command.

        The command should be callable like: ${commandName}.

        Define two flags for the command:

            --flow-json <path>: A required string flag pointing to the input DaVinci flow JSON file.

            --out <path>: An optional string flag for the output HCL file. If not provided, the output should go to stdout.

    Initial Implementation: The command's execution logic should, for now, simply print a message like "Executing DaVinci flow conversion for file: ${flowJsonPath}". This confirms the command structure is working.

    Write a Placeholder Test: Create a simple test file for the command to ensure the basic structure is testable.

**Status**: This part is complete. The project has a working CLI command structure with proper flags and tests.

#{part2}
Part 2: Complete DaVinci Resource Conversion with Terraform Validation (IN PROGRESS)

This part focuses on converting **all DaVinci resource types** to valid HCL, with complete structure mapping. This includes:

- **pingone_davinci_flow**: Complete flow structure with graphData, nodes, edges, settings, variables, input/output schemas
- **pingone_davinci_application**: Application configuration with OAuth settings, API keys
- **pingone_davinci_application_flow_policy**: Flow policies with distributions, success nodes, trigger configuration
- **pingone_davinci_connection**: Connector instances with properties (masking sensitive values)
- **pingone_davinci_variable**: Variables with context, data types, values

Each resource type must be fully mapped from its JSON structure to valid HCL. Additionally, this part includes integration tests with the actual Terraform provider to validate the generated HCL.

## Phase 2.1: Comprehensive Flow Structure Conversion (TDD)

    Create Test Case for Complete Flow:

        In internal/converter package, expand converter_test.go.

        Define test function TestCompleteFlowConversion.

        Inside the test, define a complete DaVinci flow JSON including:
            - Top-level metadata (name, description, flowId, etc.)
            - Complete graphData structure with nodes and edges
            - Settings object with all nested properties
            - Variables array
            - InputSchema array
            - OutputSchema structure

        Define the expected HCL output as a string, ensuring it captures all nested structures.

        Write the test assertion first: call Convert() function and assert output matches expected HCL. This test will fail initially.

    Map Complete Flow Structure:

        Reference the structs in pingone-go-client for the complete DaVinci flow structure.

        Review the DaVinci OpenAPI spec to understand all possible fields and their types.

        Ensure your Convert() function handles:
            - Nested JSON objects (converted to HCL blocks as appropriate)
            - Arrays of complex objects
            - Boolean, string, and numeric types
            - Null/optional fields

    Implement Deep Conversion Logic:

        Extend the Convert(flowJSON []byte) (string, error) function to handle:
            - Top-level flow attributes (name, description, enabled, etc.)
            - graphData structure (nodes, edges, positions, connections)
            - Settings object (all nested properties)
            - Variables array (with proper context and data types)
            - Input/output schemas
            - Flow metadata (created/updated timestamps, versions, etc.)

        Use appropriate HCL syntax per Terraform schema:
            - Simple attributes: `name = "value"`
            - All attributes have a type (string, bool, number, object, list), there should not be any large JSON. 
            - Blocks for repeated structures: `variable { ... }`

        Run tests until they pass.

    Handle Edge Cases:

        Test and handle flows with:
            - Empty or null fields
            - Large graphData structures (hundreds of nodes)
            - Special characters in strings
            - Boolean and numeric edge cases
            - Nested arrays and objects

## Phase 2.2: Application Resource Conversion (TDD)

    Create Test Case for Applications:

        Define test function TestApplicationConversion.

        Create sample DaVinci application JSON with:
            - Application metadata (name, apiKey settings)
            - OAuth configuration (grantTypes, redirectUris, scopes)
            - Created/updated timestamps

        Define expected HCL for pingone_davinci_application resource.

        Write test assertion and implement conversion logic.

    Implement Application Conversion:

        Create generateApplicationHCL() function.

        Handle OAuth configuration block:
            - enabled flag
            - values array with grantTypes, redirectUris, logoutUris, scopes

        Handle API key configuration.

        Mask sensitive values (client secrets, API keys).

    Test Application Conversion:

        Test with various OAuth configurations.

        Test with and without API keys.

        Test optional fields handling.

## Phase 2.3: Flow Policy Resource Conversion (TDD)

    Create Test Case for Flow Policies:

        Define test function TestFlowPolicyConversion.

        Create sample flow policy JSON with:
            - Policy metadata (name, status)
            - Trigger configuration (type, MFA/password settings)
            - Flow distributions (flow IDs, versions, weights, success nodes)

        Define expected HCL for pingone_davinci_application_flow_policy resource.

        Write test assertion and implement conversion logic.

    Implement Flow Policy Conversion:

        Create generateFlowPolicyHCL() function.

        Handle trigger block:
            - type
            - configuration with MFA and password settings

        Handle flow_policy block:
            - status
            - priority (optional)

        Handle policy_flow blocks (repeatable):
            - flow_id reference
            - version_id
            - weight
            - success_nodes

    Test Flow Policy Conversion:

        Test with single and multiple flow distributions.

        Test with various trigger types.

        Test optional fields.

## Phase 2.4: Connection (Connector Instance) Resource Conversion (TDD)

    Create Test Case for Connections:

        Define test function TestConnectionConversion.

        Create sample connector instance JSON with:
            - Connection metadata (name, connector_id)
            - Properties object (connector-specific configuration)
            - Sensitive fields (passwords, tokens, API keys)

        Define expected HCL for pingone_davinci_connection resource.

        Write test assertion and implement conversion logic.

    Implement Connection Conversion:

        Create generateConnectionHCL() function.

        Handle connector_id reference (to be resolved in Part 4).

        Handle properties map:
            - Convert to HCL map syntax
            - Identify and mask sensitive fields (add TODO comments)

        Common sensitive field patterns:
            - password, secret, token, apiKey, clientSecret, privateKey

    Test Connection Conversion:

        Test with various connector types.

        Test sensitive field masking.

        Test with empty or minimal properties.

## Phase 2.5: Variable Resource Conversion (TDD)

    Create Test Case for Variables:

        Define test function TestVariableConversion.

        Create sample variable JSON with:
            - Variable metadata (name, context, mutable)
            - Data type (string, number, boolean, object, secret)
            - Value (appropriate for data type)
            - Min/max (for numeric types)

        Define expected HCL for pingone_davinci_variable resource.

        Write test assertion and implement conversion logic.

    Implement Variable Conversion:

        Create generateVariableHCL() function.

        Handle context types:
            - company (environment-wide)
            - flow (flow-specific, requires flow_id reference)
            - flowInstance (runtime variable)
            - user (user-specific)

        Handle data types:
            - string, number, boolean: simple value assignment
            - object: HCL map or JSON encoding
            - secret: mask value with TODO comment

        Handle optional fields:
            - description
            - min/max (for numbers)
            - flow_id (for flow context)

    Test Variable Conversion:

        Test each context type.

        Test each data type.

        Test secret masking.

        Test optional fields.

## Phase 2.6: Multi-Resource Integration

    Create Comprehensive Test:

        Define test function TestMultiResourceConversion.

        Create JSON payload with all resource types.

        Define expected HCL with all resources in logical order.

        Verify all resources are generated correctly.

    Resource Ordering:

        Generate HCL in dependency order:
            1. Variables (no dependencies)
            2. Connections (no dependencies)
            3. Flows (reference connections and variables and other flows)
            4. Applications (standalone)
            5. Flow Policies (reference flows and applications)

    Test Multi-Resource Generation:

        Test with all resource types present.

        Test with subset of resource types.

        Test resource ordering is correct.

## Phase 2.2: Terraform Provider Integration Tests

    Set Up Terraform Test Environment:

        Create integration tests that validate generated HCL against actual Terraform.

        Write tests in a new file: internal/converter/terraform_integration_test.go.

        Tests require Terraform CLI to be installed in the test environment.

        Use t.Skip() if terraform command not found in PATH.

    Implement Terraform Validation Tests:

        Create test function TestTerraformValidateFlow:
            - Generate HCL from a complete test flow
            - Write HCL to temporary directory with provider configuration
            - Execute terraform init
            - Execute terraform validate
            - Assert no errors returned

        Create test function TestTerraformPlanFlow:
            - Generate HCL from test flow
            - Set up mock/test provider credentials
            - Execute terraform plan
            - Assert plan succeeds (even if it shows changes)

        Create test for syntax validation:
            - Test that settings uses attribute syntax (settings = {...})
            - Test that all resource references are valid
            - Test that all required attributes are present

        Create test for each

    Test Fixtures:

        Create a fixtures directory with:
            - Minimal provider.tf with PingOne provider configuration
            - variables.tf for required inputs
            - Sample flow JSON files for testing

    Mock Provider Configuration:

        Use test credentials or mock values to avoid requiring actual PingOne environments.

        Document how to run integration tests with real credentials (optional).

    Why This Matters:

        Unit tests verify logic, but Terraform validates actual HCL syntax and schema compliance.

        Integration tests catch issues like:
            - Incorrect attribute syntax (settings { vs settings =)
            - Invalid resource references
            - Schema violations (missing required fields, wrong types)
            - Terraform version compatibility issues

        These tests provide confidence for production Terraform workflows.

    Example Test Structure:

        ```go
        func TestTerraformValidateFlow(t *testing.T) {
            // Check if terraform is available
            if _, err := exec.LookPath("terraform"); err != nil {
                t.Skip("Terraform not found in PATH, skipping validation test")
            }

            // Generate HCL from test flow
            flowJSON := readTestFixture(t, "complete-flow.json")
            hcl, err := Convert(flowJSON)
            require.NoError(t, err)

            // Create temp directory with Terraform config
            tmpDir := t.TempDir()
            writeFile(t, filepath.Join(tmpDir, "provider.tf"), providerConfig)
            writeFile(t, filepath.Join(tmpDir, "flow.tf"), hcl)

            // Run terraform init
            cmd := exec.Command("terraform", "init")
            cmd.Dir = tmpDir
            output, err := cmd.CombinedOutput()
            require.NoError(t, err, "terraform init failed: %s", output)

            // Run terraform validate
            cmd = exec.Command("terraform", "validate")
            cmd.Dir = tmpDir
            output, err = cmd.CombinedOutput()
            require.NoError(t, err, "terraform validate failed: %s", output)
        }
        ```

**Status**: Partially complete. Basic flow conversion works. Need to expand to handle all DaVinci resource types (applications, flow policies, connections, variables) and add Terraform validation tests.

#{part3}
Part 3: Full DaVinci Environment Export via API (NEW)

This part adds API integration to export all DaVinci resources from a PingOne environment, not just flows. This creates a complete export of the DaVinci configuration.

## Prerequisites

    API Client Setup Using PingCLI Authentication:

        **IMPORTANT**: As a pingcli plugin, this tool should leverage pingcli's existing authentication mechanisms, not reinvent them.

        Use pingcli's authentication patterns:
            - Config file: `~/.pingidentity/config` (or `PINGCLI_CONFIG_PATH`)
            - Profile support: `--profile <name>` or `PINGCLI_PROFILE` env var
            - Environment variables:
                - `PINGONE_CLIENT_ID`
                - `PINGONE_CLIENT_SECRET`
                - `PINGONE_ENVIRONMENT_ID`
                - `PINGONE_REGION`
            - Command-line flags (follow pingcli conventions):
                - `--client-id`
                - `--client-secret`
                - `--environment-id`
                - `--region`

        Create internal/api package with API client factory.

        Reference pingcli's authentication code:
            - See `pingcli/cmd/platform/` for authentication patterns
            - Use similar credential precedence: flags > env vars > config file > profile
            - Reuse pingcli's config file parsing if available

        Use pingone-go-client SDK for all API calls.

        Handle token acquisition and refresh automatically.

    New Command Flags:

        Add flags following pingcli conventions (these should already exist in pingcli framework):
            - `--export`: Enable API export mode (alternative to --flow-json)
            - `--environment-id <uuid>`: Target PingOne environment (or use from config/env)
            - `--region <code>`: PingOne region (or use from config/env)
            - `--client-id <id>`: OAuth client ID (or use from config/env)
            - `--client-secret <secret>`: OAuth client secret (or use from config/env)
            - `--profile <name>`: Config profile to use (pingcli standard)

        Flags should be optional if values are available from config file or environment variables.

        Follow pingcli's flag naming and structure conventions.

    Selective Export Flags (Future Enhancement):

        Add resource filtering flags for granular export control:
            - `--include-flows <comma-separated-ids-or-names>`: Export only specified flows
            - `--include-applications <comma-separated-ids-or-names>`: Export only specified applications
            - `--include-connections <comma-separated-ids-or-names>`: Export only specified connections
            - `--include-variables <comma-separated-ids-or-names>`: Export only specified variables
            - `--exclude-flows <comma-separated-ids-or-names>`: Exclude specified flows from export
            - `--exclude-applications <comma-separated-ids-or-names>`: Exclude specified applications
            - `--exclude-connections <comma-separated-ids-or-names>`: Exclude specified connections
            - `--exclude-variables <comma-separated-ids-or-names>`: Exclude specified variables
            - `--exclude-dependencies`: <boolean> Ignore all dependencies of selected resources (default: false)

        Flag behavior:
            - Include flags: Export ONLY specified resources (plus dependencies if --with-dependencies)
            - Exclude flags: Export everything EXCEPT specified resources
            - Include + Exclude: Include takes precedence, then apply exclusions
            - IDs or names: Support both resource IDs (UUIDs) and resource names (fuzzy matching)

        Use Case Examples:
            - Export single flow with dependencies: `--include-flows "Registration Flow" --with-dependencies`
            - Export multiple flows without dependencies: `--include-flows "flow1,flow2" --no-dependencies`
            - Export all except test resources: `--exclude-flows "test" --exclude-applications "test"`

        Implementation Notes:
            - This will be implemented in Phase 3.6 (after basic export works)
            - Requires HAL link parsing to discover dependencies
            - Requires dependency graph from Part 4 to be functional first

## Phase 3.1: Export Flow Resources

    API Client for Flows:

        Create internal/api/flows.go.

        Implement ListFlows() function to retrieve all flows in an environment.

        Implement GetFlowWithVersions() to get flow details including published versions.

        Handle pagination for environments with many flows.

    Test Flow Export:

        Create internal/api/flows_test.go.

        Write tests using mock HTTP client or test server.

        Test successful flow retrieval.

        Test error handling (network errors, auth failures, etc.).

    Convert Exported Flows:

        Update Convert() to handle API response format.

        Generate HCL for all exported flows.

        Include resource names based on flow names (sanitized for Terraform).

## Phase 3.2: Export Connector Instances

    API Client for Connector Instances:

        Create internal/api/connector_instances.go.

        Implement ListConnectorInstances() to retrieve all connector instances.

        Handle sensitive properties (credentials should be marked as TODO placeholders).

    HCL Generation for Connector Instances:

        Create generateConnectorInstanceHCL() function.

        Generate pingone_davinci_connection resource blocks.

        Handle connector-specific properties correctly.

        Mask sensitive values (passwords, tokens, etc.).

    Test Connector Instance Export:

        Write tests for API client.

        Write tests for HCL generation.

        Test with various connector types.

## Phase 3.3: Export Variables

    API Client for Variables:

        Create internal/api/variables.go.

        Implement ListVariables() to retrieve all flow and company variables.

        Handle different variable contexts (flow, company, user, flowInstance).

    HCL Generation for Variables:

        Create generateVariableHCL() function.

        Generate pingone_davinci_variable resource blocks.

        Handle different data types (string, number, boolean, object, secret).

        Mask secret variable values.

    Test Variable Export:

        Write tests for API client.

        Write tests for HCL generation.

        Test with different variable contexts and types.

## Phase 3.4: Export Applications and Flow Policies

    API Client for Applications:

        Create internal/api/applications.go.

        Implement ListApplications() to retrieve DaVinci applications.

        Implement GetApplicationFlowPolicies() for application flow policies.

    HCL Generation for Applications:

        Create generateApplicationHCL() function.

        Generate pingone_davinci_application resource blocks.

        Handle OAuth configuration, API keys, etc.

    HCL Generation for Flow Policies:

        Create generateFlowPolicyHCL() function.

        Generate pingone_davinci_application_flow_policy resource blocks.

        Handle flow distributions, success nodes, trigger configuration.

    Test Application and Flow Policy Export:

        Write tests for API clients.

        Write tests for HCL generation.

        Test with various application configurations.

## Phase 3.5: Orchestration and Output

    Export Orchestrator:

        Create internal/exporter package.

        Create Exporter struct that coordinates all API calls.

        Implement Export() method that:
            - Authenticates with PingOne API
            - Exports all resource types in parallel (where possible)
            - Collects all exported data
            - Returns structured export result

    Combined HCL Generation:

        Update converter to accept structured export data.

        Generate HCL for all resources in logical order:
            1. Variables (may be referenced by other resources)
            2. Connector instances (referenced by flows)
            3. Flows (may reference variables and connectors)
            4. Applications (reference flows)
            5. Flow policies (reference flows and applications)

        Include header comments with export metadata (timestamp, environment ID, etc.).

    CLI Integration:

        Update command execution logic to support --export mode.

        Validate required flags (environment-id, region, credentials).

        Handle export errors gracefully with user-friendly messages.

    Integration Tests:

        Create end-to-end test that:
            - Uses mock API responses
            - Exports all resource types
            - Generates complete HCL
            - Validates HCL structure

## Phase 3.6: Selective Export and Dependency Discovery (Future Enhancement)

    HAL Link Parsing:

        **IMPORTANT**: PingOne API responses use HAL (Hypertext Application Language) format with `_links` sections.

        Parse HAL links in API responses to discover relationships:
            - Flow resources have links to: connector instances, variables, subflows
            - Applications have links to: flow policies
            - Flow policies have links to: flows
            - Variables (flow context) have links to: flows

        Create internal/api/hal.go with HAL link parser.

        Implement DiscoverDependencies() function that:
            - Takes a resource and its API response
            - Extracts all HAL links
            - Returns list of dependent resource IDs and types

        Example HAL link structure:
            ```json
            {
              "id": "flow123",
              "name": "Registration Flow",
              "_links": {
                "self": { "href": "/environments/env123/flows/flow123" },
                "connectorInstances": { "href": "/environments/env123/connectorInstances?flowId=flow123" },
                "variables": { "href": "/environments/env123/variables?flowId=flow123" }
              }
            }
            ```

    Resource Filter Implementation:

        Create internal/filter package.

        Create ResourceFilter struct with methods:
            - ShouldInclude(resourceType, resourceID, resourceName) bool
            - ApplyInclusions(resources) filteredResources
            - ApplyExclusions(resources) filteredResources

        Implement filter logic:
            - Parse include/exclude flags into filter rules
            - Support ID matching (exact UUID match)
            - Support name matching (case-insensitive substring or fuzzy match)
            - Support comma-separated lists

    Dependency Tracking for Filtered Exports:

        When --with-dependencies is true (default):
            - Start with explicitly included resources
            - Use HAL links to discover direct dependencies
            - Recursively discover transitive dependencies
            - Add all discovered dependencies to export set

        Create DependencyDiscoverer that:
            - Takes initial resource set (from include filters)
            - Calls DiscoverDependencies() for each resource
            - Builds complete dependency tree
            - Returns expanded resource set

        Example workflow:
            1. User specifies: `--include-flows "Registration Flow"`
            2. Export fetches "Registration Flow" resource
            3. Parse HAL links to find: connector instances, variables, subflows
            4. Fetch those resources and their HAL links
            5. Continue recursively until all dependencies found
            6. Export complete set of resources

    Update Exporter for Selective Export:

        Modify Exporter.Export() to accept filter criteria.

        Implement two-phase export:
            - Phase 1: Fetch specified resources (matching include/exclude filters)
            - Phase 2: If --with-dependencies, discover and fetch dependencies via HAL links

        Track discovered vs. explicitly selected resources (for reporting).

    CLI Integration:

        Add new flags to command definition.

        Validate flag combinations:
            - Cannot use include and exclude for same resource type
            - At least one resource must be selected

        Display export summary:
            - Number of resources explicitly selected
            - Number of dependencies discovered
            - List of resource types being exported

    Testing Selective Export:

        Create tests for ResourceFilter:
            - Test inclusion logic with IDs
            - Test inclusion logic with names
            - Test exclusion logic
            - Test combination of include + exclude

        Create tests for HAL link parsing:
            - Test parsing various HAL response formats
            - Test handling missing links
            - Test handling invalid link formats

        Create tests for dependency discovery:
            - Test discovering single-level dependencies
            - Test recursive dependency discovery
            - Test with --no-dependencies flag
            - Test circular dependency handling

    Example Test Scenarios:

        Scenario 1: Export single flow with dependencies
            - Include: One flow by name
            - Expect: Flow + its connectors + its variables + any subflows

        Scenario 2: Export application without dependencies
            - Include: One application by ID
            - Flag: --no-dependencies
            - Expect: Only the application resource

        Scenario 3: Export everything except test resources
            - Exclude: Resources with "test" in name
            - Expect: All resources except those matching "test"

**Status**: Not started. This is a future enhancement after basic export (Phase 3.1-3.5) is working. Will be implemented after Part 4 (dependency resolution) is functional.

#{part4}
Part 4: Dependency Resolution and Terraform References (Previously Part 3)

Now that we have a full export with all resources, we can intelligently link dependencies using Terraform references instead of hardcoded IDs.

## Phase 4.1: Build Dependency Graph

    Create Dependency Resolver:

        Create internal/resolver package.

        Create DependencyGraph struct that tracks all resources and their relationships.

        Implement AddResource() to register exported resources with their IDs and types.

        Implement FindReferences() to discover all dependency references in exported data.

    Identify Dependency Types:

        Connection IDs in flow nodes → connector instance resources
        Variable IDs in flows → variable resources
        Subflow IDs in flows → other flow resources
        Flow IDs in flow policies → flow resources
        Application IDs in flow policies → application resources

    Parse Multiple Dependency Sources:

        **JSON Structure References**: Parse flow graphData nodes to find:
            - connectionId fields → connector instances
            - Variable references in node properties
            - Subflow node references → other flows

        **HAL Links** (from API responses): Extract dependency information from `_links`:
            - More reliable than parsing JSON structure
            - Provides bidirectional relationships
            - Includes resource type information
            - See Phase 3.6 for HAL link structure

        **Hybrid Approach**: Use both sources for completeness:
            - HAL links for high-level relationships (flows → connectors, flows → variables)
            - JSON structure for detailed node-level dependencies
            - Cross-validate both sources for accuracy

    Test Dependency Detection:

        Write tests that verify dependency detection for each resource type.

        Test with complex flows that have multiple dependencies.

        Test with circular dependencies (should be detected and handled).

        Test with HAL link parsing (mock API responses).

        Test with JSON structure parsing (flow graphData).

## Phase 4.2: Generate Terraform References

    Reference Syntax:

        Instead of hardcoded IDs, generate Terraform references:
            - Connection ID: `pingone_davinci_connection.my_connector.id`
            - Variable ID: `pingone_davinci_variable.my_var.id`
            - Flow ID: `pingone_davinci_flow.my_flow.id`

    Resource Naming:

        Generate valid Terraform resource names from human-readable names.

        Sanitize names (remove special characters, spaces, etc.).

        Ensure uniqueness (append counter if needed).

        Create name → ID mapping for reference generation.

    Update HCL Generation:

        Modify generateFlowHCL() to use references for:
            - connection_id in flow nodes
            - variable references
            - subflow references

        Modify generateFlowPolicyHCL() to use references for:
            - flow_id in flow distributions
            - application_id

        Ensure all generated references are valid Terraform syntax.

    Test Reference Generation:

        Write tests that verify correct reference syntax.

        Test with various dependency scenarios.

        Verify that references match actual resource names in HCL.

## Phase 4.3: Handle Missing Dependencies

    Detect Orphaned References:

        Identify references to resources not included in the export.

        This can happen if:
            - Resource was deleted but still referenced
            - Export filter excluded some resources (selective export in Phase 3.6)
            - Resource is in different environment

    Generate Placeholder Comments:

        For missing dependencies, generate TODO comments:
            ```hcl
            connection_id = "" # TODO: Reference missing for original ID: abc123
            ```

        Include original ID and resource type in comment.

        Optionally warn user about missing dependencies.

    Handle Selective Export Edge Cases:

        When using selective export (Phase 3.6), some dependencies may be intentionally excluded.

        Distinguish between:
            - **Missing by exclusion**: User explicitly excluded via --exclude flag
            - **Missing by selection**: User used --include without dependencies
            - **Actually missing**: Resource doesn't exist in environment

        Different handling for each case:
            - Missing by exclusion: Generate TODO with note "excluded from export"
            - Missing by selection: Generate TODO with note "not included in export filters"
            - Actually missing: Generate TODO with note "not found in environment"

        Example placeholders:
            ```hcl
            # Missing by user exclusion
            connection_id = "" # TODO: Reference to "PingOne Connector" (ID: abc123) was excluded from export

            # Missing by selective export
            variable_id = "" # TODO: Reference to "apiKey" (ID: xyz789) was not included in export filters

            # Actually missing
            subflow_id = "" # TODO: Reference to flow ID def456 not found in environment
            ```

    Test Missing Dependency Handling:

        Write tests with incomplete exports.

        Verify placeholders are generated correctly.

        Test warning/error messages.

        Test selective export scenarios with missing dependencies.

## Phase 4.4: Validate Dependency Graph

    Circular Dependency Detection:

        Implement cycle detection in dependency graph.

        Terraform cannot handle circular dependencies directly.

        Detect and report circular dependencies to user.

    Dependency Ordering:

        Order resources in generated HCL based on dependencies.

        Resources with no dependencies first.

        Dependent resources after their dependencies.

        This improves readability and Terraform plan output.

    Test Dependency Validation:

        Write tests for cycle detection.

        Write tests for dependency ordering.

        Test with complex multi-resource scenarios.

**Status**: Not started. Requires completion of Part 3 (full export).

#{part5}
Part 5: Final Integration and Error Handling (Previously Part 4)

Finally, connect the logic to the CLI command and ensure it is robust.

    Connect Logic to Command: Update your Cobra command's execution logic from Part 1. It should now:

        Read the contents of the file specified by the --flow-json flag.

        Pass the file contents to the converter.Convert() function.

        Handle any errors returned from the converter (e.g., invalid JSON) and print a user-friendly error message to stderr.

    Handle Output:

        If the --out flag is provided, write the resulting HCL string to that file.

        If the --out flag is not provided, print the HCL string to stdout.

    Write Integration Tests: Create tests for the command itself. You can do this by executing the root command in your test code and capturing the output/error streams to verify:

        A valid flow JSON produces the expected HCL to stdout.

        Using the --out flag creates a file with the correct content.

        Providing a path to a non-existent JSON file returns an error.

        Providing a malformed JSON file returns an error.

    Plan for Deprecation:

        Ensure any logic that might be deprecated (like API client initialization, although not needed yet) is encapsulated in its own package or struct. Add a comment indicating that this component is a candidate for future replacement by a shared function from the main pingcli application.

#{part5}
Part 5: Final Integration and Error Handling (Previously Part 4)

Finally, connect all the logic and ensure the complete tool is robust and production-ready.

## Phase 5.1: Complete CLI Integration

    Connect Logic to Command: Update your Cobra command's execution logic from Part 1. It should now:

        Support two modes:
            - File mode: Read flow from --flow-json flag (original behavior)
            - Export mode: Use --export flag with API credentials to export entire environment

        For file mode:
            - Read the contents of the file specified by the --flow-json flag
            - Pass the file contents to converter.Convert() function
            - Handle any errors (invalid JSON) with user-friendly messages

        For export mode:
            - Validate required flags (environment-id, region, credentials)
            - Initialize API client with credentials
            - Call exporter.Export() to retrieve all resources
            - Pass exported data to converter.ConvertExport() function
            - Apply dependency resolution
            - Handle API errors gracefully

    Handle Output:

        If the --out flag is provided, write the resulting HCL string to that file.

        If the --out flag is not provided, print the HCL string to stdout.

        For export mode, optionally support --out-dir flag to write multiple .tf files (one per resource type).

    Error Handling and User Experience:

        Provide clear error messages for common issues:
            - Missing required flags
            - Invalid credentials
            - Network errors
            - API rate limiting
            - Invalid JSON/API responses

        Include suggestions for resolving errors.

        Support --verbose flag for detailed logging.

        Support --dry-run flag to show what would be exported without making API calls.

## Phase 5.2: Comprehensive Integration Tests

    Create Command Integration Tests:

        Write tests in cmd/convert_test.go.

        Test file mode conversion:
            - Valid flow JSON produces expected HCL to stdout
            - Using --out flag creates file with correct content
            - Providing path to non-existent JSON file returns error
            - Providing malformed JSON file returns error

        Test export mode (with mock API):
            - Successful export generates complete HCL
            - Missing credentials returns helpful error
            - API errors are handled gracefully
            - Dependency resolution works correctly

        Test selective export mode (with mock API, Phase 3.6):
            - Export with --include-flows filters correctly
            - Export with --with-dependencies includes all dependencies
            - Export with --no-dependencies excludes dependencies
            - Export with --exclude-* filters correctly
            - Invalid filter combinations return helpful errors
            - HAL link parsing discovers dependencies correctly

    End-to-End Tests:

        Create a comprehensive test that:
            - Exports a mock environment (all resource types)
            - Resolves all dependencies
            - Generates complete HCL
            - Validates HCL with Terraform (if available)
            - Verifies no hardcoded IDs in output (all are references or TODOs)

        Create selective export test that:
            - Exports single flow with dependencies via HAL links
            - Verifies all required dependencies are included
            - Verifies excluded resources are not included
            - Checks TODO comments for missing dependencies

## Phase 5.3: Documentation and Examples

    Update README:

        Document all command modes and flags.

        Provide examples for both file mode and export mode.

        Document selective export functionality (Phase 3.6):
            - Include/exclude filter syntax
            - Dependency discovery with HAL links
            - Use case examples

        Include authentication setup instructions.

        Document how to use generated HCL with Terraform.

    Create Example Configurations:

        Include sample flow JSON files in examples/ directory.

        Include sample provider configuration for Terraform.

        Include sample terraform.tfvars for common scenarios.

        Include selective export examples:
            - Export single flow: `examples/selective-export-single-flow.sh`
            - Export application with policies: `examples/selective-export-app.sh`
            - Export without dependencies: `examples/selective-export-no-deps.sh`

    Troubleshooting Guide:

        Document common errors and solutions.

        Include debugging tips.

        Document HAL link parsing issues (if any).

        Document filter matching behavior (ID vs name, case sensitivity, fuzzy matching).

        Provide contact/support information.

## Phase 5.4: Performance and Optimization

    Concurrent API Calls:

        Implement concurrent fetching of resources (flows, connectors, variables, etc.).

        Use goroutines with proper synchronization.

        Implement rate limiting to respect API quotas.

    Caching:

        Cache API responses during export to avoid redundant calls.

        Optionally support --cache flag to save API responses to disk.

        Support --use-cache flag to use cached data for development/testing.

    Progress Reporting:

        For large exports, show progress indicators.

        Report number of resources exported.

        Estimate time remaining for large exports.

## Phase 5.5: Plan for Deprecation and Future Enhancements

    Deprecation Strategy:

        Ensure any logic that might be deprecated (like API client initialization) is encapsulated in its own package.

        Add comments indicating components that are candidates for future replacement by shared functions from the main pingcli application.

        Design interfaces that allow easy swapping of implementations.

    Future Enhancements:

        Support for selective export (filter by resource type, name patterns, etc.) - **Implemented in Phase 3.6**

        HAL link-based dependency discovery - **Implemented in Phase 3.6**

        Import mode: Generate HCL that imports existing resources into Terraform state.

        Diff mode: Compare environment state with existing HCL/Terraform state.

        Template mode: Generate parameterized HCL with variables for multi-environment deployments.

        Validation mode: Check exported HCL against organization policies/best practices.

        Advanced filtering:
            - Regex patterns for resource names
            - Filter by resource attributes (e.g., enabled flows only)
            - Filter by tags or metadata
            - Time-based filtering (resources modified after date)

        Dependency visualization:
            - Generate dependency graph diagrams (DOT/Graphviz format)
            - Show which resources depend on selected resources
            - Identify circular dependencies visually

        Export profiles:
            - Save common filter combinations as named profiles
            - Share export profiles across teams
            - Version control for export configurations

**Status**: Not started. Requires completion of Parts 2-4.

#{part6}
Part 6: Production Readiness and Release (NEW)

Prepare the tool for production use and public release.

## Phase 6.1: Security Hardening

    Credential Management:

        Never log or print credentials.

        Support credential storage in system keychain/credential manager.

        Clear sensitive data from memory after use.

        Validate credential format before making API calls.

    Sensitive Data Handling:

        Identify and mask all sensitive fields in exported data (passwords, tokens, secrets).

        Ensure sensitive data is not written to logs or error messages.

        Document what data is considered sensitive.

## Phase 6.2: Testing and Quality Assurance

    Achieve High Test Coverage:

        Aim for >90% test coverage across all packages.

        Focus on critical paths (conversion logic, API calls, dependency resolution).

        Add edge case tests (empty responses, malformed data, etc.).

    Manual QA:

        Test with real PingOne environments of various sizes.

        Test with environments containing all resource types.

        Verify generated HCL works with Terraform apply.

        Test error scenarios (network failures, permission errors, etc.).

    Performance Testing:

        Benchmark with large environments (100+ flows, 1000+ nodes).

        Optimize slow operations.

        Document performance characteristics and limitations.

## Phase 6.3: Release Preparation

    Versioning:

        Implement semantic versioning (SemVer).

        Add version command: pingcli davinci version.

        Include version in generated HCL comments.

    Release Artifacts:

        Build binaries for multiple platforms (Linux, macOS, Windows).

        Create installation instructions for each platform.

        Publish to package managers if applicable (Homebrew, apt, etc.).

    Changelog:

        Maintain detailed CHANGELOG.md.

        Document all features, bug fixes, and breaking changes.

        Include migration guides for major version updates.

**Status**: Future work after core functionality is complete and stable.