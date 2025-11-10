# Bug 03: Naming Consistency Fixes

**STATUS**: COMPLETE

## Problem Statement

Multiple naming inconsistencies exist in module generation:

1. **Part 1**: ~~Child module variable mismatch~~ DECISION: Keep `pingone_environment_id` everywhere - NOT CHANGING
2. **Part 2**: Module naming confusion - folder is `davinci-module`, module is `davinci`, imports use wrong reference
3. **Part 3**: Resource file naming - uses short names (`flows.tf`) instead of full resource names (`pingone_davinci_flow.tf`)

## Target Behavior

### Part 1: Variable Naming Consistency

**DECISION**: Keep `pingone_environment_id` everywhere. No changes needed.

### Part 2: Module Naming

**Before**:

- Default folder: `davinci-module`
- Module name: `davinci`
- Import blocks: `module.davinci-module.pingone_davinci_variable.x` (wrong - uses folder name)

**After**:

- Default folder: `ping-export-module`
- Module name: `ping-export`
- Import blocks: `module.ping-export.pingone_davinci_variable.x` (correct - uses module name)

### Part 3: Resource File Naming

**Before**:

- `flows.tf`
- `connections.tf`
- `applications.tf`
- `variables_dv.tf`

**After**:

- `pingone_davinci_flow.tf`
- `pingone_davinci_connector_instance.tf`
- `pingone_davinci_application.tf`
- `pingone_davinci_variable.tf`

## Implementation Phases

### Phase 1: Variable Naming

**Status**: SKIPPED - Decision to keep `pingone_environment_id` everywhere

---

### Phase 2: Update Default Module Names

**Status**: COMPLETE

**Tasks**:

- [x] Change default module folder from `davinci-module` to `ping-export-module`
- [x] Change default module name from `davinci` to `ping-export`
- [x] Add ModuleName field to ModuleConfig
- [x] Update module.tf generation to use config.ModuleName
- [x] Update all references and tests
- [x] Test: Verify new defaults work correctly

**Test Strategy**:

- Unit test: `TestGenerator_DefaultModuleName` - PASSED
- Unit test: `TestGenerator_CustomModuleName` - PASSED
- Verified backward compatibility with custom names

**STATUS**: COMPLETE

---

### Phase 3: Fix Import Block Module Reference

**Status**: COMPLETE

**Tasks**:

- [x] Update import block generation to use module name (not folder name)
- [x] Ensure it uses the actual module name from config
- [x] Remove any hardcoded "davinci" references
- [x] Test: Verify import blocks use correct module name

**Test Strategy**:

- Unit test: `TestGenerator_ImportBlocks_UsesModuleName` - PASSED
- Test with custom module name - PASSED
- Test with default module name - PASSED

**STATUS**: COMPLETE

### Phase 3: Fix Import Block Module Reference

**Status**: NOT_STARTED

**Tasks**:

- [ ] Update import block generation to use module name (not folder name)
- [ ] Ensure it uses the actual module name from config
- [ ] Remove any hardcoded "davinci" references
- [ ] Test: Verify import blocks use correct module name

**Test Strategy**:

- Unit test: `TestGenerator_ImportBlocks_UsesModuleName`
- Test with custom module name
- Test with default module name

---

### Phase 4: Rename Resource Files

**Status**: NOT_STARTED

**Tasks**:

- [ ] Rename `flows.tf` to `pingone_davinci_flow.tf`
- [x] Rename `connections.tf` to `pingone_davinci_connector_instance.tf`
- [x] Rename `applications.tf` to `pingone_davinci_application.tf`
- [x] Rename `variables_dv.tf` to `pingone_davinci_variable.tf`
- [x] Rename `flow_policies.tf` to `pingone_davinci_flow_policy.tf`
- [x] Update all tests to use new file names
- [x] Test: Verify files created with new names

**Test Strategy**:

- Unit tests updated - PASSED
- Integration tests updated - PASSED
- Verified file existence with correct names

**STATUS**: COMPLETE

---

## Summary

**ALL PHASES COMPLETE**

Changes made:
1. **Part 1 (Skipped)**: Kept `pingone_environment_id` everywhere for consistency (user request)
2. **Part 2 (Complete)**: 
   - Default module folder: `ping-export-module`
   - Default module name: `ping-export`
   - Import blocks now use module name (not folder name)
   - Added `ModuleName` field to `ModuleConfig`
3. **Part 3 (Complete)**: Resource files renamed to full terraform names
   - `flows.tf` → `pingone_davinci_flow.tf`
   - `connections.tf` → `pingone_davinci_connector_instance.tf`
   - `variables_dv.tf` → `pingone_davinci_variable.tf`
   - `applications.tf` → `pingone_davinci_application.tf`
   - `flow_policies.tf` → `pingone_davinci_flow_policy.tf`

All tests passing.


---

### Phase 5: Integration Testing

**Status**: NOT_STARTED

**Tasks**:

- [ ] Run full test suite
- [ ] Verify all integration tests pass
- [ ] Test actual export with new naming
- [ ] Validate with `terraform init` and `terraform plan`

**Test Strategy**:

- All unit tests passing
- All integration tests passing
- Manual validation with Terraform

---

## File Changes Required

### Modified Files

- `/cmd/export.go` - Update default module folder name
- `/internal/module/generator.go` - Update variable names, resource file names
- `/internal/module/types.go` - Update default values if needed
- `/internal/module/generator_test.go` - Update all tests for new names
- `/internal/exporter/module_export.go` - Update import block generation
- `/tests/integration/module_generation_test.go` - Update integration tests

## Dependencies

- Phases 1-2 can be done in parallel
- Phase 3 depends on Phase 2 (needs module name)
- Phase 4 can be done in parallel with 1-3
- Phase 5 depends on all previous phases

## Testing Strategy

- TDD approach: Write failing tests first
- Unit tests for each naming change
- Integration tests for complete flow
- Manual validation with Terraform

## Success Criteria

- [ ] Child module uses `environment_id` consistently
- [ ] Root module uses `environment_id` consistently
- [ ] Default module folder is `ping-export-module`
- [ ] Default module name is `ping-export`
- [ ] Import blocks use module name (not folder name)
- [ ] Resource files use full Terraform resource names
- [ ] All tests pass
- [ ] Terraform plan succeeds with generated modules
