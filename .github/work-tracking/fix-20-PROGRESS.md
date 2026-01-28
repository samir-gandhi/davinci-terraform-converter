# Bug Fix 20 Progress

## Analysis
The Davinci Flow resource schema has changed significantly between the old and new versions.

### Differences
1.  `graph_data.elements.edges` and `graph_data.elements.nodes`:
    *   **Old**: Set of Attributes (`schema.SetNestedAttribute`)
    *   **New**: Map of Attributes (`schema.MapNestedAttribute`)
    *   **Impact**: The HCL generation must now produce a map `key = { ... }` instead of a list `[ ... ]`. The key appears to be the `id` of the element.
2.  `position` attributes (`x`, `y`):
    *   **Old**: `Float64Attribute`
    *   **New**: `NumberAttribute`
    *   **Impact**: Internal types need to change from `types.Float64` to `types.Number`.
3.  `nodes.data.properties`:
    *   **Old**: Not sensitive.
    *   **New**: Sensitive.
    *   **Impact**: Minimal impact on HCL generation, but good to note.
4.  Computed fields (`connection_id`, `id_unique`, `name`, `input_schema.required`):
    *   Now marked as Computed in schema.

## Plan

### Phase 1: Update Internal Types and Converter Logic
- Update internal struct definitions to match new `Map` structure for nodes/edges.
- Update `internal/converter` logic to iterate over nodes/edges and construct a Map instead of a List.
- Use `node.data.id` (or equivalent) as the map key.

### Phase 2: Update Number Types
- Change `x`, `y` and other number fields from `types.Float64` to `types.Number` where applicable in the struct definitions used for HCL generation.

### Phase 3: Test Updates
- Update `flow_converter_test.go` and other relevant tests to expect Map syntax in HCL output.
- `nodes = { "id1" = { ... }, "id2" = { ... } }` instead of `nodes = [ { ... }, { ... } ]`.

### Phase 4: Verification
- Run full test suite.
- Manual verify with an example if possible.

## Status Toggles
- [x] Phase 1: Update Internal Types and Converter Logic
- [x] Phase 2: Update Number Types
- [x] Phase 3: Test Updates
- [ ] Phase 4: Verification
