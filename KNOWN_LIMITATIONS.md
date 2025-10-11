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

## 1. toSnakeCase() Handles Consecutive Capitals Incorrectly

**Priority**: Medium  
**Category**: Code Quality / Readability

### Description

The `toSnakeCase()` function in `flow_converter.go` (line 549) produces suboptimal output for identifiers with consecutive capital letters, particularly common in connector IDs.

### Examples

| Input | Current Output | Expected Output |
|-------|---------------|-----------------|
| `pingOneSSOConnector` | `ping_one_s_s_o_connector` | `pingone_sso_connector` |
| `httpConnector` | `http_connector` | `http_connector` ✓ |
| `PingOneAPIConnector` | `ping_one_a_p_i_connector` | `pingone_api_connector` |

### Impact

- **Functional**: None (Terraform accepts both formats)
- **Readability**: Resource names and references have extra underscores
- **Consistency**: Inconsistent with common Terraform naming conventions

### Current Workaround

Tests use flexible matching that accepts both the current and expected formats. Connection ID references work correctly despite the formatting issue.

### Suggested Fix

Improve the snake_case conversion algorithm to:
1. Detect consecutive uppercase letters (acronyms)
2. Keep acronyms together as single words
3. Convert to lowercase as units

**Example Implementation Strategy**:
```go
// Detect runs of uppercase letters and treat as acronyms
// "pingOneSSOConnector" -> ["ping", "One", "SSO", "Connector"]
// -> ["ping", "one", "sso", "connector"]
// -> "ping_one_sso_connector"
```

**Estimated Effort**: 30-60 minutes

---

## 2. Complex Objects in Settings Use Map String Format

**Priority**: High  
**Category**: Correctness / Terraform Compatibility

### Description

When flow settings contain complex arrays of objects (e.g., `jsLinks`), the converter outputs them as Go map string representations instead of proper `jsonencode()` format.

### Examples

**Current Output**:
```hcl
js_links = ["map[crossorigin:anonymous defer:true integrity:sha256-abc123 label:jQuery referrerpolicy:no-referrer type:text/javascript value:https://code.jquery.com/jquery-3.6.0.min.js]"]
```

**Expected Output**:
```hcl
js_links = jsonencode([
  {
    "label": "jQuery",
    "value": "https://code.jquery.com/jquery-3.6.0.min.js",
    "defer": true,
    "crossorigin": "anonymous",
    "integrity": "sha256-abc123",
    "referrerpolicy": "no-referrer",
    "type": "text/javascript"
  }
])
```

### Impact

- **Functional**: High - Map string format may not be parseable by Terraform
- **Correctness**: Settings with complex objects may fail terraform validation
- **User Experience**: Generated Terraform code is not immediately usable

### Current Workaround

Tests check only for the presence of `js_links` field, not its content format. Manual editing required for flows using complex jsLinks.

### Affected Settings Fields

- `jsLinks` - Array of script configuration objects
- Potentially other complex nested objects in settings

### Suggested Fix

1. Detect when a settings value is a complex object (array of objects, nested objects)
2. Use `jsonencode()` formatting instead of direct string conversion
3. Apply consistent formatting (compact or pretty-printed)

**Estimated Effort**: 2-3 hours

---

## 3. Flow Variables Not Implemented

**Priority**: High  
**Category**: Missing Feature

### Description

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

### Current Decision

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

### Impact

- **Functional**: None
- **Aesthetics**: Minor - alignment looks inconsistent in some cases

### Current Decision

Using natural alignment (align to longest field name in block). This is the most common Terraform formatting style.

**No action needed** unless specific style guide requirements emerge.

---

## Summary Table

| # | Limitation | Priority | Impact | Effort | Status |
|---|-----------|----------|--------|--------|--------|
| 1 | toSnakeCase consecutive capitals | Medium | Low | 30-60m | Documented |
| 2 | Complex objects in settings | High | High | 2-3h | Documented |
| 3 | Variables not implemented | High | High | 4-6h | Documented |
| 4 | jsonencode compact format | Low | Low | 2-3h | Documented, Decision Made |
| 5 | Settings alignment | Very Low | None | N/A | Documented, No Action Needed |

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
