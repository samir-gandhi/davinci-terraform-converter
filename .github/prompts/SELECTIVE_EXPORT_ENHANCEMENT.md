# Selective Export Enhancement - Summary

## Overview

Added comprehensive guidance for implementing selective resource export with automatic dependency discovery using HAL (Hypertext Application Language) links from the PingOne API.

## What Changed

### 1. New Selective Export Flags (Part 3 Prerequisites)

Added resource filtering flags that allow users to export specific resources with granular control:

- **Include flags**: Export ONLY specified resources
  - `--include-flows`, `--include-applications`, `--include-connections`, `--include-variables`
- **Exclude flags**: Export everything EXCEPT specified resources
  - `--exclude-flows`, `--exclude-applications`, `--exclude-connections`, `--exclude-variables`
- **Dependency control**:
  - `--with-dependencies` (default: true) - Auto-include all dependencies
  - `--no-dependencies` - Export only specified resources

**Key Features**:
- Support both IDs (UUIDs) and names (fuzzy matching)
- Comma-separated lists for multiple resources
- Include takes precedence over exclude

**Use Cases**:
```bash
# Export single flow with all its dependencies
pingcli davinci convert --export --include-flows "Registration Flow" --with-dependencies

# Export multiple flows without dependencies
pingcli davinci convert --export --include-flows "flow1,flow2" --no-dependencies

# Export everything except test resources
pingcli davinci convert --export --exclude-flows "test" --exclude-applications "test"
```

### 2. Phase 3.6: HAL Link-Based Dependency Discovery

Added new phase for implementing selective export functionality:

**HAL Link Parsing**:
- PingOne API responses include HAL `_links` sections showing resource relationships
- More reliable than parsing JSON structure alone
- Provides bidirectional relationships
- Example: flows have links to connector instances, variables, subflows

**Implementation Components**:
- `internal/api/hal.go` - HAL link parser
- `internal/filter` package - Resource filtering logic
- `DependencyDiscoverer` - Recursive dependency tree builder

**Workflow**:
1. User specifies resources to include (e.g., one flow)
2. Tool fetches that resource's API response
3. Parse HAL links to discover dependencies (connectors, variables, subflows)
4. Recursively fetch dependency HAL links
5. Build complete dependency tree
6. Export all discovered resources

### 3. Enhanced Dependency Graph (Part 4.1)

Updated dependency resolution to use multiple sources:

**Hybrid Approach**:
- **HAL Links**: High-level relationships from API responses
- **JSON Structure**: Detailed node-level dependencies from flow graphData
- **Cross-validation**: Use both sources for completeness and accuracy

### 4. Enhanced Missing Dependency Handling (Part 4.3)

Distinguished between different types of missing dependencies in selective exports:

**Three Categories**:
1. **Missing by exclusion**: User explicitly excluded via `--exclude` flag
2. **Missing by selection**: User used `--include` without dependencies
3. **Actually missing**: Resource doesn't exist in environment

**Different Placeholders**:
```hcl
# Missing by user exclusion
connection_id = "" # TODO: Reference to "PingOne Connector" (ID: abc123) was excluded from export

# Missing by selective export
variable_id = "" # TODO: Reference to "apiKey" (ID: xyz789) was not included in export filters

# Actually missing
subflow_id = "" # TODO: Reference to flow ID def456 not found in environment
```

### 5. Enhanced Testing (Part 5.2)

Added test scenarios for selective export:

**New Test Cases**:
- Export with `--include-flows` filters correctly
- Export with `--with-dependencies` includes all dependencies
- Export with `--no-dependencies` excludes dependencies
- Export with `--exclude-*` filters correctly
- Invalid filter combinations return helpful errors
- HAL link parsing discovers dependencies correctly
- End-to-end selective export with dependency validation

### 6. Enhanced Documentation (Part 5.3)

Added documentation for selective export:

**README Updates**:
- Document include/exclude filter syntax
- Explain HAL link-based dependency discovery
- Provide use case examples

**New Examples**:
- `examples/selective-export-single-flow.sh`
- `examples/selective-export-app.sh`
- `examples/selective-export-no-deps.sh`

**Troubleshooting**:
- HAL link parsing issues
- Filter matching behavior (ID vs name, case sensitivity, fuzzy matching)

### 7. Future Enhancements (Part 5.5)

Documented advanced filtering capabilities for future implementation:

- Regex patterns for resource names
- Filter by resource attributes (e.g., enabled flows only)
- Filter by tags or metadata
- Time-based filtering (resources modified after date)
- Dependency visualization (generate graphs)
- Export profiles (save/share common filter combinations)

## Implementation Timeline

**Phase 3.6 Status**: Future enhancement after basic export (Phase 3.1-3.5) is working

**Prerequisites**:
- Part 3 (Phase 3.1-3.5): Basic full export must be functional
- Part 4: Dependency resolution must be implemented
- HAL link parsing infrastructure must be in place

**Implementation Order**:
1. Implement basic full export (Part 3, Phases 3.1-3.5)
2. Implement dependency resolution (Part 4)
3. Add HAL link parser (`internal/api/hal.go`)
4. Add resource filter (`internal/filter` package)
5. Implement dependency discovery via HAL links
6. Add CLI flags for selective export
7. Test with various scenarios
8. Document and provide examples

## Testing Strategy

**Unit Tests**:
- HAL link parser with various response formats
- Resource filter logic (include/exclude combinations)
- Dependency discoverer (single-level and recursive)

**Integration Tests**:
- Mock API responses with HAL links
- Selective export end-to-end scenarios
- Dependency validation with partial exports

**Manual Testing**:
- Real PingOne environments with selective exports
- Various filter combinations
- Edge cases (circular dependencies, missing resources)

## Benefits

1. **Granular Control**: Export exactly what you need, not entire environments
2. **Automatic Dependencies**: No need to manually track what each resource needs
3. **Faster Exports**: Only fetch required resources via API
4. **Smaller HCL Files**: Easier to review and manage
5. **Team Workflows**: Different teams can export their specific resources
6. **Testing**: Export test flows without production resources
7. **Migration**: Gradually migrate resources by exporting subsets

## Example Scenarios

### Scenario 1: Developer working on single flow
```bash
# Export just the registration flow with all dependencies
pingcli davinci convert --export \
  --include-flows "User Registration" \
  --with-dependencies \
  --out registration.tf
```

**Result**: Exports the flow + any connector instances, variables, and subflows it references.

### Scenario 2: QA team exporting test resources only
```bash
# Export all flows EXCEPT production flows
pingcli davinci convert --export \
  --exclude-flows "prod,production" \
  --out test-resources.tf
```

**Result**: Exports all resources except those with "prod" or "production" in their names.

### Scenario 3: Application owner exporting their app
```bash
# Export specific application and its flow policies (without the actual flows)
pingcli davinci convert --export \
  --include-applications "Customer Portal" \
  --no-dependencies \
  --out customer-portal-app.tf
```

**Result**: Exports only the application resource, not the flows it references.

## Notes

- HAL link parsing is more reliable than JSON structure parsing alone
- Dependency discovery is recursive (dependencies of dependencies)
- Circular dependencies are detected and handled
- Missing dependencies get clear TODO comments with context
- Implementation should be incremental (basic export first, then selective)
