# Roadmap: Feature Additions from dvtf-pingctl Analysis

**Last Updated**: 2025-10-27  
**Status**: Planning Phase  
**Priority**: High-value features for production readiness

---

## Overview

This document outlines capabilities to add to davinci-terraform-converter based on analysis of the legacy dvtf-pingctl tool and user workflow requirements.

**Current State**: davinci-terraform-converter has superior API export, resource coverage, and import block generation capabilities.

**Gap**: Missing validation-focused features that aid debugging and CI/CD integration.

---

## Priority 1: Critical for Production

### 1.1 Validation Command

**Status**: Not Implemented  
**Priority**: P0 (Critical)  
**Estimated Effort**: 4-6 hours

#### Description

Add standalone `validate` command that validates DaVinci flow JSON files for Terraform provider compatibility without generating HCL.

#### Use Cases

- Pre-export validation in CI/CD pipelines
- Debugging Terraform state corruption issues
- Validation gates before applying changes
- Quick syntax/structure verification

#### Implementation

**New Command Structure**:
```bash
# Validate flow JSON file
davinci-convert validate --flow-json flow.json

# Validate specific field (advanced debugging)
davinci-convert validate --flow-json flow.json --field flow_configuration_json

# Validate from stdin
cat flow.json | davinci-convert validate
```

**Return Codes** (matching dvtf-pingctl):
- `0` - Validation successful, no warnings
- `1` - Validation failed
- `2` - Validation successful with warnings

**Field Options**:
- `flow_json` (default) - Full flow validation
- `flow_configuration_json` - Configuration-only validation
- `flow_export_json` - Export format validation

#### Files to Create/Modify

- `cmd/validate.go` - New validation command
- `cmd/validate_test.go` - Validation tests
- `internal/validator/` - New validation package
  - `flow_validator.go` - Flow validation logic
  - `field_validator.go` - Field-specific validation
  - `error_reporter.go` - Detailed error reporting

#### Success Criteria

- [ ] Validates flow JSON against Terraform provider schema
- [ ] Supports field-specific validation for debugging
- [ ] Returns appropriate exit codes (0, 1, 2)
- [ ] Provides detailed error messages with line numbers
- [ ] Works with file input and stdin
- [ ] Integrates with PingCLI plugin mode

---

### 1.2 Multi-File Output (Module Structure)

**Status**: Not Implemented (single file only)  
**Priority**: P1 (High)  
**Estimated Effort**: 8-12 hours

#### Description

Generate Terraform modules with resources split into logical files instead of single monolithic HCL file.

#### Use Cases

- Better code organization for large environments
- Easier code review and version control
- Separate concerns (flows, variables, connectors, policies)
- Module reusability

#### Implementation

**CLI Flag**:
```bash
# Generate module structure
davinci-convert export --out-dir ./generated --module

# Single file (current default)
davinci-convert export --out davinci.tf
```

**Generated Structure** (matching dvtf-pingctl pattern):
```
generated/
├── versions.tf                          # Provider version constraints
├── vars.tf                              # Module input variables
├── davinci_flows.tf                     # Flow resources
├── davinci_flow_vars.tf                 # Flow-specific variables
├── davinci_flow_outputs.tf              # Flow outputs
├── davinci_variables.tf                 # Variable resources
├── davinci_variable_vars.tf             # Variable value overrides
├── davinci_connectors.tf                # Connector instance resources
├── davinci_connection_property_vars.tf  # Connector property overrides
├── davinci_applications.tf              # Application resources
├── davinci_flow_policies.tf             # Flow policy resources
└── assets/
    └── flows/
        ├── main_flow.json               # Individual flow JSON files
        ├── subflow_one.json
        └── subflow_two.json
```

**Module Variables**:
- Expose all connector properties as variables with defaults
- Expose all variable values as variables with defaults
- Required: `environment_id`
- Optional overrides for all sensitive/configurable values

#### Files to Create/Modify

- `cmd/export.go` - Add `--module` and `--out-dir` flags
- `internal/exporter/module_generator.go` - New module generator
- `internal/exporter/file_splitter.go` - Resource-to-file mapping
- `internal/converter/*_converter.go` - Variable extraction support

#### Success Criteria

- [ ] Generate multi-file module structure
- [ ] Extract variables for connector properties
- [ ] Extract variables for variable values
- [ ] Generate versions.tf with provider constraints
- [ ] Generate module outputs (flow IDs, variable IDs, etc.)
- [ ] Maintain import blocks in module mode
- [ ] Support both `--out` (single file) and `--out-dir` (module)
- [ ] Copy individual flow JSON files to assets/flows/

---

## Priority 2: Enhanced User Experience

### 2.1 Selective Export (Include/Exclude)

**Status**: Not Implemented  
**Priority**: P1 (High)  
**Estimated Effort**: 6-8 hours

#### Description

Allow users to selectively export specific resources or resource types.

#### Use Cases

- Export only flows (exclude variables, connectors)
- Export specific flows by name/ID pattern
- Exclude sensitive resources
- Incremental migration workflows

#### Implementation

**CLI Flags**:
```bash
# Export specific resource types
davinci-convert export --resources flows,variables

# Export specific flows by name pattern
davinci-convert export --include-flows "Registration*,Login*"

# Exclude specific flows
davinci-convert export --exclude-flows "Test*,Draft*"

# Export specific flows by ID
davinci-convert export --flow-ids "uuid1,uuid2,uuid3"
```

**Resource Type Options**:
- `flows`
- `variables`
- `connectors`
- `applications`
- `flow-policies`

#### Files to Create/Modify

- `cmd/export.go` - Add selection flags
- `internal/exporter/filter.go` - New filtering logic
- `internal/exporter/exporter.go` - Apply filters before conversion

#### Success Criteria

- [ ] Filter by resource type
- [ ] Filter flows by name pattern (glob/regex)
- [ ] Filter flows by ID list
- [ ] Maintain dependency resolution with filtered resources
- [ ] Generate TODO comments for excluded dependencies
- [ ] Document filtering behavior

---

### 2.2 Dry-Run Mode

**Status**: Not Implemented  
**Priority**: P2 (Medium)  
**Estimated Effort**: 2-4 hours

#### Description

Preview export operations without generating output files.

#### Use Cases

- Verify API connectivity
- Preview resource counts
- Validate filters before full export
- CI/CD validation

#### Implementation

**CLI Flag**:
```bash
# Dry run with summary
davinci-convert export --dry-run

# Dry run with resource list
davinci-convert export --dry-run --verbose
```

**Output Example**:
```
Dry Run: Export Preview
========================
Worker Environment:  abc-123
Export Environment:  def-456
Region:              NA

Resources Found:
  Flows:             12
  Variables:         34
  Connectors:        8
  Applications:      3
  Flow Policies:     15

Total Resources:     72
Estimated Size:      ~450 KB

Dependencies:
  ✓ All dependencies resolved
  ⚠ 2 flows reference missing variables (will generate TODO)

Run without --dry-run to generate HCL.
```

#### Files to Create/Modify

- `cmd/export.go` - Add `--dry-run` flag
- `internal/exporter/exporter.go` - Add dry-run mode
- `internal/exporter/summary.go` - Generate summary report

#### Success Criteria

- [ ] Preview resource counts
- [ ] Show dependency warnings
- [ ] Estimate output size
- [ ] Verify API connectivity
- [ ] No file generation in dry-run mode
- [ ] Exit code 0 on success, 1 on validation failure

---

### 2.3 Verbose Logging

**Status**: Partially Implemented (logger exists)  
**Priority**: P2 (Medium)  
**Estimated Effort**: 3-4 hours

#### Description

Enhanced logging for debugging and progress tracking.

#### Implementation

**CLI Flag**:
```bash
# Verbose output
davinci-convert export --verbose

# Debug output
davinci-convert export --debug

# Log to file
davinci-convert export --log-file export.log
```

**Log Levels**:
- `ERROR` - Errors only (default)
- `WARN` - Warnings and errors
- `INFO` - Progress messages (`--verbose`)
- `DEBUG` - Detailed debug info (`--debug`)

**Enhanced Output**:
```
INFO  Authenticating with PingOne API (region: NA)
INFO  Connected to environment: def-456
INFO  Fetching flows... (12 found)
INFO  Fetching variables... (34 found)
DEBUG Processing flow: Registration Flow (id: abc-123)
DEBUG   - Found 5 nodes
DEBUG   - Found 2 subflow references
DEBUG   - Found 3 connector dependencies
INFO  Resolving dependencies...
DEBUG Dependency graph: 72 nodes, 145 edges
INFO  Generating HCL...
INFO  ✓ Export complete: 72 resources, 450 KB
```

#### Files to Create/Modify

- `cmd/export.go` - Add logging flags
- `cmd/davinci_to_hcl.go` - Add logging flags
- `internal/api/client.go` - Add debug logging
- `internal/exporter/exporter.go` - Add progress logging

#### Success Criteria

- [ ] Support `--verbose` and `--debug` flags
- [ ] Support `--log-file` for file output
- [ ] Progress messages during long operations
- [ ] Debug logging for troubleshooting
- [ ] Compatible with PingCLI logger integration

---

## Priority 3: Advanced Features

### 3.1 Configuration File Support

**Status**: Not Implemented  
**Priority**: P2 (Medium)  
**Estimated Effort**: 4-6 hours

#### Description

Support configuration file for commonly used settings (similar to dvtf-pingctl's `.dvtf-pingctl.yaml`).

#### Implementation

**Config File**: `.davinci-convert.yaml`

```yaml
# API Configuration
api:
  worker_environment_id: "abc-123"
  export_environment_id: "def-456"
  region_code: "NA"
  # Secrets should still use environment variables

# Export Options
export:
  services:
    - pingone-davinci
  skip_dependencies: false
  generate_imports: true
  output: "./generated"
  module: true

# Filtering
filters:
  resources:
    - flows
    - variables
    - connectors
  include_flows:
    - "Registration*"
    - "Login*"
  exclude_flows:
    - "Test*"

# Logging
logging:
  level: INFO
  file: "./export.log"
```

#### Files to Create/Modify

- `internal/config/config.go` - Config file parser
- `cmd/export.go` - Load config file
- `cmd/davinci_to_hcl.go` - Load config file
- `.davinci-convert.yaml.example` - Example config file

#### Success Criteria

- [ ] Load config from `.davinci-convert.yaml`
- [ ] CLI flags override config file values
- [ ] Environment variables override config file
- [ ] Validate config file schema
- [ ] Document configuration options

---

### 3.2 Enhanced Connector Property Handling

**Status**: ✅ Implemented (properties export working)  
**Priority**: P3 (Low) - Enhancement only  
**Estimated Effort**: 4-6 hours

#### Description

**Current State**: Connector properties are successfully exported from API. The tool reads property values from connector instances and generates them in HCL format with proper masking for secrets.

**Enhancement Opportunity**: Add schema-based validation and type checking using the embedded `connector_schema.json` to improve property handling.

#### What Works Now

✅ Properties exported from API responses  
✅ Secret masking (e.g., `clientSecret` → `"TODO: Replace with actual client secret"`)  
✅ Type preservation (string, bool, number)  
✅ JSON encoding in HCL  
✅ Nil value handling

**Example Current Output**:
```hcl
resource "pingone_davinci_connector_instance" "pingcli__PingOne-0020-Protect" {
  environment_id = var.environment_id
  name           = "PingOne Protect"
  
  connector = {
    id = "pingOneRiskConnector"
  }
  
  properties = jsonencode({
    "clientId"     : "d2671735-e614-486c-9ae6-bdd72c5cd716",
    "clientSecret" : "TODO: Replace with actual client secret",  # ✅ Masked
    "envId"        : "62f10a04-6c54-40c2-a97d-80a98522ff9a",
    "region"       : "NA"
  })
}
```

#### Enhancement Opportunities

**Schema-Based Validation**:
- Use `connector_schema.json` to validate property types
- Warn about missing required properties
- Identify unknown properties
- Provide property descriptions in comments

**Better Nil Handling**:
- Currently: `"customAuth" : "<nil>"` (works but not ideal)
- Enhanced: Omit nil properties or use proper HCL null syntax

**Module Mode Integration** (see section 2.2):
- Generate variables for all configurable properties
- Provide default values from schema
- Document property purposes

#### Files to Modify

- `internal/converter/connector_instance_converter.go` - Add schema validation
- `internal/converter/connector_schema.go` - Expose schema lookup functions
- `internal/converter/connector_properties.go` - Add validation logic (if created)

#### Success Criteria

- [ ] Validate properties against schema
- [ ] Warn about missing required properties
- [ ] Better nil value handling (omit or use HCL null)
- [ ] Add property descriptions as comments
- [ ] Schema-based type checking

#### Note on dvtf-pingctl Limitation

The dvtf-pingctl README states:
> `davinci_connector` `property` blocks are not yet generated

This was a **different limitation** - dvtf-pingctl worked with **exported JSON files** that didn't include property values, requiring users to manually define properties using connector documentation.

**davinci-terraform-converter doesn't have this issue** because it:
1. Fetches from API (includes property values)
2. Exports those properties automatically
3. Masks secrets appropriately

This enhancement is about improving what already works, not implementing missing functionality.

---

### 3.3 Flow Variable Support

**Status**: Not Implemented  
**Priority**: P1 (High - from KNOWN_LIMITATIONS.md)  
**Estimated Effort**: 6-8 hours

#### Description

Convert flow-embedded variables to `pingone_davinci_variable` resources with proper flow association.

**Current Gap**:
Flow variables in JSON exports are ignored. Users must manually create variable resources.

#### Implementation

**Approach**:
1. Parse variables from flow JSON
2. Generate separate `pingone_davinci_variable` resources
3. Link variables to flows via `flow_id` attribute
4. Handle variable references in flow properties

**Example Output**:
```hcl
resource "pingone_davinci_variable" "userid" {
  environment_id = var.environment_id
  flow_id        = pingone_davinci_flow.registration.id
  name           = "userId"
  context        = "flowInstance"
  type           = "string"
  mutable        = true
  description    = "User ID"
}
```

#### Files to Create/Modify

- `internal/converter/flow_variables_extractor.go` - Extract variables from flow JSON
- `internal/converter/variable_converter.go` - Extend for flow variables
- `internal/exporter/exporter.go` - Include flow variables in export

#### Success Criteria

- [ ] Extract variables from flow JSON
- [ ] Generate variable resources with flow association
- [ ] Preserve variable references in flow properties (`{{variableName}}`)
- [ ] Handle sensitive variables correctly
- [ ] Support all variable contexts (flowInstance, company, user)

---

### 3.4 Comparison/Diff Mode

**Status**: Not Implemented  
**Priority**: P3 (Low)  
**Estimated Effort**: 6-10 hours

#### Description

Compare exported HCL with existing Terraform state or files to detect drift.

#### Use Cases

- Drift detection in CI/CD
- Audit changes between environments
- Pre-apply verification

#### Implementation

**CLI Command**:
```bash
# Compare exported HCL with existing state
davinci-convert diff --state terraform.tfstate

# Compare two HCL files
davinci-convert diff --source current.tf --target exported.tf

# Compare two environments
davinci-convert diff \
  --env1 "env-id-1" \
  --env2 "env-id-2"
```

#### Success Criteria

- [ ] Compare HCL files
- [ ] Compare with Terraform state
- [ ] Highlight differences (added/modified/removed resources)
- [ ] Exit code 0 if identical, 1 if differences found

---

## Priority 4: Quality of Life

### 4.1 Interactive Mode

**Status**: Not Implemented  
**Priority**: P3 (Low)  
**Estimated Effort**: 8-12 hours

#### Description

Interactive prompts for required parameters when not provided via flags.

#### Example

```
$ davinci-convert export

PingOne Configuration
=====================
Worker Environment ID: [input or press Enter for $PINGCLI_PINGONE_WORKER_ENVIRONMENT_ID]
Export Environment ID: [input or press Enter to use worker environment]
Region Code: [NA] 
Client ID: [input or press Enter for $PINGCLI_PINGONE_WORKER_CLIENT_ID]
Client Secret: [hidden input]

Export Options
==============
Output format: [1] Single file  [2] Module  [1]
Generate import blocks? [Y/n] Y
Skip dependencies? [y/N] N

Starting export...
```

---

### 4.2 Update Check

**Status**: Not Implemented  
**Priority**: P3 (Low)  
**Estimated Effort**: 2-3 hours

#### Description

Check for tool updates on execution.

```bash
$ davinci-convert export
ℹ A new version is available: v1.2.0 (current: v1.1.0)
  Run: brew upgrade davinci-terraform-converter

Exporting...
```

---

### 4.3 Shell Completions

**Status**: Not Implemented  
**Priority**: P3 (Low)  
**Estimated Effort**: 2-3 hours

#### Description

Generate shell completions for bash, zsh, fish.

```bash
# Generate completions
davinci-convert completion bash > /etc/bash_completion.d/davinci-convert
davinci-convert completion zsh > ~/.zsh/completions/_davinci-convert
davinci-convert completion fish > ~/.config/fish/completions/davinci-convert.fish
```

---

## Implementation Phases

### Phase 1: Production Readiness (Q1 2026)
- [x] API Export (complete)
- [x] Import Blocks (complete)
- [ ] Validation Command (P0)
- [ ] Flow Variables (P1)
- [ ] Multi-File Output (P1)

### Phase 2: Enhanced UX (Q2 2026)
- [ ] Selective Export (P1)
- [ ] Verbose Logging (P2)
- [ ] Dry-Run Mode (P2)
- [ ] Configuration File (P2)

### Phase 3: Advanced Features (Q3 2026)
- [ ] Connector Property Generation (P2)
- [ ] Comparison/Diff Mode (P3)

### Phase 4: Quality of Life (Q4 2026)
- [ ] Interactive Mode (P3)
- [ ] Update Check (P3)
- [ ] Shell Completions (P3)

---

## Success Metrics

### Adoption Metrics
- 80% reduction in manual Terraform setup time
- Zero manual import commands required
- 100% resource type coverage for DaVinci

### Quality Metrics
- CI/CD validation integration (via validate command)
- Successful module generation for large environments (200+ resources)
- Zero manual connector property definitions required

### User Satisfaction
- Easier than dvtf-pingctl for getting started
- Feature parity with dvtf-pingctl validation capabilities
- Production-ready for continuous management

---

## Migration from dvtf-pingctl

**Deprecation Timeline**:
1. **Phase 1 Complete**: Announce deprecation of dvtf-pingctl
2. **Phase 2 Complete**: Provide migration guide
3. **Phase 3 Complete**: Archive dvtf-pingctl repository
4. **Phase 4 Complete**: Final compatibility release, EOL announced

**Migration Guide** (to be created):
- Command mapping (generate → export + module mode)
- Validation workflow changes
- Configuration file migration
- Module structure differences
