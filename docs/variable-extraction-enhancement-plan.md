# Variable Extraction Enhancement Plan

## Current State

### Implementation Status
The variable extraction feature (Phase 2 of Feature 1) is partially complete:

**Completed:**
- Generic `VariableEligibleAttribute` architecture for any resource type
- `AttributeExtractionContext` with configurable extraction rules
- DaVinci variable extraction (extracts 'value' attribute)
- Connector property extraction using hardcoded pattern matching (~20 common properties)
- HCL generation with `var.{name}` references for both resource types
- Unit tests covering extraction and HCL generation (91 tests passing)
- ExportedData structure updated with ExtractedVariables field

**In Progress:**
- Integration into module export flow (structure ready, needs wiring)

**Not Started:**
- End-to-end integration tests

### Current Property Extraction Approach

**Method:** Hardcoded pattern matching with two static maps:

```go
variableEligibleProperties := map[string]bool{
    "baseUrl":     true,
    "clientId":    true,
    "tenantId":    true,
    "domain":      true,
    "region":      true,
    // ~15 more common properties
}

secretProperties := map[string]bool{
    "clientSecret": true,
    "apiKey":       true,
    "password":     true,
    "privateKey":   true,
    // ~3 more secret patterns
}
```

**Coverage:**
- Captures common configuration properties (URLs, IDs, domains)
- Identifies obvious secrets (clientSecret, apiKey, password)
- Works for ~70-80% of typical connector configurations

**Limitations:**
- Property name variations missed (e.g., "baseURL" vs "baseUrl")
- Connector-specific properties not captured
- Cannot leverage DaVinci connector schema metadata
- Manual maintenance required for new property patterns
- No connector-specific logic

## Gap Analysis

### Available Resource: connector-schema.json

Located at: `dvtf-pingctl/internal/generate/connector_schema/connector-schema.json`

**Schema Structure:**
```json
[
  {
    "name": "Adobe Marketo",
    "connectorId": "adobemarketoConnector",
    "connectorCategories": [...],
    "properties": {
      "clientId": {
        "type": "string",
        "displayName": "Client ID",
        "preferredControlType": "textField",
        "info": "Your Adobe Marketo client ID."
      },
      "clientSecret": {
        "type": "string",
        "displayName": "Client Secret",
        "preferredControlType": "textField",
        "info": "Your Adobe Marketo client secret."
      },
      "endpoint": {
        "displayName": "API URL",
        "preferredControlType": "textField",
        "info": "The API endpoint for your Adobe Marketo instance"
      }
    }
  }
]
```

**Metadata Available:**
- 100+ connectors with complete property definitions
- `preferredControlType`: "textField", "passwordField", "toggleSwitch", etc.
- `displayName`: Human-readable property names
- `info`: Descriptions for variable documentation
- `type`: JSON Schema types (string, boolean, array, object)
- Property names as defined in DaVinci API

**Key Insights:**
- `preferredControlType: "passwordField"` → Automatic secret detection
- Property names match actual DaVinci API property keys
- Covers all property name variations per connector
- Provides rich metadata for variable descriptions

### What's Missing

1. **Schema Loading Infrastructure**
   - No JSON parser for connector-schema.json
   - No in-memory schema representation
   - No schema initialization on startup

2. **Connector ID → Property Metadata Lookup**
   - Cannot match `instance.Connector.ID` to schema `connectorId`
   - No property metadata retrieval by connector + property name

3. **Schema-Driven Classification**
   - Cannot use `preferredControlType` for secret detection
   - Cannot leverage property metadata for better descriptions
   - No connector-specific extraction rules

4. **Property Type Mapping**
   - Schema `type` field not mapped to Terraform types
   - No handling of complex types (arrays, objects) from schema

5. **Fallback Strategy**
   - No graceful degradation if schema unavailable
   - No hybrid approach (schema + pattern matching)

## Enhancement Plan

### Phase 1: Schema Integration Foundation

**Goal:** Load and query connector-schema.json at runtime

**Tasks:**
1. Copy connector-schema.json into converter project
   - Location: `internal/converter/data/connector-schema.json`
   - Embedded using go:embed directive

2. Create schema data structures
   ```go
   type ConnectorSchema struct {
       Name               string                      `json:"name"`
       ConnectorID        string                      `json:"connectorId"`
       ConnectorCategories []ConnectorCategory        `json:"connectorCategories"`
       Properties         map[string]PropertyMetadata `json:"properties"`
   }

   type PropertyMetadata struct {
       Type                  string `json:"type"`
       DisplayName           string `json:"displayName"`
       PreferredControlType  string `json:"preferredControlType"`
       Info                  string `json:"info"`
   }
   ```

3. Implement schema loader
   ```go
   // LoadConnectorSchemas loads and parses connector-schema.json
   func LoadConnectorSchemas() (map[string]ConnectorSchema, error)

   // GetConnectorSchema retrieves schema by connector ID
   func GetConnectorSchema(connectorID string) (*ConnectorSchema, bool)
   ```

4. Initialize schema on package init
   - Parse JSON once at startup
   - Store in package-level map for O(1) lookups
   - Log warnings if schema unavailable

**Estimated Effort:** 4-6 hours
**Priority:** High
**Dependencies:** None

### Phase 2: Schema-Driven Property Classification

**Goal:** Use schema metadata to identify variable-eligible properties

**Tasks:**
1. Enhance GetConnectorInstanceVariableEligibleAttributes
   - Look up connector schema by instance.Connector.ID
   - Iterate through instance.Properties
   - For each property, check schema metadata

2. Implement classification logic
   ```go
   func classifyProperty(
       propName string,
       propValue interface{},
       metadata *PropertyMetadata,
   ) PropertyClassification {
       // Classification result
       type PropertyClassification struct {
           IsVariableEligible bool
           IsSecret           bool
           VariableType       string
           Description        string
       }

       // Rules:
       // 1. preferredControlType == "passwordField" → secret variable
       // 2. Property name matches URL/ID/Key patterns → regular variable
       // 3. textField + config-like name → regular variable
       // 4. Other types → skip or extract based on context
   }
   ```

3. Use schema metadata for enrichment
   - `displayName` → Variable description prefix
   - `info` → Full variable description
   - `type` → Terraform type mapping

4. Implement fallback to pattern matching
   - If schema not found for connector
   - If property not in schema
   - Use existing hardcoded maps as backup

**Estimated Effort:** 6-8 hours
**Priority:** High
**Dependencies:** Phase 1

### Phase 3: Advanced Property Handling

**Goal:** Handle complex property types and nested values

**Tasks:**
1. Support array properties
   - Extract array items that are variable-eligible
   - Example: customAuth arrays with URLs/tokens

2. Support object properties
   - Recurse into nested objects
   - Extract nested configuration values

3. Handle conditional properties
   - Properties that depend on other property values
   - Connector-specific extraction rules

4. Add property whitelisting/blacklisting
   - Configuration to include/exclude specific properties
   - Per-connector override rules

**Estimated Effort:** 8-10 hours
**Priority:** Medium
**Dependencies:** Phase 2

### Phase 4: Testing & Validation

**Goal:** Comprehensive test coverage for schema-driven extraction

**Tasks:**
1. Unit tests for schema loading
   - Test schema parsing
   - Test connector lookup
   - Test property metadata retrieval

2. Unit tests for classification
   - Test passwordField detection
   - Test pattern matching fallback
   - Test complex type handling

3. Integration tests
   - Export real connector instances
   - Verify all expected variables extracted
   - Validate HCL with var references
   - Check module.tf variable definitions

4. Regression tests
   - Ensure existing pattern matching still works
   - Verify backward compatibility
   - Test with schema unavailable

**Estimated Effort:** 6-8 hours
**Priority:** High
**Dependencies:** Phases 1-3

### Phase 5: Documentation & Refinement

**Goal:** Document enhancement and enable configurability

**Tasks:**
1. Update ARCHITECTURE.md
   - Document schema-driven extraction
   - Explain classification rules
   - Show fallback strategy

2. Add configuration options
   - Enable/disable schema-driven extraction
   - Configure property classification rules
   - Connector-specific overrides

3. Create migration guide
   - How to update existing code
   - When to use schema vs. patterns
   - How to add new classification rules

4. Performance optimization
   - Benchmark schema lookup performance
   - Optimize property iteration
   - Consider caching strategies

**Estimated Effort:** 4-6 hours
**Priority:** Medium
**Dependencies:** Phases 1-4

## Implementation Strategy

### Recommended Approach

**Hybrid Strategy: Schema-First with Pattern Fallback**

```go
func GetConnectorInstanceVariableEligibleAttributes(
    instanceJSON []byte,
    resourceName string,
) ([]VariableEligibleAttribute, error) {
    // Parse instance
    var instance ConnectorInstanceResponse
    // ... error handling ...

    // Try schema-driven extraction first
    if schema, found := GetConnectorSchema(instance.Connector.ID); found {
        return extractUsingSchema(instance, schema, resourceName)
    }

    // Fallback to pattern matching
    return extractUsingPatterns(instance, resourceName)
}
```

**Benefits:**
- Comprehensive coverage for known connectors
- Graceful degradation for unknown connectors
- Easy to test and maintain
- Clear separation of concerns

### Configuration Design

```go
type ExtractionConfig struct {
    // Use schema-driven extraction when available
    UseSchema bool `default:"true"`

    // Fallback to pattern matching if schema not found
    FallbackToPatterns bool `default:"true"`

    // Extract all textField properties as variables
    ExtractAllTextFields bool `default:"false"`

    // Connector-specific overrides
    ConnectorOverrides map[string]ConnectorExtractionConfig

    // Global property exclusions
    ExcludeProperties []string
}

type ConnectorExtractionConfig struct {
    // Include these properties even if not in schema
    IncludeProperties []string

    // Exclude these properties even if in schema
    ExcludeProperties []string

    // Custom classification rules
    CustomRules []PropertyClassificationRule
}
```

## Benefits of Enhancement

### Immediate Benefits
1. **Comprehensive Coverage**: Extract variables from all connector properties, not just common ones
2. **Automatic Secret Detection**: Leverage passwordField metadata instead of guessing
3. **Better Documentation**: Use schema displayName and info for variable descriptions
4. **Reduced Maintenance**: No manual updates needed for new connectors
5. **Consistency**: Same property names and types as DaVinci API

### Long-Term Benefits
1. **Scalability**: Handles new connectors automatically
2. **Accuracy**: Schema is source of truth, not guesswork
3. **Flexibility**: Easy to add connector-specific rules
4. **Testability**: Clear schema makes testing easier
5. **User Experience**: Better variable names and descriptions

## Migration Path

### For Existing Code
1. Phase 1 implementation is backward compatible
2. Pattern matching remains as fallback
3. No breaking changes to API
4. Gradual migration connector by connector

### For New Development
1. Start with schema-driven extraction
2. Add pattern fallback for edge cases
3. Test with real connector exports
4. Document any connector-specific quirks

## Success Metrics

### Coverage Metrics
- % of connector properties correctly classified
- % of secrets automatically detected
- % of connectors with full schema coverage

### Quality Metrics
- Variable description quality (using schema info)
- Terraform type accuracy
- Reduction in manual property additions

### Performance Metrics
- Schema load time
- Property lookup latency
- Memory usage for schema cache

## Next Steps

**Immediate Action (Complete Current Work):**
1. Wire variable extraction into module export flow
2. Add end-to-end integration tests
3. Verify all 91 tests still passing

**Next Enhancement (Schema Integration):**
1. Start with Phase 1: Schema loading infrastructure
2. Add unit tests for schema operations
3. Integrate into existing extraction logic
4. Document hybrid approach

**Timeline Estimate:**
- Current work completion: 4-6 hours
- Schema integration (Phases 1-2): 10-14 hours
- Advanced features (Phase 3): 8-10 hours
- Testing & documentation (Phases 4-5): 10-14 hours
- **Total estimated effort**: 32-44 hours (4-6 days)

## References

- connector-schema.json: `dvtf-pingctl/internal/generate/connector_schema/connector-schema.json`
- Current implementation: `internal/converter/connector_instance_converter.go`
- Variable types: `internal/converter/variable_eligible.go`
- Module export: `internal/exporter/module_export.go`
