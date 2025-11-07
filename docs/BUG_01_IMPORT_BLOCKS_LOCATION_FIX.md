# Bug 01: Import Blocks Location Fix

## Problem Statement

Import blocks are generated within the child module alongside resources instead of in the root module's `imports.tf` file. This violates Terraform module best practices where import blocks should exist at the consumption layer (root module), not within the reusable module itself.

### Current Behavior

```text
output/
├── davinci-module/           # Child module
│   ├── flows.tf             # Has import blocks mixed with resources ❌
│   ├── variables.tf         # Has import blocks mixed with resources ❌
│   ├── connectors.tf        # Has import blocks mixed with resources ❌
│   └── versions.tf
└── module.tf                # Root module
```

### Expected Behavior

```text
output/
├── davinci-module/           # Child module
│   ├── flows.tf             # Resources only ✓
│   ├── variables.tf         # Resources only ✓
│   ├── connectors.tf        # Resources only ✓
│   └── versions.tf
├── imports.tf               # Import blocks in root module ✓
└── module.tf                # Root module
```

## Root Cause Analysis

Import blocks are generated in the exporter functions (`ExportVariablesForModule`, `ExportConnectorInstancesForModule`, etc.) and appended directly to `hclBlocks` alongside resource definitions:

```go
// In variable_exporter.go line 97
if importGen != nil {
    importBlock, err := importGen.GenerateImportBlock(...)
    hclBlocks = append(hclBlocks, importBlock)  // ❌ Mixed with resources
}
// ...
hclBlocks = append(hclBlocks, hcl)  // Resource HCL
```

These blocks are then returned as a single concatenated string and written to child module files. There is no mechanism to separate and track import blocks independently.

In `ConvertExportedDataToModuleStructure` (line 208):

```go
// TODO: Import blocks support - will need to parse from HCL or track separately
```

The `ModuleStructure.ImportBlocks` field exists but is never populated.

## Solution Design

### Phase 1: Update Data Structures

Add import block tracking to `ExportedData`:

```go
// internal/exporter/module_export.go
type ExportedData struct {
    // ... existing fields ...
    
    // Import blocks for root module (separate from resource HCL)
    ImportBlocks []RawImportBlock
}

// RawImportBlock represents an import block before module path transformation
type RawImportBlock struct {
    ResourceType string // "pingone_davinci_variable"
    ResourceName string // "company_name"
    ResourceID   string // The import ID (e.g., "env-id/var-id")
}
```

### Phase 2: Update Exporter Functions

Modify each `*ForModule` function to return import blocks separately:

```go
// Before (variable_exporter.go)
func ExportVariablesForModule(...) (string, []converter.VariableEligibleAttribute, map[string][]byte, map[string]string, error)

// After
func ExportVariablesForModule(...) (string, []converter.VariableEligibleAttribute, map[string][]byte, map[string]string, []RawImportBlock, error)
```

Separate import block generation from resource HCL:

```go
var hclBlocks []string
var importBlocks []RawImportBlock

for _, variable := range variables {
    // Track import block separately
    if importGen != nil {
        importBlocks = append(importBlocks, RawImportBlock{
            ResourceType: "pingone_davinci_variable",
            ResourceName: actualName,
            ResourceID:   buildImportID("pingone_davinci_variable", client.EnvironmentID, variableID),
        })
    }
    
    // Generate resource HCL only
    hcl, err := converter.ConvertVariableWithOptions(variableJSON, skipDeps)
    hclBlocks = append(hclBlocks, hcl)
}
```

### Phase 3: Populate ExportedData.ImportBlocks

In `ExportEnvironmentForModule`, collect import blocks from all exporters:

```go
variablesHCL, variablesExtracted, variablesJSON, variableNames, variableImports, err := ExportVariablesForModule(...)
data.ImportBlocks = append(data.ImportBlocks, variableImports...)

connectorsHCL, connectorsExtracted, connectorsJSON, connectorNames, connectorImports, err := ExportConnectorInstancesForModule(...)
data.ImportBlocks = append(data.ImportBlocks, connectorImports...)

// ... repeat for all resource types
```

### Phase 4: Transform Import Blocks for Module References

In `ConvertExportedDataToModuleStructure`, convert raw import blocks to module-scoped references:

```go
func ConvertExportedDataToModuleStructure(data *ExportedData, config module.ModuleConfig) (*module.ModuleStructure, error) {
    // ... existing code ...
    
    // Transform import blocks to reference module resources
    importBlocks := make([]module.ImportBlock, 0, len(data.ImportBlocks))
    for _, raw := range data.ImportBlocks {
        importBlocks = append(importBlocks, module.ImportBlock{
            To: fmt.Sprintf("module.%s.%s.%s", config.ModuleDirName, raw.ResourceType, raw.ResourceName),
            ID: raw.ResourceID,
        })
    }
    
    structure.ImportBlocks = importBlocks
    return structure, nil
}
```

### Phase 5: Update Generator

The `generator.generateImportsTF` function already writes import blocks to root module correctly (line 316). No changes needed here.

## Testing Strategy

### Unit Tests

1. **Test ExportedData tracking** - Verify import blocks are stored separately
2. **Test exporter separation** - Verify exporters return import blocks in separate slice
3. **Test module reference transformation** - Verify import blocks reference `module.{name}.{type}.{resource}`
4. **Test import ID format** - Verify import IDs match provider requirements

### Integration Tests

1. Export environment with `--include-imports --module` flags
2. Verify `imports.tf` exists in root module
3. Verify child module files have no import blocks
4. Verify import blocks reference correct module path
5. Run `terraform plan` and verify import count matches resource count

## Implementation Order

1. ✅ **Analysis** - Understand current flow and identify issue
2. 🔄 **Design** - Document approach (this file)
3. ⏳ **Phase 1** - Update data structures
4. ⏳ **Phase 2** - Modify exporters to separate import blocks
5. ⏳ **Phase 3** - Update module structure conversion
6. ⏳ **Phase 4** - Write comprehensive tests
7. ⏳ **Phase 5** - Integration testing with real exports

## Files to Modify

- `internal/exporter/module_export.go` - Add ImportBlocks tracking
- `internal/exporter/variable_exporter.go` - Separate import block generation
- `internal/exporter/connector_exporter.go` - Separate import block generation
- `internal/exporter/flow_exporter.go` - Separate import block generation
- `internal/exporter/application_exporter.go` - Separate import block generation
- `internal/exporter/flow_policy_exporter.go` - Separate import block generation

## Success Criteria

- [ ] Import blocks appear in root module `imports.tf` only
- [ ] Child module resources have no import blocks
- [ ] Import blocks reference `module.{module_name}.{resource_type}.{resource_name}`
- [ ] Import IDs are in correct format per resource type
- [ ] `terraform plan` shows correct import count
- [ ] All existing tests pass
- [ ] New tests validate import block location and format
