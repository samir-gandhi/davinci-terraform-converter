# Architecture Redesign Addendum: Schema Generation & SDK Integration

## Questions Addressed

### Q1: Should ResourceHandler include API Schema and Terraform Schema?

**Short Answer**: Yes, but indirectly through schema mapping, not direct inclusion.

**Detailed Answer**:

The `ResourceHandler` should reference **both** schemas through a mapping layer:

```go
type ResourceHandler[T any] struct {
    Schema           *ResourceSchema      // Our converter schema (maps between API and TF)
    TerraformSchema  *TFProviderSchema    // Reference to TF provider schema (read-only)
    APISchema        *OpenAPISchema       // Reference to API schema (read-only)
    
    // Existing fields
    APIClient        APIClientInterface
    Graph            *DependencyGraph
    ImportGen        *ImportBlockGenerator
    Logger           Logger
}

// Schema mapping contains the intelligence to bridge API ↔ TF
type ResourceSchema struct {
    // Metadata
    TerraformType    string
    APIPath          string
    SDKType          reflect.Type
    
    // Schema references (read-only, for validation and generation)
    TFSchemaPath     string              // Path to TF provider schema file
    APISchemaPath    string              // Path to OpenAPI spec section
    
    // Mapping rules (the core intelligence)
    AttributeMappings []AttributeMapping  // How API fields map to TF attributes
    
    // Custom handlers for edge cases
    CustomMappers     map[string]MappingFunc
    PostProcessors    []ProcessorFunc
    
    // Everything else from original design...
}

// Mapping between API response field and TF attribute
type AttributeMapping struct {
    // API side (from pingone-go-client)
    APIFieldPath     string              // "Metadata.Name" or "Properties[0].Value"
    APIFieldType     reflect.Type        // Type in Go SDK
    
    // Terraform side (from provider schema)
    TFAttributePath  string              // "name" or "properties.value"
    TFAttributeType  string              // "string", "number", "list(object({...}))"
    
    // Transformation
    Transform        MappingFunc         // How to convert API → TF
    ReverseTransform MappingFunc         // How to convert TF → API (future: import)
    
    // Flags
    VariableEligible bool
    IsComputed       bool
    IsRequired       bool
    IsSensitive      bool
}
```

**Why This Approach?**

1. **Single Source of Truth**: TF provider schema is authoritative for TF attributes
2. **API Schema is Authoritative**: OpenAPI spec is authoritative for API structure
3. **Mapping is Our Intelligence**: We define how to bridge the gap
4. **Validation**: Can validate our mappings against both schemas
5. **Code Generation**: Can generate mappings from schemas + manual overrides

---

### Q2: Does This Account for pingone-go-client SDK?

**Short Answer**: Yes, explicitly designed for it.

**Detailed Answer**:

The SDK integration is central to the design:

#### Current SDK Usage Pattern

```go
// Current code (variable_exporter.go)
variables, err := client.ListVariables(ctx, client.EnvironmentID)
// variables is []pingone.DaVinciVariable (SDK type)

// Manual conversion to our JSON format
variableJSON, err := convertVariableToJSON(&variable)

// Then our converter takes over
hcl, err := converter.ConvertVariableWithOptions(variableJSON, skipDeps)
```

#### Proposed SDK Integration

```go
// SDK type is the generic parameter T
type ResourceHandler[T any] struct {
    Schema    *ResourceSchema
    APIClient APIClientInterface  // Wraps pingone-go-client
    // ...
}

// For variables: ResourceHandler[pingone.DaVinciVariable]
// For flows: ResourceHandler[pingone.DaVinciFlow]
// etc.

// Schema knows how to work with SDK types directly
type ResourceSchema struct {
    SDKType          reflect.Type              // reflect.TypeOf(pingone.DaVinciVariable{})
    
    // SDK-specific operations using the actual SDK client
    ListOperation    func(context.Context, *api.Client, string) ([]T, error)
    GetOperation     func(context.Context, *api.Client, string) (T, error)
    
    // Attribute mappings FROM SDK fields TO Terraform attributes
    AttributeMappings []AttributeMapping
}

// Example: Variable resource registration
func init() {
    GlobalRegistry.Register(&ResourceSchema{
        TerraformType: "pingone_davinci_variable",
        SDKType:       reflect.TypeOf(pingone.DaVinciVariable{}),
        
        // Direct SDK method reference
        ListOperation: func(ctx context.Context, client *api.Client, envID string) ([]pingone.DaVinciVariable, error) {
            // Call SDK directly - no manual conversion needed
            return client.DaVinciAPI.EnvironmentVariablesGet(ctx, envID)
        },
        
        // Mappings from SDK fields to TF attributes
        AttributeMappings: []AttributeMapping{
            {
                APIFieldPath:    "Name",                    // SDK field: variable.Name
                TFAttributePath: "name",                    // TF attr: name
                Transform:       passthrough,               // No transformation needed
            },
            {
                APIFieldPath:    "Value",                   // SDK field: variable.Value
                TFAttributePath: "value",                   // TF attr: value
                Transform:       stringValue,               // Convert interface{} to string
                VariableEligible: true,                     // Can become module variable
            },
            {
                APIFieldPath:    "Flow.Id",                 // SDK field: variable.Flow.Id
                TFAttributePath: "flow.id",                 // TF attr: flow.id
                Transform:       uuidToString,              // Convert UUID to string
                ReferencesType:  "pingone_davinci_flow",    // Dependency
            },
        },
    })
}
```

**Key Benefits**:

1. **No Manual JSON Conversion**: Work directly with SDK types
2. **Type Safety**: Generic `T` matches SDK type
3. **SDK Changes Tracked**: If SDK changes, compiler errors guide updates
4. **Reflection for Flexibility**: Can introspect SDK structs dynamically
5. **Manual Overrides**: Can specify custom mapping functions for edge cases

#### SDK Client Wrapper

```go
// Wrapper around pingone-go-client for resource operations
type APIClientInterface interface {
    // Generic list operation
    ListResources(ctx context.Context, resourceType string, envID string) ([]interface{}, error)
    
    // Generic get operation
    GetResource(ctx context.Context, resourceType string, envID, resourceID string) (interface{}, error)
    
    // Direct SDK access for custom operations
    DaVinciAPI() *pingone.DaVinciAPI
    PingOneAPI() *pingone.PingOneAPI
}

// Implementation wraps actual SDK client
type SDKClient struct {
    client *api.Client  // Your existing client wrapper
    registry *ResourceRegistry
}

func (c *SDKClient) ListResources(ctx context.Context, resourceType string, envID string) ([]interface{}, error) {
    schema, err := c.registry.GetSchema(resourceType)
    if err != nil {
        return nil, err
    }
    
    // Use schema's ListOperation which knows the SDK method
    return schema.ListOperation(ctx, c.client, envID)
}
```

---

### Q3: Information About SDK and Provider Code Generation

**Short Answer**: Yes, please share! This will dramatically improve the design.

**What Would Be Helpful**:

#### About pingone-go-client Generation

1. **Generator Tool**: What generates the SDK? (OpenAPI Generator, custom tool, manual?)
2. **Source Schema**: OpenAPI spec file location? Version?
3. **Generated Structure**:
   - Where are types generated? (`pingone/model_*.go`?)
   - Naming conventions? (`DaVinciVariable`, `DaVinciFlow`, etc.)
   - Response wrapper structure? (pagination, metadata, etc.)
4. **Custom Overrides**: Any manual post-processing after generation?
5. **Stability**: How often does the SDK change? Breaking vs non-breaking changes?

#### About Terraform Provider Generation

1. **Generator Tool**: What generates the provider schema?
2. **Schema Format**: 
   - Framework used? (Plugin SDK v2, Plugin Framework?)
   - Where are schemas defined? (`internal/service/davinci/resource_*.go`?)
   - Example schema structure?

3. **Manual Customizations**:
   - What requires manual post-processing?
   - Complex attribute handling (nested blocks, dynamic attributes)?
   - Custom validation functions?

4. **Attribute Naming**: 
   - Conventions: snake_case, camelCase?
   - How do API fields map to TF attributes? (automated or manual?)

#### Ideal Architecture with Generation

If both SDK and Provider are generated, optimal architecture:

```text
┌─────────────────────────────────────────────────────────────────┐
│                     OpenAPI Specification                        │
│                   (Single Source of Truth)                       │
└─────────────────────────────────────────────────────────────────┘
                    │                          │
                    ↓                          ↓
        ┌───────────────────────┐  ┌───────────────────────┐
        │  pingone-go-client    │  │  Terraform Provider   │
        │  SDK Generation       │  │  Schema Generation    │
        └───────────────────────┘  └───────────────────────┘
                    │                          │
                    ↓                          ↓
        ┌───────────────────────┐  ┌───────────────────────┐
        │  Generated SDK Types  │  │  Generated TF Schema  │
        │  + Manual Overrides   │  │  + Manual Overrides   │
        └───────────────────────┘  └───────────────────────┘
                    │                          │
                    └────────────┬─────────────┘
                                 ↓
                    ┌───────────────────────────┐
                    │  Converter Mapping Gen    │
                    │  (OUR CODE GENERATION)    │
                    │                           │
                    │  Generates:               │
                    │  - ResourceSchema         │
                    │  - AttributeMappings      │
                    │  - Validation Rules       │
                    │  + Manual Overrides       │
                    └───────────────────────────┘
                                 ↓
                    ┌───────────────────────────┐
                    │   Generic Converter       │
                    │   Runtime (No Changes)    │
                    └───────────────────────────┘
```

**Benefits**:

1. **Triple Code Generation**: OpenAPI → SDK + TF Schema → Converter Mappings
2. **Automatic Synchronization**: When API changes, all layers regenerate
3. **Manual Override Points**: Each layer can have customizations
4. **Validation**: Can detect API/TF schema drift automatically
5. **Zero Manual Coding**: Standard resources require zero manual code

---

## Updated ResourceSchema Design

Based on SDK integration, here's the refined schema:

```go
type ResourceSchema struct {
    // Identity
    TerraformType    string              // "pingone_davinci_variable"
    SDKType          reflect.Type        // reflect.TypeOf(pingone.DaVinciVariable{})
    
    // Schema references (for generation and validation)
    APISchemaRef     string              // "#/components/schemas/DaVinciVariable"
    TFSchemaPath     string              // Path in TF provider code
    
    // SDK Integration
    SDKListMethod    string              // "DaVinciAPI.EnvironmentVariablesGet"
    SDKGetMethod     string              // "DaVinciAPI.EnvironmentVariableGet"
    SDKTypeWrapper   bool                // Does SDK wrap response in envelope?
    
    // Attribute Mappings (GENERATED from schemas + manual overrides)
    Attributes       []AttributeMapping
    
    // Dependencies (can be auto-detected from TF schema)
    DependsOn        []DependencyRule
    
    // Variable extraction (can be generated from TF schema + flags)
    VariableRules    []VariableRule
    
    // Import configuration (can be generated from TF schema)
    ImportIDFormat   string
    
    // Manual overrides for complex cases
    CustomHCLGenerator    func(interface{}) (string, error)
    CustomValidator       func(interface{}) error
    PostProcessors        []ProcessorFunc
}

type AttributeMapping struct {
    // Source (SDK/API)
    SDKFieldPath     string              // "Flow.Id" - path in SDK struct
    SDKFieldType     reflect.Type        // Go type
    APISchemaRef     string              // "#/paths/.../parameters/flowId"
    
    // Target (Terraform)
    TFAttributePath  string              // "flow.id" - path in TF resource
    TFAttributeType  schema.ValueType    // From TF schema
    TFSchemaRef      string              // Reference in TF provider code
    
    // Transformation
    Transform        MappingFunc         // SDK value → HCL value
    DefaultTransform TransformType       // Enum: Passthrough, ToString, UUID, etc.
    
    // Metadata (can be generated from schemas)
    Required         bool
    Computed         bool
    Sensitive        bool
    VariableEligible bool
    
    // Dependencies (auto-detected from TF schema)
    ReferencesType   string              // Other resource type
    ReferenceField   string              // Field in other resource
}
```

---

## Code Generation Workflow

### Phase 1: Extract Schema Information

```go
// Tool: cmd/generate-mappings/main.go

// 1. Parse OpenAPI spec
apiSpec := parseOpenAPISpec("path/to/pingone-openapi.yaml")

// 2. Introspect SDK types (using reflection)
sdkTypes := introspectSDKTypes(pingone.DaVinciVariable{}, pingone.DaVinciFlow{}, ...)

// 3. Parse TF provider schemas
tfSchemas := parseTFProviderSchemas("path/to/terraform-provider-pingone/internal/service/davinci/")

// 4. Generate mappings
for resourceType := range tfSchemas {
    mapping := generateResourceMapping(
        apiSpec.GetResourceSchema(resourceType),
        sdkTypes.GetType(resourceType),
        tfSchemas.GetSchema(resourceType),
    )
    
    // 5. Write generated code
    writeGeneratedMapping(mapping, "internal/generated/resource_"+resourceType+".go")
}
```

### Phase 2: Manual Override Layer

```go
// File: internal/resources/variable_overrides.go

func init() {
    // Start with generated schema
    schema := generated.VariableResourceSchema()
    
    // Apply manual overrides for complex cases
    schema.Attributes["value"].Transform = customValueTransform
    schema.CustomHCLGenerator = generateVariableHCLWithSpecialHandling
    
    // Register the overridden schema
    GlobalRegistry.Register(schema)
}
```

---

## Questions for You

To refine this design further, please share:

1. **SDK Generation**:
   - Command/tool used to generate pingone-go-client?
   - Location of OpenAPI spec file?
   - Example of a generated SDK type (e.g., `DaVinciVariable` struct)?
   - Any post-generation scripts or manual changes?

2. **Terraform Provider Generation**:
   - Tool used to generate provider schemas?
   - Example of a resource schema definition file?
   - How are attribute names derived from API fields?
   - Location of schema files in provider repo?

3. **Mapping Complexity**:
   - Examples of complex mappings (API field → TF attribute)?
   - Are there cases where 1 API field → N TF attributes?
   - Are there computed attributes that require special handling?

4. **Current Pain Points**:
   - What's currently the most time-consuming part of adding a resource?
   - What breaks most often when API changes?
   - Any resources that are particularly complex?

With this information, I can refine the design to:

- **Auto-generate 90%+ of mappings** from schemas
- **Minimize manual code** to only true edge cases
- **Detect API/TF drift** automatically
- **Provide migration path** when schemas change

---

## Revised Implementation Phases

### Phase 0: Schema Analysis (NEW - Week 0.5)

```text
Tasks:
1. Analyze pingone-go-client generation process
2. Analyze Terraform provider schema generation
3. Document current mapping patterns (API → TF)
4. Identify which mappings can be automated
5. Design code generation templates

Deliverables:
- docs/SDK_GENERATION_ANALYSIS.md
- docs/TF_SCHEMA_ANALYSIS.md
- docs/MAPPING_PATTERNS.md
- Code generation plan

Success Criteria:
- Understand 100% of current generation workflows
- Identify all manual override points
- Can generate 1 resource mapping as proof of concept
```

### Phase 1: Foundation (Week 1)

- Add schema reference fields to ResourceSchema
- Add SDK integration layer
- Add attribute mapping system
- Build reflection utilities for SDK types

### Phase 2: Code Generation (Week 2)

- Build mapping generator from schemas
- Create template system for generated code
- Generate all 5 existing resource mappings
- Validate generated code matches manual code

### Phase 3-7: Continue as originally planned

---

## Next Steps

1. **Share schema information** so I can refine the design
2. **Prototype Phase 0** to validate generation approach
3. **Review generated code** to ensure it matches manual code
4. **Decide on override strategy** for complex cases
5. **Build tooling** to automate the rest

This approach ensures we're building on top of existing code generation rather than duplicating effort.
