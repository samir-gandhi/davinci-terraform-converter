# Export/Import Changes Analysis

## Objective
Achieve zero-change terraform plan after export → import cycle. Export should define resources matching imported state exactly, avoiding upstream breaks and requiring no manual intervention.

## Analysis Date
October 24, 2025

## Plan Summary

**Total Resources**: 41 resources affected
- **Will be created**: 1 resource
- **Will be updated**: 22 resources  
- **Clean imports**: 18 resources (no changes required)

---

## Change Categories

### CRITICAL - Category 1: Flow Policy Trigger Field Missing
**Priority**: P0 - Causes changes on 5 flow policies
**Affected Resources**: 5 `pingone_davinci_application_flow_policy` resources

**Issue**: Exported HCL omits `trigger` block entirely when it should export default values.

**Changes Required**:
```hcl
+ trigger = {
+   configuration = {
+     mfa = {
+       enabled     = false
+       time        = 0
+       time_format = "min"
+     }
+     pwd = {
+       enabled     = false  
+       time        = 0
+       time_format = "min"
+     }
+   }
+   type = "AUTHENTICATION"
+ }
```

**Affected Resources**:
1. `pingone_davinci_application_flow_policy.pingcli__DaVinci-0020-API-0020-Protect-0020-Sample-0020-Policy`
2. `pingone_davinci_application_flow_policy.pingcli__DaVinci-0020-API-0020-Protect-0020-Sample-0020-Policy_2`
3. `pingone_davinci_application_flow_policy.pingcli__New-0020-Policy`
4. `pingone_davinci_application_flow_policy.pingcli__New-0020-Policy_2`
5. `pingone_davinci_application_flow_policy.pingcli__New-0020-Policy_3`
6. `pingone_davinci_application_flow_policy.pingcli__abcpolicy`

**Note**: 2 other flow policies (`New-0020-Policy_4`, `pingonePolicy`) are imported with explicit trigger blocks and work correctly.

**Root Cause**: Export logic not handling default/computed trigger attribute.

**Fix Location**: `cmd/davinci_to_hcl.go` - application flow policy export logic

---

### CRITICAL - Category 2: Application OAuth Scopes Missing
**Priority**: P0 - Causes changes on 1 application
**Affected Resources**: 1 `pingone_davinci_application` resource

**Issue**: When application has OAuth enabled but no explicit scopes, export omits scopes array. Import adds default scopes `["openid", "profile"]`.

**Changes Required**:
```hcl
oauth = {
  # ... other oauth fields ...
  scopes = [
+   "openid",
+   "profile",
  ]
}
```

**Affected Resources**:
1. `pingone_davinci_application.pingcli__applicationAMinimal`

**Root Cause**: Export not handling OAuth default scopes when none specified.

**Fix Location**: `cmd/davinci_to_hcl.go` - application OAuth export logic

---

### CRITICAL - Category 3: Flow Color Field Removal
**Priority**: P1 - Non-functional but causes changes on 7 flows
**Affected Resources**: 7 `pingone_davinci_flow` resources

**Issue**: Export includes `color` field, but provider reads it as null/empty on import (likely deprecated or not persisted).

**Changes Required**:
```hcl
- color = "#CACED3" -> null
- color = "#ff661c" -> null  
- color = "#FFC8C1" -> null
```

**Affected Resources**:
1. `pingone_davinci_flow.pingcli__PingOne-0020-DaVinci-0020-API-0020-Protect-0020-Example`
2. `pingone_davinci_flow.pingcli__PingOne-0020-DaVinci-0020-API-0020-Protect-0020-Example_2`
3. `pingone_davinci_flow.pingcli__PingOne-0020-DaVinci-0020-API-0020-Protect-0020-Example_3`
4. `pingone_davinci_flow.pingcli__PingOne-0020-DaVinci-0020-API-0020-Protect-0020-Example_4`
5. `pingone_davinci_flow.pingcli__PingOne-0020-Sign-0020-On-0020-with-0020-Registration-002C--0020-Password-0020-Reset-0020-and-0020-Recovery`
6. `pingone_davinci_flow.pingcli__PingOne-0020-Sign-0020-On-0020-with-0020-Sessions`
7. `pingone_davinci_flow.pingcli__happypath`

**Root Cause**: Export includes attribute not maintained by provider/API.

**Fix Location**: `cmd/davinci_to_hcl.go` - flow export logic - suppress color field

---

### HIGH - Category 4: Flow Computed Attributes
**Priority**: P1 - Causes changes on all 7 flows with updates
**Affected Resources**: 7 `pingone_davinci_flow` resources (same as Category 3)

**Issue**: Multiple computed attributes trigger changes on every apply:
- `connectors` - array becomes computed after import
- `current_version` - computed, changes to `(known after apply)`
- `enabled` - becomes computed
- `graph_data.elements.nodes` - marked sensitive/computed
- `graph_data` itself partially computed
- `published_version` - becomes `(known after apply)`

**Changes Pattern**:
```hcl
~ connectors = [...] -> (known after apply)
~ current_version = 1 -> (known after apply)
~ enabled = true -> (known after apply)
~ graph_data = {
    ~ elements = {
        ~ nodes = (sensitive value)
      }
  }
~ published_version = 1 -> (known after apply)
```

**Root Cause**: These are computed attributes that should not be set in config. Export includes them when it should rely on provider computation.

**Fix Location**: `cmd/davinci_to_hcl.go` - flow export logic
- Remove `connectors` from export (computed by provider from graph_data)
- Remove `current_version` from export (computed)
- Remove `enabled` from export if always true (or make optional)
- Suppress `published_version` (always computed)
- Review graph_data.elements.nodes handling

---

### HIGH - Category 5: Connector Instance Properties Changes
**Priority**: P1 - Sensitive field changes on 4 connectors
**Affected Resources**: 4 `pingone_davinci_connector_instance` resources

**Issue**: Connector properties marked as sensitive show changes (likely credentials/secrets).

**Changes Required**:
```hcl
~ properties = (sensitive value)
```

**Affected Resources**:
1. `pingone_davinci_connector_instance.pingcli__OIDC-0020--0026--0020-OAuth-0020-IdP`
2. `pingone_davinci_connector_instance.pingcli__PingOne`
3. `pingone_davinci_connector_instance.pingcli__PingOne-0020-Protect`
4. `pingone_davinci_connector_instance.pingcli__SAML-0020-IdP`

**Root Cause**: Cannot export actual sensitive values. These changes are expected/unavoidable unless we:
- Support variable substitution for sensitive fields
- Document as known limitation
- Use lifecycle ignore_changes

**Fix Location**: Multiple options:
1. Add lifecycle blocks in export: `lifecycle { ignore_changes = [properties] }`
2. Support sensitive variable placeholders
3. Document as requiring manual intervention

---

### HIGH - Category 6: DaVinci Variable Issues (Two Sub-Issues)
**Priority**: P1 - Causes changes on 9 variables

#### 6A: Variable Mutable Field Incorrect Default
**Affected Resources**: 3 variables show mutable changing `false -> true`

**Issue**: Export FORCES `mutable = true` with override comment, but API actually returns `mutable = false` for these variables. Export must respect API value, not override it.

**Current Export Logic** (INCORRECT):
```hcl
mutable = true
# NOTE: mutable overridden to true because no value is provided (provider requirement)
```

**What API Actually Has**: `mutable = false`

**Terraform Plan Shows**:
```hcl
~ mutable = false -> true  # Export trying to change API value!
```

**Changes Required**:
```hcl
# Export should write:
mutable = false  # When API says false
# Not force to true
```

**Affected Resources**:
1. `pingone_davinci_variable.pingcli__companyBool` - API has `false`, export forces `true`
2. `pingone_davinci_variable.pingcli__companySecret` - API has `false`, export forces `true`
3. `pingone_davinci_variable.pingcli__userNumber` - API has `false`, export forces `true`

**Root Cause**: Export logic incorrectly assumes `mutable` must be `true` when no value provided. This is wrong - must export actual API value.

**Fix Location**: Export code that overrides mutable to true - remove this logic, export API value directly.

#### 6B: Variable Value Structure Changes
**Affected Resources**: 9 variables show value block changes

**Issue**: Variable `value` attribute structure not matching provider expectations. Two patterns observed:

**Pattern 1 - Value removed entirely** (boolean/number/basic types):
```hcl
- value = {
-   string = "false" -> null
- } -> null
```

**Pattern 2 - Value structure changed** (object types):
```hcl
~ value = {
    + key = "value"  # New structure
  }
```

**Affected Resources** (all 9 updated variables):
1. `pingone_davinci_variable.pingcli__companyBool` - value removed
2. `pingone_davinci_variable.pingcli__companyNumber` - value removed
3. `pingone_davinci_variable.pingcli__companyObject` - value structure change
4. `pingone_davinci_variable.pingcli__companySecret` - unknown (sensitive)
5. `pingone_davinci_variable.pingcli__flowBoolean` - value removed
6. `pingone_davinci_variable.pingcli__flowNumber` - value removed  
7. `pingone_davinci_variable.pingcli__flowObject` - value structure change
8. `pingone_davinci_variable.pingcli__userNumber` - value removed
9. `pingone_davinci_variable.pingcli__userObject` - value structure change

**Root Cause**: Variable value export logic incorrect. Possible issues:
- Exporting value when it should be omitted (for computed values)
- Wrong value structure for different data types
- Provider expects values managed outside resource definition

**Fix Location**: `cmd/davinci_to_hcl.go` - variable value attribute export logic

---

### MEDIUM - Category 7: Missing Connector Resource
**Priority**: P2 - Missing resource causes create operation
**Affected Resources**: 1 connector not exported

**Issue**: `User-0020-Pool` connector instance exists in environment but not exported.

**Changes Required**:
```hcl
+ resource "pingone_davinci_connector_instance" "pingcli__User-0020-Pool" {
+   connector = {
+     id = "skUserPool"
+   }
+   environment_id = "62f10a04-6c54-40c2-a97d-80a98522ff9a"
+   name = "User Pool"
+ }
```

**Root Cause**: Export logic missed this connector. Possible reasons:
- Connector not returned by export API
- Filtering logic incorrectly excluded it
- Connector added after export but before import test

**Fix Location**: `cmd/export.go` - connector instance enumeration logic

---

## Implementation Phases

### Phase 1: Critical Flow Policy & Application Fixes
**Target**: Fix P0 issues that affect multiple resources

1. **Fix flow policy trigger export**
   - Add logic to export default trigger block when not specified
   - Test with all 6 affected flow policies
   - Verify imported flow policies with explicit triggers still work

2. **Fix application OAuth scopes export**  
   - Export default scopes `["openid", "profile"]` when OAuth enabled but no scopes
   - Test with applicationAMinimal

**Success Criteria**: 6 flow policy + 1 application changes eliminated

---

### Phase 2: Flow Attribute Cleanup
**Target**: Fix flow computed attribute exports

1. **Remove color field from export**
   - Simply stop exporting color attribute
   - Verify flows still deploy correctly

2. **Remove computed flow attributes**
   - Remove: connectors, current_version, enabled, published_version
   - Keep only: name, description, graph_data, settings
   - Test flow recreation works

**Success Criteria**: All 7 flow resources show no changes

---

### Phase 3: Variable Investigation
**Target**: Understand and fix variable mutable changes

1. **Detailed diff analysis**
   - Get actual before/after values for mutable field
   - Check provider schema documentation
   - Determine correct export logic

2. **Implement fix based on findings**

**Success Criteria**: 9 variable resources show no changes

---

### Phase 4: Connector Properties Strategy
**Target**: Handle sensitive connector properties

1. **Decision point**: Choose approach:
   - A) Lifecycle ignore_changes blocks
   - B) Variable substitution system
   - C) Document as limitation

2. **Implement chosen approach**

**Success Criteria**: Documented strategy for sensitive values

---

### Phase 5: Missing Resources Investigation
**Target**: Ensure complete export coverage

1. **Investigate User Pool connector**
   - Check export API responses
   - Verify connector exists at export time
   - Fix enumeration logic if needed

2. **Add coverage tests**
   - Compare exported vs actual resources
   - Alert on missing resources

**Success Criteria**: All environment resources exported

---

## Validation Process

After each phase:

1. **Export** environment
2. **Import** all resources to new state file
3. **Plan** and verify zero changes
4. **Document** any remaining issues
5. **Iterate** on fix or document as limitation

---

## Known Acceptable Changes

These changes may be acceptable and documented as limitations:

1. **Sensitive connector properties** - Cannot export actual secrets
2. **API-managed computed values** - Some fields always computed by provider
3. **Flow version increments** - May change if flow modified

---

## Success Metrics

**Target**: 95%+ resources with zero changes after import

Current: 44% clean (18/41)  
After Phase 1: ~65% clean (27/41)  
After Phase 2: ~82% clean (34/41)  
After Phase 3: ~100% clean (41/41 excluding sensitive properties)

---

## Notes

- Plan output is from `/tmp/plan-output-full.txt`
- Test environment: `62f10a04-6c54-40c2-a97d-80a98522ff9a`
- Using provider development overrides
- Some changes may be provider bugs requiring upstream fixes

---

## Summary for Implementation

### Current State
- **41 total resources** in exported environment
- **18 resources (44%)** import cleanly with no changes
- **22 resources (54%)** require updates after import
- **1 resource (2%)** missing from export entirely

### Critical Findings

1. **Flow Policy Trigger Block** - 6 flow policies missing default trigger configuration
2. **Application OAuth Scopes** - 1 application missing default scopes when OAuth enabled  
3. **Variable Mutable Override Bug** - Export forces `mutable=true` but should use API value
4. **Variable Value Export** - Value blocks exported incorrectly for all variable types
5. **Flow Computed Attributes** - Exporting fields that should be provider-computed
6. **Flow Color Field** - Exporting deprecated/unsupported field
7. **Connector Properties** - Sensitive fields expected to change (may require lifecycle rules)
8. **Missing User Pool Connector** - One connector not exported

### Implementation Approach

Organized into 5 phases prioritized by impact:
- **P0 (Critical)**: Breaks functionality or affects multiple resources
- **P1 (High)**: Causes unnecessary changes but not breaking
- **P2 (Medium)**: Edge cases or single resources

Each phase documented with:
- Specific resources affected
- Root cause analysis
- Code location for fix
- Success criteria
- Test validation steps

### Target Outcome

Achieve **95%+ clean import rate** (zero changes after export→import cycle for non-sensitive resources).
