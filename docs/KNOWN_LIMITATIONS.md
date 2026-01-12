# Known Limitations

This document tracks known limitations and technical debt in the DaVinci Terraform Converter that are acceptable for current functionality but may need addressing in future iterations.

**Last Updated**: 2025-10-11  
**Project Phase**: Part 2 Phase 2.1 Complete (Resource Name Sanitization Updated)

---

## Recent Changes

### Resource Name Sanitization Now Uses pingcli Format

**Date**: 2025-10-11  
**Category**: Compatibility Enhancement

The converter now uses the same resource name sanitization as pingcli's `ImportBlock.Sanitize()` method to ensure consistency between the converter (intended as a pingcli plugin) and pingcli's export functionality.

**Changes**:
- Resource names are now prefixed with `pingcli__`
- Special characters (spaces, punctuation, etc.) are hex-encoded (e.g., space becomes `-0020-`)
- Alphanumeric characters, underscores, and hyphens are preserved

**Examples**:
| Input | Old Format | New Format (pingcli-compatible) |
|-------|-----------|--------------------------------|
| `Simple Test Flow` | `simple_test_flow` | `pingcli__Simple-0020-Test-0020-Flow` |
| `My-Flow@2024!` | `my_flow_2024` | `pingcli__My-Flow-0040-2024-0021-` |
| `Customer HTML Form (PF)` | `customer_html_form_pf` | `pingcli__Customer-0020-HTML-0020-Form-0020--0028-PF-0029-` |

**Rationale**: As this tool is intended to be integrated as a pingcli plugin, maintaining consistent resource naming conventions across both tools is critical for a seamless user experience.

---

## 1. Legacy Tests Need Resource Name Updates

**Priority**: Medium  
**Category**: Test Maintenance

### Context and Cause (Import Drift)

Several tests in `converter_test.go` still use old resource name expectations that don't match the updated pingcli-compatible sanitization format.

### Examples

| Test | Current Expectation | Should Be |
|------|-------------------|-----------|
| `TestSimpleFlowConversion` | `simple_test_flow` | `pingcli__Simple-0020-Test-0020-Flow` |
| `TestFlowWithSingleNode` | `http_connector_conn-123-abc` | `httpconnector_conn-123-abc` |
| `TestMultiFlowExport` | `main_flow`, `subflow_one` | `pingcli__Main-0020-Flow`, `pingcli__Subflow-0020-One` |

### Impact

- **Functional**: None - implementation is correct
- **Test Coverage**: Tests fail due to outdated expectations, not implementation bugs
- **Consistency**: All flow_comprehensive_test.go tests updated and passing

### Current Status

- ✅ `flow_comprehensive_test.go` - All tests updated and passing
- ✅ `flow_converter_test.go` - `toSnakeCase` tests updated and passing
- ❌ `converter_test.go` - Legacy tests need updating (7 tests failing)

### Suggested Fix

Update test expectations in `converter_test.go` to use:
1. pingcli-compatible resource names with hex encoding
2. Lowercase connector IDs without underscores in camelCase portions
3. Connection references like `httpconnector_<connectionId>` not `http_connector_<connectionId>`

**Estimated Effort**: 1-2 hours

---

## 2. toSnakeCase() Function Name Misleading

**Priority**: Low  
**Category**: Code Quality / Naming

### Description

The `toSnakeCase()` function in `flow_converter.go` doesn't actually convert to snake_case. It lowercases the input and removes non-alphanumeric characters (except underscores).

### Context and Cause

- `httpConnector` → `httpconnector` (not `http_connector`)
- `pingOneSSOConnector` → `pingonessoconnector` (not `ping_one_sso_connector`)
### Decision: Lifecycle Ignore (Flows)

  
  ```hcl

- **Functional**: None (works correctly for its purpose)
- **Readability**: Function name is misleading
- **Maintainability**: Future developers may expect actual snake_case conversion

### Suggested Fix

Rename function to better reflect what it does:
- `toLowerAlphaNum()` - more descriptive
- `sanitizeConnectorId()` - purpose-oriented
- Document clearly that it doesn't insert underscores between camelCase words

### Effects (Import Drift)

---

## 3. Flow Variables Not Implemented

### Context (Graph Data)

## 3. Flow Variables Not Implemented

### Decision (Graph Data)
**Category**: Missing Feature
  
  ```hcl
---

## 3. Flow Variables Not Implemented

**Priority**: High  
**Category**: Missing Feature

### Description
### Effects (Graph Data)
DaVinci flows can contain variables (flow-scoped, company-scoped, etc.) that are not currently converted to Terraform format.

### Impact

- **Functional**: High - Flows with variables will not work without manual intervention
- **Completeness**: Missing critical feature for many flows
- **User Experience**: Users must manually handle variables

### Current Workaround

Variables are completely ignored. Some old test expectations included variable comments but these have been removed.

### Example Flow Variable

**Input JSON**:
```json
{
  "variables": [
    {
      "name": "userId",
      "context": "flowInstance",
      "dataType": "string",
      "mutable": true,
      "value": "",
      "displayName": "User ID"
    }
  ]
}
```

**Possible Output Options**:

**Option A - Separate Resources** (preferred for Terraform modules):
```hcl
resource "pingone_davinci_variable" "userid" {
  environment_id = var.environment_id
  flow_id        = pingone_davinci_flow.my_flow.id
  name           = "userId"
  context        = "flowInstance"
  type           = "string"
  mutable        = true
  description    = "User ID"
}
```

**Option B - Comments** (simpler but manual):
```hcl
# Variables (must be created separately):
# - userId (flowInstance, string) - User ID
# - apiKey (company, secret) - API key for integration
```

### Decision Required

Need to determine:
1. Should variables be separate resources or comments? Variables will be separate resources, and will be mapped as dependencies in the flow resource or kept hard coded if it is not a complete environment export. 
2. How to handle variable references within flow properties? These should be left as the davinci variable syntax `{{variableName}}`, this syntax would not break the configuration as code principles because it is evaluated by davinci. 
3. How to represent sensitive variables? The value field in variables is marked as sensitive and would not show up in a terraform plan output. 

**Estimated Effort**: 4-6 hours (design + implementation)

---

## 4. jsonencode Uses Compact Format

**Priority**: Low  
**Category**: Code Readability

### Description

Properties that use `jsonencode()` output compact JSON with no formatting or indentation.

### Examples

**Current Output**:
```hcl
properties = jsonencode({"matchAttributes":{"value":["email"]},"userIdentifierForFindUser":{"value":"[{\"children\":[{\"text\":\"\"}]}]"}})
```

**Possible Alternative**:
```hcl
properties = jsonencode({
  "matchAttributes" : {
    "value" : ["email"]
  },
  "userIdentifierForFindUser" : {
    "value" : "[{\"children\":[{\"text\":\"\"}]}]"
  }
})
```

### Impact

- **Functional**: None (both formats are valid Terraform)
- **Readability**: Compact format is harder to read for debugging
- **Maintainability**: If generated code needs manual editing, pretty format helps

### Trade-offs

**Compact Format (Current)**:
- ✅ Smaller file size
- ✅ No multi-line formatting to maintain
- ✅ Less prone to whitespace issues
- ❌ Difficult to read
- ❌ Hard to diff in version control

**Pretty Format**:
- ✅ Human-readable
- ✅ Better for debugging
- ✅ Easier to diff
- ❌ Larger files
- ❌ More complex generation logic
- ❌ Alignment/indentation complexity

### Decision: Lifecycle Ignore

Using compact format based on assumption that generated Terraform code is not typically hand-edited. If this assumption changes, pretty-printing should be reconsidered.

**Estimated Effort**: 2-3 hours (if needed)

---

## 5. Settings Field Alignment Varies

**Priority**: Very Low  
**Category**: Formatting / Aesthetics

### Description

Settings fields use varying column alignment for the `=` operator depending on field name lengths.

### Example

**Current Output**:
```hcl
settings = {
  csp                                  = "value"
  intermediate_loading_screen_css      = ""
  log_level                            = 2
}
```

Some tests expected 2-space alignment, converter uses natural alignment based on longest field name.

### Effects

- **Functional**: None
- **Aesthetics**: Minor - alignment looks inconsistent in some cases

### Current Decision

Using natural alignment (align to longest field name in block). This is the most common Terraform formatting style.

**No action needed** unless specific style guide requirements emerge.

---

## Summary Table

| # | Limitation | Priority | Impact | Effort | Status |
|---|-----------|----------|--------|--------|--------|
| 1 | Legacy tests need resource name updates | Medium | None | 1-2h | Documented |
| 2 | toSnakeCase function name misleading | Low | None | 15-30m | Documented |
| 3 | Variables not implemented | High | High | 4-6h | Documented |
| 4 | jsonencode compact format | Low | Low | 2-3h | Documented, Decision Made |
| 5 | Settings alignment varies | Very Low | None | N/A | Documented, No Action Needed |


## Provider-Managed Fields and Import Drift

**Priority**: Medium  
**Category**: Provider Behavior / Import Drift

### Context

After importing existing DaVinci flows, Terraform can report in-place updates on fields that the PingOne provider or backend manages. Typical attributes: `connectors`, `current_version`, `enabled`, `published_version`.

### Cause

- **Computed/Server-managed**: These values are determined by the service (e.g., publish/version counters, enabled flags, connector sets derived from `graphData`). They may change outside Terraform.
- **State vs Config Mismatch**: Post-import, state reflects provider-managed values. Configuration may omit these or present differently, yielding no-op diffs.

### Decision

- The converter appends a lifecycle block on every `pingone_davinci_flow` resource:

  ```hcl
  lifecycle {
    ignore_changes = [
      connectors,
      current_version,
      enabled,
      published_version
    ]
  }
  ```

- Terraform may emit "Redundant ignore_changes" warnings because some attributes are purely computed. This is acceptable; the block suppresses plan churn after import.

### Effects

- **Functional**: Prevents noisy diffs and in-place updates on provider-managed fields.
- **Diagnostics**: Users may see validation warnings; safe to ignore.

### Notes

- If organizational policy requires eliminating warnings, remove purely computed attributes from `ignore_changes`. The converter keeps the full set to maximize drift suppression post-import.

---

## Empty `graph_data.data` Inclusion

**Priority**: Low  
**Category**: Compatibility / Diff Suppression

### Description

When the API returns `"graphData": { "data": {} }`, the provider persists that empty object. Omitting it from HCL causes diffs during plan/apply.

### Current Decision

- The converter always includes the empty data object as:

  ```hcl
  graph_data = {
    data = jsonencode({})
    # ...
  }
  ```

### Impact

- **Functional**: Aligns generated HCL with provider state; avoids false-positive changes.

## Resolved Limitations

### js_links Complex Object Support

**Resolved**: 2025-10-11  
**Category**: Correctness / Terraform Compatibility

Settings field `js_links` now outputs proper HCL object arrays instead of Go map string representations.

**Before**:
```hcl
js_links = ["map[crossorigin:anonymous defer:true integrity:sha256-abc123 ...]"]
```

**After**:
```hcl
js_links = [
  {
    crossorigin    = "anonymous"
    defer          = true
    integrity      = "sha256-abc123def456"
    label          = "jQuery"
    referrerpolicy = "no-referrer"
    type           = "text/javascript"
    value          = "https://code.jquery.com/jquery-3.6.0.min.js"
  }
]
```

**Implementation**: Added special case handling in `writeSettingsBlock()` to properly format `js_links` as an array of HCL objects with all required fields.

---

## Contributing

When addressing these limitations:

1. Update this document with the fix details and date
2. Move resolved items to a "Resolved Limitations" section
3. Update relevant test expectations
4. Document any new limitations discovered during the fix

---

## Related Documents

- [PART2_PHASE2.1_PROGRESS.md](./PART2_PHASE2.1_PROGRESS.md) - Current phase progress
- [PART2_MULTIFLOW_EXPANSION.md](./PART2_MULTIFLOW_EXPANSION.md) - Multi-flow expansion details
- [README.md](./README.md) - Project overview
