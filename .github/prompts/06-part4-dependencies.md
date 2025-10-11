---
mode: agent
---

# Part 4: Dependency Resolution and Terraform References

**Status**: ⏳ NOT STARTED (After Part 3 complete)

**Prerequisites**:
- Part 3 (full export) must be complete
- All resource types can be exported
- HAL link parsing functional (for Phase 3.6)

**Goal**: Replace hardcoded resource IDs with Terraform references to enable proper resource dependencies.

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

**Goal**: Discover all resource relationships.

### Create Dependency Resolver

Create `internal/resolver/resolver.go`:

```go
package resolver

type ResourceRef struct {
    Type string  // "flow", "connection", "variable", etc.
    ID   string  // Original resource ID
    Name string  // Sanitized Terraform resource name
}

type Dependency struct {
    From     ResourceRef  // Dependent resource
    To       ResourceRef  // Dependency target
    Field    string       // Field name containing reference
    Location string       // Location in structure (e.g., "graphData.nodes[5].data.properties.connectionId")
}

type DependencyGraph struct {
    resources    map[string]ResourceRef  // ID -> ResourceRef
    dependencies []Dependency
}

func NewDependencyGraph() *DependencyGraph {
    return &DependencyGraph{
        resources:    make(map[string]ResourceRef),
        dependencies: make([]Dependency, 0),
    }
}

func (g *DependencyGraph) AddResource(resourceType, id, name string) {
    // Register resource
}

func (g *DependencyGraph) AddDependency(from, to ResourceRef, field, location string) {
    // Register dependency relationship
}

func (g *DependencyGraph) GetDependencies(resourceID string) []Dependency {
    // Get all dependencies for a resource
}

func (g *DependencyGraph) GetReferenceName(resourceType, resourceID string) (string, error) {
    // Get Terraform resource name for a given ID
    // Returns error if resource not found
}
```

### Identify Dependency Types

Parse exported data to find all references:

**Flow Dependencies**:
- `connectionId` fields in flow nodes → connector instance resources
- Variable references in node properties → variable resources
- Subflow node references → other flow resources

**Flow Policy Dependencies**:
- `flowId` in policy_flow blocks → flow resources
- `applicationId` (if present) → application resources

**Application Dependencies**:
- Flow policy associations → flow policy resources

### Parse Multiple Dependency Sources

Use hybrid approach combining two sources:

**1. JSON Structure Parsing** (detailed):
Parse flow `graphData` to find:
```json
{
  "graphData": {
    "elements": {
      "nodes": [
        {
          "data": {
            "connectionId": "abc123",  // Reference to connection
            "properties": {
              "variableId": "xyz789"  // Reference to variable
            }
          }
        },
        {
          "data": {
            "nodeType": "FLOW",
            "capabilityName": "startSubFlowNode",
            "properties": {
              "subFlowId": "def456"  // Reference to subflow
            }
          }
        }
      ]
    }
  }
}
```

**2. HAL Link Parsing** (high-level):
Extract from API response `_links`:
```json
{
  "_links": {
    "connectorInstances": { "href": ".../connectorInstances?flowId=flow123" },
    "variables": { "href": ".../variables?flowId=flow123" },
    "subflows": { "href": ".../flows?parentFlowId=flow123" }
  }
}
```

HAL links are more reliable but less detailed. JSON structure gives exact field locations.

Create `internal/resolver/parser.go`:

```go
func FindReferencesInFlow(flow *Flow) []Dependency {
    dependencies := []Dependency{}
    
    // Parse JSON structure for connectionId fields
    // Parse JSON structure for variable references
    // Parse JSON structure for subflow references
    // Parse HAL links for validation
    
    return dependencies
}

func FindReferencesInFlowPolicy(policy *FlowPolicy) []Dependency {
    // Parse flowId references
    // Parse applicationId references
    return dependencies
}
```

### Test Dependency Detection

Create `internal/resolver/resolver_test.go`:

Test cases:
- Flow with single connection dependency
- Flow with multiple connection dependencies
- Flow with variable references
- Flow with subflow references
- Flow with all dependency types
- Flow policy with flow reference
- Application with flow policy references
- Resources with circular dependencies (should detect)
- Resources with no dependencies

Mock data for each test case showing expected dependency graph.

---

## Phase 4.2: Generate Terraform References

**Goal**: Replace IDs with valid Terraform references.

### Reference Syntax

Terraform reference format:
```
<resource_type>.<resource_name>.<attribute>
```

Examples:
- Connection: `pingone_davinci_connection.http_connector.id`
- Variable: `pingone_davinci_variable.api_key.id`
- Flow: `pingone_davinci_flow.registration.id`
- Application: `pingone_davinci_application.my_app.id`

### Resource Naming

Generate valid Terraform resource names from human-readable names.

Create `internal/resolver/naming.go`:

```go
func SanitizeName(name string) string {
    // Convert to lowercase
    // Replace spaces with underscores
    // Remove special characters (keep alphanumeric and underscores)
    // Ensure starts with letter
    // Ensure uniqueness (append counter if needed)
}

// Example transformations:
// "My HTTP Connector" -> "my_http_connector"
// "API Key (Production)" -> "api_key_production"
// "Registration Flow" -> "registration_flow"
// "Registration Flow" (duplicate) -> "registration_flow_2"
```

Maintain bidirectional mappings:
- Original name → Terraform name
- Original ID → Terraform name

### Update HCL Generation

Modify converter functions to use references:

**For Flows** (update `generateFlowHCL()`):
```go
func (c *Converter) generateFlowHCL(flow *Flow, graph *DependencyGraph) (string, error) {
    // When writing connectionId field:
    connectionName, err := graph.GetReferenceName("connection", connectionID)
    if err != nil {
        // Connection not found - generate TODO placeholder
        return fmt.Sprintf(`connection_id = "" # TODO: Reference missing for ID: %s`, connectionID)
    }
    // Generate reference
    return fmt.Sprintf("connection_id = pingone_davinci_connection.%s.id", connectionName)
}
```

**For Flow Policies** (update `generateFlowPolicyHCL()`):
```go
func (c *Converter) generateFlowPolicyHCL(policy *FlowPolicy, graph *DependencyGraph) (string, error) {
    // When writing flow_id field:
    flowName, err := graph.GetReferenceName("flow", flowID)
    if err != nil {
        return fmt.Sprintf(`flow_id = "" # TODO: Reference missing for ID: %s`, flowID)
    }
    return fmt.Sprintf("flow_id = pingone_davinci_flow.%s.id", flowName)
}
```

**For Variables in Flow Nodes**:
Handle variable references in node properties using same pattern.

### Test Reference Generation

Create `internal/resolver/reference_test.go`:

Test cases:
- Generate reference for existing connection
- Generate reference for existing variable
- Generate reference for existing flow
- Generate TODO placeholder for missing connection
- Generate TODO placeholder for missing variable
- Verify reference syntax validity
- Test name sanitization (special chars, spaces, duplicates)

---

## Phase 4.3: Handle Missing Dependencies

**Goal**: Gracefully handle references to resources not in export.

### Detect Orphaned References

Occurs when:
- Resource was deleted but still referenced
- Selective export (Phase 3.6) excluded dependency
- Resource is in different environment
- HAL link points to non-existent resource

### Generate Placeholder Comments

Three types of missing dependencies:

**1. Missing by User Exclusion**:
```hcl
connection_id = ""  # TODO: Reference to "PingOne Connector" (ID: abc123) was excluded from export
```

**2. Missing by Selection** (not in include filters):
```hcl
variable_id = ""  # TODO: Reference to "apiKey" (ID: xyz789) was not included in export filters
```

**3. Actually Missing** (doesn't exist):
```hcl
subflow_id = ""  # TODO: Reference to flow ID def456 not found in environment
```

Create `internal/resolver/missing.go`:

```go
type MissingReason int

const (
    MissingExcluded MissingReason = iota  // User excluded with --exclude flag
    MissingNotIncluded                    // Not in --include filter
    MissingNotFound                       // Doesn't exist in environment
)

type MissingDependency struct {
    ResourceType string
    ResourceID   string
    ResourceName string  // If available
    Reason       MissingReason
}

func GenerateTODOPlaceholder(missing MissingDependency) string {
    switch missing.Reason {
    case MissingExcluded:
        if missing.ResourceName != "" {
            return fmt.Sprintf(`"" # TODO: Reference to "%s" (ID: %s) was excluded from export`,
                missing.ResourceName, missing.ResourceID)
        }
        return fmt.Sprintf(`"" # TODO: Reference to %s %s was excluded from export`,
            missing.ResourceType, missing.ResourceID)
    
    case MissingNotIncluded:
        if missing.ResourceName != "" {
            return fmt.Sprintf(`"" # TODO: Reference to "%s" (ID: %s) was not included in export filters`,
                missing.ResourceName, missing.ResourceID)
        }
        return fmt.Sprintf(`"" # TODO: Reference to %s %s was not included in export filters`,
            missing.ResourceType, missing.ResourceID)
    
    case MissingNotFound:
        return fmt.Sprintf(`"" # TODO: Reference to %s %s not found in environment`,
            missing.ResourceType, missing.ResourceID)
    }
}
```

### Track Missing Dependencies

Update `DependencyGraph`:

```go
type DependencyGraph struct {
    // ... existing fields
    missing map[string]MissingDependency  // ID -> MissingDependency
}

func (g *DependencyGraph) RecordMissing(resourceType, resourceID string, reason MissingReason) {
    // Record missing dependency with reason
}

func (g *DependencyGraph) GetMissingSummary() map[MissingReason][]MissingDependency {
    // Group missing dependencies by reason
    // For user reporting
}
```

### Warn User About Missing Dependencies

After export complete, print summary:

```
Export complete! Generated HCL with 42 resources.

WARNING: Found 3 missing dependencies:
  - 1 resource excluded by filters (see TODO comments in output)
  - 2 resources not found in environment (may be deleted)

Review TODO comments in generated HCL before applying.
```

### Test Missing Dependency Handling

Create `internal/resolver/missing_test.go`:

Test cases:
- Generate placeholder for excluded resource
- Generate placeholder for not-included resource
- Generate placeholder for not-found resource
- Verify placeholder includes original ID
- Verify placeholder includes resource name (if available)
- Test summary generation
- Test with multiple missing dependencies

---

## Phase 4.4: Validate Dependency Graph

**Goal**: Detect and report issues in dependency graph.

### Circular Dependency Detection

Terraform cannot handle circular dependencies.

Create `internal/resolver/cycles.go`:

```go
func (g *DependencyGraph) DetectCycles() ([][]ResourceRef, error) {
    // Use depth-first search to detect cycles
    // Return all cycles found
    // Each cycle is a slice of resources forming the loop
}
```

Algorithm:
1. Start DFS from each unvisited node
2. Track path from root to current node
3. If we visit a node already in current path → cycle detected
4. Record cycle and continue to find all cycles

Report cycles to user:
```
ERROR: Circular dependencies detected!

Cycle 1: flow_a → flow_b → flow_c → flow_a
  - flow_a (ID: abc123) references subflow flow_b
  - flow_b (ID: def456) references subflow flow_c
  - flow_c (ID: ghi789) references subflow flow_a

Terraform cannot handle circular dependencies. Manual intervention required.
```

### Dependency Ordering

Generate HCL with resources in dependency order:

Resources with no dependencies first.
Dependent resources after their dependencies.

Create `internal/resolver/ordering.go`:

```go
func (g *DependencyGraph) TopologicalSort() ([]ResourceRef, error) {
    // Perform topological sort on dependency graph
    // Return resources in order: dependencies before dependents
    // Return error if cycles detected (cannot sort)
}
```

Algorithm:
1. Calculate in-degree for each node (number of dependencies)
2. Add nodes with in-degree 0 to queue
3. Process queue:
   - Remove node from queue, add to result
   - Decrease in-degree of dependent nodes
   - Add nodes with in-degree 0 to queue
4. If all nodes processed → valid order
5. If nodes remain → cycle exists

Benefits:
- Improved readability of generated HCL
- Better Terraform plan output (dependencies applied first)
- Easier debugging of reference issues

### Test Dependency Validation

Create `internal/resolver/validation_test.go`:

**Cycle Detection Tests**:
- No cycles (linear dependencies)
- Simple cycle (A → B → A)
- Complex cycle (A → B → C → D → B)
- Multiple separate cycles
- Large graph with no cycles (performance test)

**Ordering Tests**:
- Simple linear chain (A → B → C)
- Diamond dependency (A → B, A → C, B → D, C → D)
- No dependencies (any order valid)
- Multiple independent chains
- Verify error when cycles present

**Complex Scenario Tests**:
- Multi-resource export with realistic dependencies
- Flow → multiple connections
- Flow → variables → connections (transitive)
- Application → flow policy → flow → connection
- Verify final HCL has correct resource order

---

## Integration with Converter

Update `converter.ConvertExport()`:

```go
func (c *Converter) ConvertExport(export *exporter.ExportResult) (string, error) {
    // 1. Build dependency graph
    graph := resolver.NewDependencyGraph()
    
    // 2. Register all resources
    for _, flow := range export.Flows {
        name := resolver.SanitizeName(flow.Name)
        graph.AddResource("flow", flow.ID, name)
    }
    // ... register other resource types
    
    // 3. Parse dependencies
    for _, flow := range export.Flows {
        deps := resolver.FindReferencesInFlow(&flow)
        for _, dep := range deps {
            graph.AddDependency(dep.From, dep.To, dep.Field, dep.Location)
        }
    }
    // ... parse dependencies for other types
    
    // 4. Detect cycles
    cycles, err := graph.DetectCycles()
    if len(cycles) > 0 {
        return "", fmt.Errorf("circular dependencies detected: %v", cycles)
    }
    
    // 5. Order resources
    orderedResources, err := graph.TopologicalSort()
    if err != nil {
        return "", err
    }
    
    // 6. Generate HCL in dependency order
    var hcl strings.Builder
    for _, resource := range orderedResources {
        switch resource.Type {
        case "flow":
            flow := findFlow(export.Flows, resource.ID)
            flowHCL, err := c.generateFlowHCL(flow, graph)
            if err != nil {
                return "", err
            }
            hcl.WriteString(flowHCL)
        // ... handle other resource types
        }
    }
    
    // 7. Report missing dependencies
    missing := graph.GetMissingSummary()
    if len(missing) > 0 {
        printMissingSummary(missing)
    }
    
    return hcl.String(), nil
}
```

---

## Success Criteria

- ✅ Dependency graph correctly identifies all resource relationships
- ✅ Hardcoded IDs replaced with Terraform references
- ✅ Resource names are sanitized and unique
- ✅ Missing dependencies generate TODO placeholders with reasons
- ✅ Circular dependencies are detected and reported
- ✅ Generated HCL has resources in dependency order
- ✅ Integration tests verify complete dependency resolution
- ✅ User receives clear warnings about missing dependencies

---

**Next Step**: After Part 4 complete, proceed to Part 5 (Final Integration and Error Handling).
