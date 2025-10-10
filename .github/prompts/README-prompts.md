---
mode: agent
---

# DaVinci Terraform Converter - Modular Prompts

This directory contains modular prompt files for the DaVinci Terraform Converter project. The prompts are organized by project phase to reduce context requirements and enable focused development.

## Why Modular Prompts?

**Problem**: The original monolithic prompt (1,289 lines) required excessive context for focused work.

**Solution**: Break into 10 focused modules that can be loaded selectively based on current work phase.

**Result**: 73% context reduction (from 1,289 lines to ~350 lines for active work).

---

## Prompt Structure

### Core Context (Always Load)

**00-project-context.md** (~96 lines)
- Project overview and goals
- Technology stack
- Current state across all parts
- File structure
- Development approach (TDD principles)

**Use**: Load at the start of every work session.

### Completed Work (Reference Only)

**01-part1-scaffolding.md** (~60 lines)
- ✅ COMPLETE - Reference only
- Project initialization
- Cobra command structure
- Initial converter setup
- Tests and validation

**Use**: Context for what's already done, not active work.

### Active Development

**02-part2-flow-conversion.md** (~260 lines)
- 🔧 IN PROGRESS - Current focus (88.9% complete, 24/27 tests)
- Phase 2.1: Complete flow structure conversion
- TDD approach with detailed steps
- Flow structure mapping (graphData, settings, variables, schemas)
- HCL syntax requirements (CRITICAL: attribute vs block syntax)
- Edge cases and test status

**Use**: Load when working on Phase 2.1 (flow conversion).

**03-part2-other-resources.md** (~520 lines)
- ⏳ NOT STARTED - After Phase 2.1
- Phase 2.2: Applications
- Phase 2.3: Flow Policies
- Phase 2.4: Connections
- Phase 2.5: Variables
- Phase 2.6: Multi-resource integration
- Complete JSON → HCL examples for each type

**Use**: Load when working on Phases 2.2-2.6.

**04-part2-terraform-validation.md** (~370 lines)
- ⏳ NOT STARTED - After Phases 2.1-2.6
- Integration tests with Terraform CLI
- Test environment setup
- Five test implementations (with Go code)
- Mock vs real testing strategies

**Use**: Load when implementing Terraform validation tests.

**05-part3-api-export.md** (~580 lines)
- ⏳ NOT STARTED - After Part 2
- PingCLI authentication integration
- API client for all resource types
- Export orchestration
- Phase 3.6: Selective export with HAL link parsing

**Use**: Load when implementing API export functionality.

**06-part4-dependencies.md** (~420 lines)
- ⏳ NOT STARTED - After Part 3
- Dependency graph building
- Terraform reference generation
- Missing dependency handling
- Circular dependency detection

**Use**: Load when implementing dependency resolution.

**07-part5-integration.md** (~550 lines)
- ⏳ NOT STARTED - After Parts 2-4
- Complete CLI integration (file mode + export mode)
- Comprehensive integration tests
- Documentation and examples
- Performance optimization

**Use**: Load when integrating all components.

**08-part6-production.md** (~370 lines)
- ⏳ NOT STARTED - After Parts 2-5
- Security hardening
- Testing and QA
- Release preparation
- Versioning and distribution

**Use**: Load when preparing for production release.

---

## Usage Patterns

### Pattern 1: Current Work (Phase 2.1)

Load these files:
```
@00-project-context.md
@02-part2-flow-conversion.md
```

**Context**: 96 + 260 = ~356 lines (vs 1,289 in original)
**Focus**: Complete flow structure conversion, fix failing tests

### Pattern 2: Other Resource Types (Phases 2.2-2.6)

Load these files:
```
@00-project-context.md
@03-part2-other-resources.md
```

**Context**: 96 + 520 = ~616 lines
**Focus**: Implement application, flow policy, connection, variable conversion

### Pattern 3: Terraform Validation

Load these files:
```
@00-project-context.md
@02-part2-flow-conversion.md
@04-part2-terraform-validation.md
```

**Context**: 96 + 260 + 370 = ~726 lines
**Focus**: Add integration tests with Terraform CLI

### Pattern 4: API Export Implementation

Load these files:
```
@00-project-context.md
@05-part3-api-export.md
```

**Context**: 96 + 580 = ~676 lines
**Focus**: Implement PingCLI auth, API clients, export orchestration

### Pattern 5: Dependency Resolution

Load these files:
```
@00-project-context.md
@06-part4-dependencies.md
```

**Context**: 96 + 420 = ~516 lines
**Focus**: Build dependency graph, generate Terraform references

### Pattern 6: Final Integration

Load these files:
```
@00-project-context.md
@07-part5-integration.md
```

**Context**: 96 + 550 = ~646 lines
**Focus**: Wire everything together, comprehensive tests, documentation

### Pattern 7: Production Release

Load these files:
```
@00-project-context.md
@08-part6-production.md
```

**Context**: 96 + 370 = ~466 lines
**Focus**: Security, QA, versioning, release process

### Pattern 8: Cross-Phase Context

When working on dependencies between phases, load relevant combination:
```
@00-project-context.md
@02-part2-flow-conversion.md
@06-part4-dependencies.md
```

**Context**: 96 + 260 + 420 = ~776 lines
**Focus**: Understand how converter output feeds into dependency resolver

---

## Quick Reference Table

| Work Phase | Load Files | Total Context | Focus Area |
|------------|-----------|---------------|------------|
| Phase 2.1 | 00, 02 | ~356 lines | Flow conversion |
| Phase 2.2-2.6 | 00, 03 | ~616 lines | Other resources |
| Terraform tests | 00, 02, 04 | ~726 lines | Integration tests |
| Part 3 | 00, 05 | ~676 lines | API export |
| Part 4 | 00, 06 | ~516 lines | Dependencies |
| Part 5 | 00, 07 | ~646 lines | Integration |
| Part 6 | 00, 08 | ~466 lines | Production |

---

## File Status Legend

- ✅ **COMPLETE**: Work finished, reference only
- 🔧 **IN PROGRESS**: Active development (includes test pass %)
- ⏳ **NOT STARTED**: Future work, load when ready to begin

---

## Critical Notes

### HCL Syntax (Phase 2.1)

**CRITICAL**: `flow_configuration_json` uses attribute syntax, NOT block syntax:

```hcl
# CORRECT - Attribute with =
flow_configuration_json = {
  graphData = { ... }
}

# WRONG - Block syntax
flow_configuration_json {
  graphData = { ... }
}

# WRONG - JSON string
flow_configuration_json = jsonencode({ ... })
```

### HAL Link Parsing (Phases 3.6, 4.1)

PingOne API responses use HAL format with `_links` sections for dependency discovery:

```json
{
  "id": "flow123",
  "_links": {
    "connectorInstances": { "href": "..." },
    "variables": { "href": "..." }
  }
}
```

Use HAL links for dependency discovery in selective exports.

### Resource Ordering (Phase 2.6, 4.4)

Resources must be generated in dependency order:
1. Variables
2. Connections
3. Flows
4. Applications
5. Flow Policies

---

## Updating Prompts

When updating prompts:

1. **Keep 00-project-context.md in sync**: Update "Current State" section as phases complete
2. **Maintain focus**: Each module should cover one coherent phase
3. **Cross-reference**: Link to related modules when dependencies exist
4. **Update this README**: Reflect any structural changes

---

## Legacy File

**davinci-converter-LEGACY.prompt.md** (1,289 lines)
- Original monolithic prompt
- Preserved for historical reference
- Not used for active development

---

## Questions?

If unsure which files to load:
1. Always start with `00-project-context.md`
2. Load the module matching your current phase
3. Add related modules if working across phases
4. When in doubt, check this README's Quick Reference Table

---

**Last Updated**: 2024-01-15
**Prompt Version**: 1.0
**Project Status**: Part 1 Complete, Part 2.1 88.9% (24/27 tests)
