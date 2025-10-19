---
mode: agent
---

# Part 4: Dependency Resolution and Terraform References

**Status**: 🟡 IN PROGRESS

**Prerequisites**:
- ✅ Part 3 (full export) complete
- ✅ All resource types can be exported
- ✅ HAL link parsing functional

**Goal**: Replace hardcoded resource IDs with Terraform references to enable proper resource dependencies.

---

## Current Implementation Status

### Phase 4.1: Build Dependency Graph - ✅ COMPLETE

**Architecture**: Four-tier system
1. **Schema (HARDCODED)** - Defines WHERE dependencies exist in resource JSON
2. **Parser (DYNAMIC)** - Extracts dependency IDs using schema paths
3. **Graph (DYNAMIC)** - Stores discovered dependency relationships
4. **Hierarchy (DYNAMIC)** - Tracks parent-child from HAL links (separate from dependencies)

**Implemented Files**:
- ✅ `internal/resolver/resolver.go` - Core dependency graph (115 lines)
  - ResourceRef, Dependency, DependencyGraph structures
  - AddResource(), AddDependency(), GetDependencies(), GetReferenceName()
  - 9/9 tests passing
  
- ✅ `internal/resolver/schema.go` - Hardcoded dependency schemas (129 lines)
  - FieldPath structure: Path (JSON path), TargetType, FieldName, IsArray, IsOptional
  - GetFlowDependencySchema() - 3 paths: nodes[*].data.connectionId, nodes[*].data.properties.variableId, nodes[*].data.properties.subFlowId
  - GetFlowPolicyDependencySchema() - 2 paths: flowDistributions[*].id, applicationId
  - GetApplicationDependencySchema() - empty (no embedded dependencies)
  - GetConnectorInstanceDependencySchema() - empty
  - GetVariableDependencySchema() - empty
  - 9/9 schema tests passing
  
- ✅ `internal/resolver/hierarchy.go` - Parent-child relationship tracking (79 lines)
  - HierarchyRelationship: ParentType, ParentID, ChildType, Children[]
  - ResourceHierarchyGraph.AddRelationship(parentType, parentID, childType, childIDs[])
  - ResourceHierarchyGraph.GetChildren(parentType, parentID)
  - Separate from field-level dependencies - tracks HAL link ownership
  - 8/8 hierarchy tests passing

- ✅ `internal/resolver/parser.go` - Schema-driven dependency extraction (200 lines)
  - ParseResourceDependencies(type, id, data, schema) - Main parsing function
  - extractValuesAtPath(data, path) - JSON path navigation with array wildcard support
  - traversePath(values, pathParts) - Recursive traversal handling items[*] notation
  - navigateToField(data, field) - Field extraction from map[string]interface{}
  - splitPath(path) - Path parsing preserving bracket notation
  - FindReferencesInFlow/FlowPolicy/Application/ConnectorInstance/Variable() - Type-specific wrappers
  - 17/17 parser tests passing

- ✅ `internal/resolver/resolver_manager.go` - Parent constructor orchestrating components (350 lines)
  - ResolverManager coordinates: schema + parser + graph + hierarchy
  - ProcessResource(type, id, data, links) - Main entry point
    * Uses schema to get dependency paths
    * Calls parser to extract IDs
    * Stores in dependency graph
    * Parses HAL links for hierarchy
  - GenerateOutput() - Produces ResourceWithDependencies output
  - Preserves original data for re-parsing
  - Tracks unresolved dependencies separately

- ✅ `internal/resolver/README.md` - Architecture documentation (comprehensive developer guide)
- ✅ `internal/resolver/COMPARISON_TO_TERRAFORMER.md` - Analysis proving our approach superior for DaVinci

**Test Coverage**:
- ✅ 53/53 resolver package tests passing
  - 9 resolver tests (graph operations)
  - 17 parser tests (path traversal, dependency extraction)
  - 9 schema tests (schema definitions, lookups)
  - 8 hierarchy tests (relationship tracking)
  - 6 naming tests (sanitization, uniqueness)
  - 4 integration tests (complete workflow validation)
- ✅ All internal tests passing

**Integration Test Coverage**:
- ✅ Complete workflow: flowData → schema → parser → dependencies → Terraform references
- ✅ Flow policy dependency resolution with applications and flows
- ✅ Real-world JSON parsing from DaVinci API format
- ✅ Name uniqueness enforcement across multiple resources
- ✅ Missing dependency error handling and TODO generation

**Key Architecture Decisions**:
- **Schema is HARDCODED source of truth** - Defines field paths like "graphData.elements.nodes[*].data.connectionId"
- **Parser is DYNAMIC** - Uses schema paths to navigate runtime JSON and extract dependency IDs
- **Graph is DYNAMIC** - Stores discovered dependencies at runtime
- **Hierarchy separate from dependencies** - HAL links show ownership (app owns policies), field parsing shows references (policy references flow)
- **Array wildcard support** - Parser handles paths like "items[*].id" to extract from all array elements
- **Optional vs Required** - Schema marks fields as IsOptional to handle missing dependencies gracefully
- **ResolverManager delegates** - Reduced from 420 to 350 lines by delegating path traversal to parser module
- **Data preservation** - Original JSON preserved in output for re-parsing if needed

### Phase 4.2: Generate Terraform References - ✅ COMPLETE

**Goal**: Replace extracted dependency IDs with Terraform reference syntax

**Implemented**:
- ✅ `internal/resolver/naming.go` - Name sanitization (70 lines)
  - SanitizeName() - Converts human-readable names to valid Terraform identifiers
  - Handles: lowercase, special chars, spaces → underscores, uniqueness tracking
  - toSnakeCase() - Converts camelCase/PascalCase for connector IDs
  - 3/3 naming tests passing
  
- ✅ `internal/resolver/reference.go` - Terraform reference generation (40 lines)
  - GenerateTerraformReference(graph, type, id, attribute) → "pingone_davinci_flow.my_flow.id"
  - GenerateTODOPlaceholder(type, id, error) → Comment for missing dependencies
  - mapToTerraformResourceType() - Internal type → Terraform provider resource type
  - 3/3 reference tests passing

**Functionality**:
- GetReferenceName() already implemented in resolver.go
- Resource name uniqueness enforced via nameUsage map in DependencyGraph
- Reference format: `resource_type.resource_name.attribute`
- TODO placeholders include resource type, ID, and error context

**Test Coverage**:
- ✅ 49/49 resolver package tests passing (43 + 6 new)

### Phase 4.3: Handle Missing Dependencies - ⏳ NOT STARTED

**Goal**: Generate informative placeholders when dependencies not found

**Planned Implementation**:
- Track missing dependency reasons (excluded, not included, not found)
- Generate TODO comments with original IDs and names
- Summary report after export

### Phase 4.4: Validate Dependency Graph - ⏳ NOT STARTED

**Goal**: Detect circular dependencies and generate optimal ordering

**Planned Implementation**:
- Cycle detection using DFS
- Topological sort for resource ordering
- Error reporting for unresolvable cycles

---

## CRITICAL: Naming Consistency

**⚠️ IMPORTANT**: The current flow converter generates dependency references based on **expected naming patterns**:
- Connection references: `pingone_davinci_connector_instance.{connectorId}_{connectionId}.id`
- Format uses: `toSnakeCase(connectorId)` + `_` + `connectionId`

**When implementing Part 4**:
1. **Verify naming alignment**: Ensure connection resource names generated in Part 3 (connector instance export) match the expected format used in flow references
2. **Use consistent sanitization**: Both flow converter and connection exporter must use identical naming logic
3. **Test name matching**: Write integration tests that verify flow references resolve to actual generated connection resources
4. **Document naming contract**: Clearly specify the naming pattern that both systems must follow

**Current implementation** (flow_converter.go lines ~603-605):
```go
// Format: pingone_davinci_connector_instance.<connector_id>_<connection_id>.id
connectorName := toSnakeCase(connectorID)
return fmt.Sprintf("pingone_davinci_connector_instance.%s_%s.id", connectorName, connectionID)
```

This must match the resource names generated when exporting connector instances in Part 3.

---

## Overview

Instead of hardcoded IDs in generated HCL:
```hcl
# BAD - Hardcoded IDs
connection_id = "abc123-def456-ghi789"
variable_id   = "xyz789-abc123-def456"
```

Generate Terraform references:
```hcl
# GOOD - Terraform references
connection_id = pingone_davinci_connector_instance.httpconnector_conn-123.id
variable_id   = pingone_davinci_variable.api_key.id
```

This allows Terraform to understand resource dependencies and apply them in correct order.

---

## Phase 4.1: Build Dependency Graph

**STATUS**: ✅ COMPLETE

**Implementation approach**:

Our implementation uses a **schema-driven architecture** where dependency locations are defined once and applied dynamically to runtime data. This approach is superior to Terraformer's hardcoded connection mappings (see COMPARISON_TO_TERRAFORMER.md).

### Architecture Overview

**Four-tier system**:

1. **Schema (HARDCODED)** - JSON path definitions pointing to dependency fields
2. **Parser (DYNAMIC)** - Extracts IDs from JSON using schema paths
3. **Graph (DYNAMIC)** - Stores discovered dependencies
4. **Hierarchy (DYNAMIC)** - Tracks HAL link parent-child relationships

**Flow**: Schema → Parser → Graph → Output

### Implemented Components

**Core Graph** (`internal/resolver/resolver.go` - 115 lines):

```go
type ResourceRef struct {
    Type string  // "flow", "connector_instance", "variable"
    ID   string  // Original resource ID
}

type Dependency struct {
    From     ResourceRef
    To       ResourceRef
    Field    string  // Terraform field name
}

type DependencyGraph struct {
    resources    map[string]ResourceRef
    dependencies []Dependency
}
```

**Schema Definitions** (`internal/resolver/schema.go` - 129 lines):

```go
type FieldPath struct {
    Path         string  // "graphData.elements.nodes[*].data.connectionId"
    TargetType   string  // "connector_instance"
    FieldName    string  // "connection_id"
    IsArray      bool    // true for arrays
    IsOptional   bool    // true for optional fields
}

// Schemas for each resource type
GetFlowDependencySchema() // 3 paths: connectionId, variableId, subFlowId
GetFlowPolicyDependencySchema() // 2 paths: flowDistributions[*].id, applicationId
GetApplicationDependencySchema() // empty
GetConnectorInstanceDependencySchema() // empty
GetVariableDependencySchema() // empty
```

**Parser** (`internal/resolver/parser.go` - 200 lines):

```go
// Main parsing function
ParseResourceDependencies(type, id, data, schema) []Dependency

// Path traversal with array wildcard support
extractValuesAtPath(data, "nodes[*].data.connectionId")
// Returns: ["conn-1", "conn-2", "conn-3"]

// Path parsing helpers
splitPath("graphData.elements.nodes[*].data.connectionId")
traversePath(values, pathParts)
navigateToField(data, field)
```

**Hierarchy Tracking** (`internal/resolver/hierarchy.go` - 79 lines):

```go
type HierarchyRelationship struct {
    ParentType string   // "application"
    ParentID   string   // "app-123"
    ChildType  string   // "flow_policy"
    Children   []string // ["policy-1", "policy-2"]
}

// Separate from field-level dependencies
// Tracks HAL link ownership: app owns policies, policy owns flows
```

**Orchestration** (`internal/resolver/resolver_manager.go` - 350 lines):

```go
type ResolverManager struct {
    graph     *DependencyGraph
    hierarchy *ResourceHierarchyGraph
    // Stores original data for re-parsing
}

// Main entry point
ProcessResource(type, id, data, links)
// 1. Get schema for resource type
// 2. Call parser to extract dependencies
// 3. Store in graph
// 4. Parse HAL links for hierarchy

// Output generation
GenerateOutput() ResourceWithDependencies
```

### Test Coverage

**43/43 tests passing**:

- 9 resolver tests (graph operations)
- 17 parser tests (path traversal, dependency extraction)
- 9 schema tests (schema definitions, lookups)
- 8 hierarchy tests (relationship tracking)

### Key Design Decisions

**Schema is HARDCODED** - Single source of truth for dependency locations

**Parser is DYNAMIC** - Uses schema paths to navigate runtime JSON

**Hierarchy separate from dependencies**:

- HAL links: Ownership (app owns policies)
- Field parsing: References (policy references flow)

**Array wildcard support** - Handles `items[*].id` notation

**Original data preserved** - Stored in output for re-parsing

**Comparison to Terraformer**:

- Terraformer: Simple hardcoded mappings, loses original data
- Our approach: Schema-driven, preserves data, tracks hierarchy
- Verdict: Our approach better for DaVinci's complex nested structures

---

## Phase 4.2: Generate Terraform References

**STATUS**: ⏳ NOT STARTED

**Goal**: Replace extracted dependency IDs with Terraform reference syntax.

### Reference Syntax

Terraform reference format:

```
<resource_type>.<resource_name>.<attribute>
```

Examples:

- Connection: `pingone_davinci_connector_instance.http_connector.id`
- Variable: `pingone_davinci_variable.api_key.id`
- Flow: `pingone_davinci_flow.registration.id`

### Implementation Plan

Update converter to use `DependencyGraph.GetReferenceName(type, id)`:

```go
func (c *Converter) generateFlowHCL(flow *Flow, graph *DependencyGraph) string {
    // When writing connectionId field:
    connectionName, err := graph.GetReferenceName("connector_instance", connectionID)
    if err != nil {
        // Generate TODO placeholder for missing dependency
        return fmt.Sprintf(`connection_id = "" # TODO: %s`, err)
    }
    // Generate reference
    return fmt.Sprintf("connection_id = pingone_davinci_connector_instance.%s.id", connectionName)
}
```

---

## Phase 4.3: Handle Missing Dependencies

**STATUS**: ⏳ NOT STARTED

**Goal**: Generate informative placeholders when dependencies not found.

### Missing Dependency Scenarios

Three types:

1. **Excluded by user** - `--exclude` flag filtered out
2. **Not included** - Not in `--include` filters
3. **Not found** - Doesn't exist in environment

### Placeholder Format

```hcl
connection_id = ""  # TODO: Reference to "PingOne Connector" (ID: abc123) was excluded from export
variable_id = ""    # TODO: Reference to variable xyz789 not found in environment
```

### Implementation Plan

Track missing dependencies with reasons:

```go
type MissingDependency struct {
    ResourceType string
    ResourceID   string
    Reason       MissingReason  // Excluded, NotIncluded, NotFound
}
```

Generate summary report after export.

---

## Phase 4.4: Validate Dependency Graph

**STATUS**: ⏳ NOT STARTED

**Goal**: Detect circular dependencies and optimize resource ordering.

### Circular Dependency Detection

Use depth-first search to find cycles:

```go
func (g *DependencyGraph) DetectCycles() [][]ResourceRef
```

Report to user:

```
ERROR: Circular dependency detected!
Cycle: flow_a → flow_b → flow_c → flow_a
```

### Topological Sort

Order resources by dependencies:

```go
func (g *DependencyGraph) TopologicalSort() ([]ResourceRef, error)
```

Benefits:

- Improved HCL readability
- Better Terraform plan output
- Easier debugging
---

## Integration with Converter (Phase 4.2-4.4)

After Phase 4.1 completion, remaining work:

### Update Converter to Use Resolver

```go
func (c *Converter) ConvertExport(export *exporter.ExportResult) (string, error) {
    // 1. Initialize resolver manager
    manager := resolver.NewResolverManager()
    
    // 2. Process all resources
    for _, flow := range export.Flows {
        manager.ProcessResource("flow", flow.ID, flow.Data, flow.Links)
    }
    // Process other resource types...
    
    // 3. Generate output with resolved dependencies
    output := manager.GenerateOutput()
    
    // 4. Validate (detect cycles)
    if cycles := output.Graph.DetectCycles(); len(cycles) > 0 {
        return "", fmt.Errorf("circular dependencies: %v", cycles)
    }
    
    // 5. Generate HCL with Terraform references
    hcl := c.generateHCLWithReferences(output)
    
    return hcl, nil
}
```

---

## Testing Strategy

### Unit Tests (Complete)

- ✅ 43/43 resolver package tests passing
- ✅ Parser path traversal
- ✅ Schema lookups
- ✅ Graph operations
- ✅ Hierarchy tracking

### Integration Tests (Phase 4.2)

Test complete workflow:

- Resource with dependencies → schema → parser → graph → Terraform references
- Multiple resource types with cross-dependencies
- Missing dependency handling
- Circular dependency detection

### End-to-End Tests (Phase 4.3-4.4)

- Export application with all dependencies
- Verify generated HCL has correct references
- Verify resources ordered by dependencies
- Verify missing dependencies generate TODOs

---

## Success Criteria

Phase 4.1: ✅ COMPLETE

- ✅ Dependency graph correctly identifies all resource relationships
- ✅ Schema-driven parser extracts dependencies from JSON
- ✅ Array wildcard paths supported (items[*].id)
- ✅ Hierarchy tracking separate from field dependencies
- ✅ Original data preserved for re-parsing
- ✅ 43/43 tests passing

Phase 4.2-4.4: ⏳ IN PROGRESS

Phase 4.2: ✅ COMPLETE
- ✅ Naming sanitization and uniqueness tracking implemented
- ✅ Terraform reference generation functional
- ✅ TODO placeholder generation for missing dependencies
- ✅ Integration tests validate complete workflow

Phase 4.3-4.4: ⏳ NOT STARTED
- ⏳ Missing dependency tracking (excluded, not included, not found reasons)
- ⏳ Circular dependency detection
- ⏳ Topological sort for resource ordering
- ⏳ User-facing missing dependency summary reports
- ⏳ Cycle detection error messages

**Integration Required**:
- ⏳ Update flow_converter.go to use GenerateTerraformReference()
- ⏳ Update flow_policy_converter.go to use GenerateTerraformReference()
- ⏳ Main converter integration with ResolverManager

---

**Next Steps**:

**Phase 4.1-4.2**: ✅ COMPLETE (53/53 tests passing)

**Immediate Options**:

1. **Complete Phase 4.3-4.4** (missing deps + validation)
   - Implement circular dependency detection (DFS)
   - Implement topological sort for resource ordering
   - Enhanced missing dependency tracking with reasons
   - User-facing error reports and warnings

2. **Integrate with Converters** (recommended)
   - Update flow_converter.go to use GenerateTerraformReference()
   - Update flow_policy_converter.go to use GenerateTerraformReference()
   - Validate resolver with real-world export data
   - Return to Phase 4.3-4.4 after proving architecture

3. **Proceed to Phase 5**
   - Move to final integration and error handling
   - Complete Phase 4.3-4.4 when integration requirements clearer

**Recommendation**: Option 2 - Integrate with converters to validate the resolver works end-to-end with real data before implementing advanced features like cycle detection.

**See**: `internal/resolver/PHASE_4_1_4_2_SUMMARY.md` for detailed implementation summary.

**After Part 4 complete**: Proceed to Part 5 (Final Integration and Error Handling).
