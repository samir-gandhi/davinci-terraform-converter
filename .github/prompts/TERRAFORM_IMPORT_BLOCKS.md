# Terraform Import Blocks Feature

## Objective

Generate Terraform `import` blocks alongside resource definitions to enable immediate adoption of existing DaVinci environments into Terraform management without manual import commands.

## Problem Statement

Current export workflow requires manual import:

```bash
# 1. Export resources
davinci-convert export --out environment.tf

# 2. Manually import each resource (tedious, error-prone)
terraform import pingone_davinci_variable.companyname <env-id>/<var-id>
terraform import pingone_davinci_connector_instance.httpconnector_abc123 <env-id>/<instance-id>
terraform import pingone_davinci_flow.signin_flow <env-id>/<flow-id>
# ... repeat for 62+ resources
```

**Issues:**
- Manual process for large environments (50+ resources)
- Resource ID mapping errors
- Time-consuming (5-10 minutes per environment)
- Blocks automation and CI/CD pipelines

## Proposed Solution

Generate `import` blocks in exported HCL (Terraform 1.5+ feature):

```hcl
# Import block (Terraform handles the import automatically)
import {
  to = pingone_davinci_variable.companyname
  id = "62f10a04-6c54-40c2-a97d-80a98522ff9a/abc123-def456-..."
}

# Resource definition
resource "pingone_davinci_variable" "companyname" {
  environment_id = var.environment_id
  name           = "companyName"
  description    = "Company branding variable"
  type           = "string"
  value          = "Ping Identity"
}
```

**Workflow becomes:**
```bash
# 1. Export with import blocks
davinci-convert export --out environment.tf

# 2. Apply (Terraform imports automatically)
terraform init
terraform plan   # Shows: Plan: 62 to import, 0 to add, 0 to change, 0 to destroy
terraform apply  # Imports all resources into state
```

## Terraform Import Block Requirements

### Import ID Format by Resource Type

Based on PingOne Terraform Provider documentation:

**Variables:**
```hcl
import {
  to = pingone_davinci_variable.example
  id = "<environment_id>/<variable_id>"
}
```

**Connector Instances:**
```hcl
import {
  to = pingone_davinci_connector_instance.example
  id = "<environment_id>/<connector_instance_id>"
}
```

**Flows:**
```hcl
import {
  to = pingone_davinci_flow.example
  id = "<environment_id>/<flow_id>"
}
```

**Applications:**
```hcl
import {
  to = pingone_davinci_application.example
  id = "<environment_id>/<application_id>"
}
```

**Flow Policies:**
```hcl
import {
  to = pingone_davinci_flow_policy.example
  id = "<environment_id>/<flow_policy_id>"
}
```

**Flow Policy Assignments:**
```hcl
import {
  to = pingone_davinci_application_flow_policy_assignment.example
  id = "<environment_id>/<application_id>/<flow_policy_id>"
}
```

## Implementation Design

### Architecture Changes

**1. Converter Interface Extension**

Add import block generation to converter interface:

```go
// internal/converter/converter.go
type ResourceConverter interface {
    Convert(resource interface{}) (string, error)
    ConvertWithOptions(resource interface{}, opts ConvertOptions) (string, error)
    
    // NEW: Generate import block
    GenerateImportBlock(resource interface{}, opts ConvertOptions) (string, error)
}

type ConvertOptions struct {
    SkipDependencies bool
    GenerateImports  bool  // NEW: Enable import block generation
}
```

**2. Import Block Generator**

New package for import block generation:

```go
// internal/importgen/import_generator.go
package importgen

import (
    "fmt"
)

// ImportBlockGenerator generates Terraform import blocks
type ImportBlockGenerator struct{}

// GenerateImportBlock creates an import block for a resource
func (g *ImportBlockGenerator) GenerateImportBlock(
    resourceType string,
    resourceName string,
    resourceID string,
    environmentID string,
) string {
    importID := buildImportID(resourceType, environmentID, resourceID)
    
    return fmt.Sprintf(`import {
  to = %s.%s
  id = "%s"
}`, resourceType, resourceName, importID)
}

// buildImportID constructs the import ID based on resource type
func buildImportID(resourceType, environmentID, resourceID string) string {
    switch resourceType {
    case "pingone_davinci_variable":
        return fmt.Sprintf("%s/%s", environmentID, resourceID)
    case "pingone_davinci_connector_instance":
        return fmt.Sprintf("%s/%s", environmentID, resourceID)
    case "pingone_davinci_flow":
        return fmt.Sprintf("%s/%s", environmentID, resourceID)
    case "pingone_davinci_application":
        return fmt.Sprintf("%s/%s", environmentID, resourceID)
    case "pingone_davinci_flow_policy":
        return fmt.Sprintf("%s/%s", environmentID, resourceID)
    case "pingone_davinci_application_flow_policy_assignment":
        // Special case: three-part ID
        // Need to extract application_id from resource
        return "" // Handled separately
    default:
        return fmt.Sprintf("%s/%s", environmentID, resourceID)
    }
}
```

**3. Orchestrator Integration**

Update orchestrator to generate import blocks:

```go
// internal/exporter/orchestrator.go

func (o *Orchestrator) ExportEnvironment(ctx context.Context, opts ExportOptions) (*ExportResult, error) {
    // ... existing export logic ...
    
    var hclOutput strings.Builder
    
    for _, resource := range orderedResources {
        // Generate import block if enabled
        if opts.GenerateImports {
            importBlock := o.importGenerator.GenerateImportBlock(
                resource.Type,
                resource.Name,
                resource.ID,
                o.environmentID,
            )
            hclOutput.WriteString(importBlock)
            hclOutput.WriteString("\n\n")
        }
        
        // Generate resource definition
        hcl := converter.Convert(resource)
        hclOutput.WriteString(hcl)
        hclOutput.WriteString("\n\n")
    }
    
    return &ExportResult{HCL: hclOutput.String()}, nil
}
```

**4. CLI Flag**

Add flag to export command:

```go
// cmd/export.go

func init() {
    // ... existing flags ...
    
    exportCmd.Flags().Bool("generate-imports", false, "Generate Terraform import blocks")
}
```

### Output Format

**Without `--generate-imports` (current behavior):**
```hcl
resource "pingone_davinci_variable" "companyname" {
  environment_id = var.environment_id
  name           = "companyName"
  description    = "Company branding variable"
  type           = "string"
  value          = "Ping Identity"
}

resource "pingone_davinci_flow" "signin_flow" {
  environment_id = var.environment_id
  name           = "Sign-In Flow"
  # ...
}
```

**With `--generate-imports`:**
```hcl
import {
  to = pingone_davinci_variable.companyname
  id = "62f10a04-6c54-40c2-a97d-80a98522ff9a/abc123-def456-..."
}

resource "pingone_davinci_variable" "companyname" {
  environment_id = var.environment_id
  name           = "companyName"
  description    = "Company branding variable"
  type           = "string"
  value          = "Ping Identity"
}

import {
  to = pingone_davinci_flow.signin_flow
  id = "62f10a04-6c54-40c2-a97d-80a98522ff9a/flow-abc123-..."
}

resource "pingone_davinci_flow" "signin_flow" {
  environment_id = var.environment_id
  name           = "Sign-In Flow"
  # ...
}
```

## Skip-Dependencies Interaction

Import blocks work with both dependency modes:

**With dependencies (default):**
```hcl
import {
  to = pingone_davinci_flow.signin_flow
  id = "62f10a04-6c54-40c2-a97d-80a98522ff9a/flow-abc123-..."
}

resource "pingone_davinci_flow" "signin_flow" {
  environment_id = var.environment_id  # Variable reference
  
  connection_link {
    id = pingone_davinci_connector_instance.httpconnector.id  # Terraform reference
  }
}
```

**With `--skip-dependencies`:**
```hcl
import {
  to = pingone_davinci_flow.signin_flow
  id = "62f10a04-6c54-40c2-a97d-80a98522ff9a/flow-abc123-..."
}

resource "pingone_davinci_flow" "signin_flow" {
  environment_id = "62f10a04-6c54-40c2-a97d-80a98522ff9a"  # Hardcoded
  
  connection_link {
    id = "conn-abc123-def456-..."  # Hardcoded UUID
  }
}
```

**Both modes support import blocks.** The `--skip-dependencies` flag only affects resource attribute values, not import block generation.

## Usage Examples

### Basic Usage

```bash
# Export with import blocks
davinci-convert export \
  --generate-imports \
  --out environment.tf

# Apply
cd terraform-workspace/
cp ../environment.tf .
terraform init
terraform plan   # Shows resources to import
terraform apply  # Imports all resources
```

### Combined with Skip-Dependencies

```bash
# Export standalone resources with import blocks
davinci-convert export \
  --generate-imports \
  --skip-dependencies \
  --out standalone.tf

# Useful for:
# - Testing individual resources
# - Migrating specific resources between environments
# - Quick prototyping
```

### Selective Import (Future Enhancement)

```bash
# Generate imports only for specific resource types
davinci-convert export \
  --generate-imports \
  --import-types "flows,applications" \
  --out partial-import.tf

# Use case: Import flows/apps but manually create connector instances
```

## Testing Strategy

### Unit Tests

**Import Block Generation:**
```go
// internal/importgen/import_generator_test.go

func TestGenerateImportBlock_Variable(t *testing.T) {
    gen := &ImportBlockGenerator{}
    
    result := gen.GenerateImportBlock(
        "pingone_davinci_variable",
        "companyname",
        "var-abc123",
        "env-def456",
    )
    
    expected := `import {
  to = pingone_davinci_variable.companyname
  id = "env-def456/var-abc123"
}`
    
    assert.Equal(t, expected, result)
}

func TestGenerateImportBlock_FlowPolicyAssignment(t *testing.T) {
    gen := &ImportBlockGenerator{}
    
    // Test three-part ID format
    result := gen.GenerateImportBlock(
        "pingone_davinci_application_flow_policy_assignment",
        "web_app_policy",
        "policy-abc123",
        "env-def456",
        WithApplicationID("app-ghi789"),
    )
    
    expected := `import {
  to = pingone_davinci_application_flow_policy_assignment.web_app_policy
  id = "env-def456/app-ghi789/policy-abc123"
}`
    
    assert.Equal(t, expected, result)
}
```

**Orchestrator Integration:**
```go
// internal/exporter/orchestrator_test.go

func TestExportEnvironment_WithImportBlocks(t *testing.T) {
    orchestrator := setupTestOrchestrator()
    
    result, err := orchestrator.ExportEnvironment(context.Background(), ExportOptions{
        GenerateImports: true,
    })
    
    require.NoError(t, err)
    
    // Verify import blocks present
    assert.Contains(t, result.HCL, "import {")
    assert.Contains(t, result.HCL, "to = pingone_davinci_variable")
    assert.Contains(t, result.HCL, "id = \"")
    
    // Verify import block before resource definition
    importIdx := strings.Index(result.HCL, "import {")
    resourceIdx := strings.Index(result.HCL, "resource \"pingone_davinci_variable\"")
    assert.Less(t, importIdx, resourceIdx, "Import block should appear before resource")
}
```

### Integration Tests

**End-to-End Import Test:**
```go
// tests/acceptance/import_test.go

func TestImportBlocks_FullEnvironmentImport(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // 1. Export with import blocks
    cmd := exec.Command("davinci-convert", "export", "--generate-imports", "--out", "/tmp/import-test.tf")
    err := cmd.Run()
    require.NoError(t, err)
    
    // 2. Set up Terraform workspace
    setupTerraformWorkspace(t, "/tmp/import-test.tf")
    
    // 3. Run terraform plan
    planOutput := runTerraformPlan(t)
    
    // 4. Verify plan shows imports
    assert.Contains(t, planOutput, "Plan: ")
    assert.Contains(t, planOutput, "to import")
    assert.NotContains(t, planOutput, "to add", "Should import, not create new")
    
    // 5. Run terraform apply
    applyOutput := runTerraformApply(t)
    
    // 6. Verify successful import
    assert.Contains(t, applyOutput, "Import complete")
    
    // 7. Verify state
    state := getTerraformState(t)
    assert.Greater(t, len(state.Resources), 0)
}
```

### Manual Testing Checklist

- [ ] Export environment with `--generate-imports`
- [ ] Verify import blocks present for all resource types
- [ ] Verify import ID format matches provider documentation
- [ ] Run `terraform plan` - should show "X to import, 0 to add"
- [ ] Run `terraform apply` - should successfully import all resources
- [ ] Verify state file contains all resources
- [ ] Run `terraform plan` again - should show "No changes"
- [ ] Test with `--skip-dependencies` flag
- [ ] Test with large environment (50+ resources)
- [ ] Test error handling for invalid import IDs

## Edge Cases and Validation

### Environment ID Handling

**Issue:** Import blocks need actual environment ID, but resource definitions may use `var.environment_id`.

**Solution:**
- Import blocks always use actual environment ID (from API)
- Resource definitions follow `--skip-dependencies` flag
- Export must have access to actual environment ID

```go
func (o *Orchestrator) ExportEnvironment(ctx context.Context, opts ExportOptions) (*ExportResult, error) {
    // Always store actual environment ID for import blocks
    actualEnvID := o.client.GetEnvironmentID()
    
    for _, resource := range resources {
        if opts.GenerateImports {
            // Import block uses actual ID
            importBlock := generateImport(resource.Type, resource.Name, resource.ID, actualEnvID)
        }
        
        // Resource definition respects skip-dependencies flag
        envIDValue := actualEnvID
        if !opts.SkipDependencies {
            envIDValue = "var.environment_id"
        }
        hcl := converter.Convert(resource, envIDValue)
    }
}
```

### Flow Policy Assignment Special Case

Flow policy assignments have three-part import ID:

```hcl
import {
  to = pingone_davinci_application_flow_policy_assignment.web_app_policy
  id = "<environment_id>/<application_id>/<flow_policy_id>"
}
```

**Implementation:**
```go
type FlowPolicyAssignment struct {
    EnvironmentID string
    ApplicationID string
    FlowPolicyID  string
    Priority      int
}

func (g *ImportBlockGenerator) GenerateFlowPolicyAssignmentImport(
    assignment *FlowPolicyAssignment,
    resourceName string,
) string {
    importID := fmt.Sprintf("%s/%s/%s",
        assignment.EnvironmentID,
        assignment.ApplicationID,
        assignment.FlowPolicyID,
    )
    
    return fmt.Sprintf(`import {
  to = pingone_davinci_application_flow_policy_assignment.%s
  id = "%s"
}`, resourceName, importID)
}
```

### Resource Name Conflicts

Import blocks must reference exact resource names:

**Issue:** Resource name sanitization must match between import block and resource definition.

**Solution:** Use same name generator for both:

```go
func (o *Orchestrator) ExportResource(resource interface{}) string {
    // Generate consistent resource name
    resourceName := utils.SanitizeResourceName(resource.GetName())
    
    // Use same name for import block and resource
    importBlock := generateImport(resource.Type, resourceName, resource.ID, o.envID)
    resourceDef := generateResource(resource.Type, resourceName, resource)
    
    return importBlock + "\n\n" + resourceDef
}
```

## Terraform Version Compatibility

**Requirements:**
- Terraform >= 1.5.0 (import blocks introduced)
- PingOne Terraform Provider >= 0.28.0

**Validation:**
```hcl
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = ">= 0.28.0"
    }
  }
}
```

**Error Handling:**
```go
func ValidateTerraformVersion() error {
    cmd := exec.Command("terraform", "version", "-json")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to check Terraform version: %w", err)
    }
    
    var versionInfo struct {
        Version string `json:"terraform_version"`
    }
    
    if err := json.Unmarshal(output, &versionInfo); err != nil {
        return err
    }
    
    if !semver.IsGreaterOrEqual(versionInfo.Version, "1.5.0") {
        return fmt.Errorf("Terraform 1.5.0+ required for import blocks, found %s", versionInfo.Version)
    }
    
    return nil
}
```

## Documentation Updates

### README.md

Add section:

```markdown
### Import Blocks (Terraform 1.5+)

Generate import blocks to automatically import existing resources:

```bash
davinci-convert export --generate-imports --out environment.tf
cd terraform-workspace/
terraform init
terraform plan   # Shows: Plan: 62 to import
terraform apply  # Imports all resources
```

**Benefits:**
- No manual import commands
- Automated environment adoption
- Safe state migration
```

### ARCHITECTURE.md

Add section:

```markdown
## Import Block Generation

The tool can generate Terraform import blocks (Terraform 1.5+) alongside resource definitions:

```
Export Command
     ↓
Fetch Resources from API
     ↓
For Each Resource:
├─→ Generate Import Block (if --generate-imports)
│   └─→ import { to = ... id = "..." }
└─→ Generate Resource Definition
    └─→ resource "..." "..." { ... }
```

Import ID format varies by resource type according to provider documentation.
```

### Examples

Create example script:

```bash
#!/usr/bin/env bash
# examples/05-import-blocks-usage.sh

# Export with import blocks
davinci-convert export \
  --generate-imports \
  --out environment-with-imports.tf

# Set up Terraform workspace
mkdir -p terraform-import-test
cd terraform-import-test
cp ../environment-with-imports.tf .
cp ../examples/terraform/provider.tf .
cp ../examples/terraform/variables.tf .
cp ../examples/terraform/terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars with your values
# vim terraform.tfvars

# Initialize and import
terraform init
terraform plan   # Review import plan
terraform apply  # Import all resources

# Verify
terraform state list
terraform plan   # Should show "No changes"
```

## Implementation Phases

### Phase 1: Core Import Block Generation
- [ ] Create `internal/importgen` package
- [ ] Implement `ImportBlockGenerator` interface
- [ ] Add import ID builders for each resource type
- [ ] Unit tests for import block generation

### Phase 2: Orchestrator Integration
- [ ] Add `GenerateImports` to `ExportOptions`
- [ ] Update orchestrator to generate import blocks
- [ ] Ensure import blocks appear before resource definitions
- [ ] Handle environment ID correctly (actual vs variable)

### Phase 3: CLI Integration
- [ ] Add `--generate-imports` flag to export command
- [ ] Update help text and documentation
- [ ] Add validation for Terraform version (warn if < 1.5.0)

### Phase 4: Testing
- [ ] Unit tests for all resource types
- [ ] Integration tests with real Terraform
- [ ] End-to-end import workflow test
- [ ] Edge case testing (policy assignments, special IDs)

### Phase 5: Documentation
- [ ] Update README.md
- [ ] Update ARCHITECTURE.md
- [ ] Create example scripts
- [ ] Add troubleshooting guide

### Phase 6: Polish
- [ ] Performance optimization (parallel generation)
- [ ] Error handling improvements
- [ ] Logging and progress reporting
- [ ] User feedback for import success

## Benefits

**User Experience:**
- Single command to adopt existing environments
- No manual import commands (saves 5-10 minutes per environment)
- Reduced error potential (no manual ID mapping)
- CI/CD pipeline friendly

**Operational:**
- Faster environment migrations
- Simplified disaster recovery (export → import)
- Easier multi-environment management
- Support for large-scale deployments

**Strategic:**
- Lowers barrier to Terraform adoption
- Enables "try before you buy" workflows
- Supports Infrastructure as Code best practices
- Competitive feature (not common in conversion tools)

## Future Enhancements

### Selective Import
```bash
# Import only specific resource types
davinci-convert export \
  --generate-imports \
  --import-types "flows,applications"
```

### Import Validation
```bash
# Validate import IDs before generating
davinci-convert export \
  --generate-imports \
  --validate-imports
```

### Import Report
```bash
# Generate report of what will be imported
davinci-convert export \
  --generate-imports \
  --dry-run \
  --out import-report.txt
```

### Terraform Cloud Integration
```bash
# Generate imports compatible with Terraform Cloud
davinci-convert export \
  --generate-imports \
  --terraform-cloud
```

## References

- [Terraform Import Blocks](https://developer.hashicorp.com/terraform/language/import)
- [PingOne Provider Import Documentation](https://registry.terraform.io/providers/pingidentity/pingone/latest/docs)
- [Terraform 1.5 Release Notes](https://github.com/hashicorp/terraform/releases/tag/v1.5.0)
