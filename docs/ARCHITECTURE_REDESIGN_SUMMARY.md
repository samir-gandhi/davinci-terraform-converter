# Architecture Redesign - Summary & Next Steps

## What Was Created

Two comprehensive design documents:

1. **GENERIC_ARCHITECTURE_REDESIGN.md** - Full generic architecture proposal
2. **GENERIC_ARCHITECTURE_REDESIGN_ADDENDUM.md** - SDK/Schema integration details

## Answers to Your Questions

### 1. Should ResourceHandler include API Schema and Terraform Schema?

**YES - Through Reference and Mapping Layer**

```go
type ResourceHandler[T any] struct {
    Schema           *ResourceSchema      // Our mapping schema
    TerraformSchema  *TFProviderSchema    // TF provider schema (read-only reference)
    APISchema        *OpenAPISchema       // API schema (read-only reference)
    
    APIClient        APIClientInterface   // Wraps pingone-go-client
    Graph            *DependencyGraph
    ImportGen        *ImportBlockGenerator
}
```

**Why**: 
- Validate our mappings against both authoritative schemas
- Generate mappings automatically from schemas
- Detect drift when schemas change
- Single source of truth for each layer

### 2. Does This Account for pingone-go-client SDK?

**YES - Explicitly Designed Around It**

The SDK is the foundation:

```go
// SDK type IS the generic parameter T
type ResourceHandler[pingone.DaVinciVariable] struct {
    // ...
}

// Schema knows SDK structure
type ResourceSchema struct {
    SDKType          reflect.Type              // reflect.TypeOf(pingone.DaVinciVariable{})
    SDKListMethod    string                    // "DaVinciAPI.EnvironmentVariablesGet"
    
    // Map SDK fields directly to TF attributes
    AttributeMappings []AttributeMapping
}

// No manual JSON conversion - work directly with SDK types
ListOperation: func(ctx context.Context, client *api.Client, envID string) ([]pingone.DaVinciVariable, error) {
    return client.DaVinciAPI.EnvironmentVariablesGet(ctx, envID)
}
```

**Benefits**:
- Type safety: Compiler catches SDK changes
- No manual conversion step
- Direct SDK method invocation
- Reflection for dynamic field access

### 3. Information Needed About Code Generation

**Critical for Design Refinement**

Please share:

#### A. pingone-go-client SDK Generation

```
❓ Questions:
1. What tool generates the SDK? (OpenAPI Generator? Custom?)
2. Where is the OpenAPI spec file?
3. Example of generated struct (DaVinciVariable)?
4. Any manual post-processing?
5. How often does SDK change?
```

#### B. Terraform Provider Schema Generation

```
❓ Questions:
1. What generates provider schemas?
2. Framework used? (Plugin SDK v2? Plugin Framework?)
3. Where are schemas defined? (resource_*.go files?)
4. Example schema definition?
5. What requires manual customization?
6. How do API field names map to TF attribute names?
```

#### C. Current Mapping Patterns

```
❓ Questions:
1. Examples of complex API → TF mappings?
2. Any 1:N mappings (1 API field → N TF attributes)?
3. Most time-consuming part of adding a resource?
4. What breaks when API changes?
5. Any particularly complex resources?
```

## Optimal Architecture (If Both Are Generated)

```text
┌─────────────────────────────────────────────────────────┐
│              OpenAPI Specification                      │
│           (Single Source of Truth)                      │
└─────────────────────────────────────────────────────────┘
                │                        │
                ↓                        ↓
    ┌──────────────────────┐  ┌──────────────────────┐
    │  SDK Generation      │  │  TF Schema Gen       │
    │  (pingone-go-client) │  │  (TF Provider)       │
    └──────────────────────┘  └──────────────────────┘
                │                        │
                ↓                        ↓
    ┌──────────────────────┐  ┌──────────────────────┐
    │  SDK Types           │  │  TF Schema           │
    │  + Overrides         │  │  + Overrides         │
    └──────────────────────┘  └──────────────────────┘
                │                        │
                └───────────┬────────────┘
                            ↓
                ┌──────────────────────────┐
                │  CONVERTER MAPPING GEN   │
                │  (Our Code Generation)   │
                │                          │
                │  - ResourceSchema        │
                │  - AttributeMappings     │
                │  - Validation Rules      │
                │  + Manual Overrides      │
                └──────────────────────────┘
                            ↓
                ┌──────────────────────────┐
                │  Generic Converter       │
                │  (Runtime - Zero Changes)│
                └──────────────────────────┘
```

**Result**: Triple code generation means adding a new resource requires:
1. ~30 lines of generated mapping code
2. ~10 lines of manual overrides (if needed)
3. **Zero manual converter logic**

## Key Design Principles

### 1. Three-Layer Schema Architecture

```
API Schema (OpenAPI)     ← Authoritative for API
      ↓
SDK Types (Go Structs)   ← Generated from API schema
      ↓
Mapping Schema (Ours)    ← Maps SDK → TF (mostly generated)
      ↓
TF Schema (Provider)     ← Authoritative for Terraform
```

### 2. Direct SDK Integration

- No intermediate JSON representation
- Work directly with `pingone.DaVinciVariable`, etc.
- Type-safe with generics: `ResourceHandler[T]`
- Compiler catches SDK changes

### 3. Mapping Intelligence

The converter's value is in the **mapping layer**:

```go
type AttributeMapping struct {
    SDKFieldPath:    "Value"                    // In SDK struct
    TFAttributePath: "value"                    // In TF resource
    Transform:       stringValue                // How to convert
    VariableEligible: true                      // Can be module variable
}
```

Most mappings can be generated; manual overrides for edge cases.

### 4. Code Generation Workflow

```bash
# When API changes:
1. OpenAPI spec updated
2. SDK regenerated (pingone-go-client)
3. TF provider schemas regenerated
4. OUR TOOL: Regenerate converter mappings
5. Manual review of override points
6. Done - new resource ready
```

## Implementation Strategy

### Phase 0: Schema Analysis (NEW)

**Before starting implementation**, we need to:

1. Analyze SDK generation process
2. Analyze TF provider schema generation  
3. Document current mapping patterns
4. Design code generation templates
5. Build proof-of-concept generator

**Deliverable**: Can generate 1 complete resource mapping that matches current manual code

### Phase 1-7: As Originally Planned

But with code generation integrated from the start.

## Benefits Summary

| Aspect | Current | With Redesign | With Schema Gen |
|--------|---------|---------------|-----------------|
| Code per resource | 500 lines | 80 lines | 30-40 lines |
| Manual work | 4-8 hours | 1 hour | 15 minutes |
| Generation | None | Partial | 90% automated |
| Consistency | Manual | Schema-driven | Schema-derived |
| API changes | Manual updates | Remap once | Regenerate |
| Bug fixes | N places | 1 place | 1 place |

## Risk Analysis

### If We Have Access to Schemas

**LOW RISK**: Most mappings can be automated

### If Schemas Are Complex/Unavailable

**MEDIUM RISK**: More manual work, but still huge improvement over current

### Mitigation

Start with Phase 0 analysis to determine:
- How much can be automated?
- What requires manual mapping?
- Is schema-driven generation worth it?

If generation isn't feasible, fall back to manual schemas (still 84% code reduction).

## Next Steps

### Immediate (You)

Share information about:
1. SDK generation process and source schemas
2. TF provider schema generation
3. Complex mapping examples
4. Pain points when adding resources

### Short Term (Me)

After receiving info:
1. Refine design based on actual schemas
2. Create schema analysis document
3. Build proof-of-concept generator
4. Validate generated code matches manual code

### Medium Term (Both)

1. Review generated code together
2. Identify manual override points
3. Decide: Full code gen or manual schemas?
4. Plan implementation phases

## Questions?

Key decision point: **How much of the mapping can we automate?**

The answer depends on:
- Complexity of API → TF transformations
- Availability and quality of schemas
- Patterns in current manual mappings

Once we analyze the schemas, we'll know if we can achieve:
- **90% automation** (ideal)
- **50-70% automation** (still valuable)
- **Manual schemas only** (fallback, still 84% reduction)

All three options are significant improvements. Schema analysis determines which is achievable.
