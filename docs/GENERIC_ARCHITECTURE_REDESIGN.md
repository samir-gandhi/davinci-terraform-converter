# Generic Architecture Redesign

## Executive Summary

Current architecture requires creating 5-7 separate functions for each new resource type:

- Exporter function (`ExportXWithImports`)
- Converter function (`ConvertXToHCL`)
- JSON struct (`XResponse`)
- Variable extraction (`GetXVariableEligibleAttributes`)
- HCL generation with variables (`GenerateXHCLWithVariableReferences`)
- Import block logic
- Dependency registration

**Goal**: Reduce to **zero new functions** for standard resources by using a generic, schema-driven foundation.

**Approach**: Type-parameterized resource handlers with declarative schema definitions, inspired by Terraform provider SDK patterns and OpenAPI code generation.

---

## Current State Analysis

### Problems

1. **Duplication**: Each resource type duplicates ~80% of logic
2. **Scalability**: Adding 50 resources means 350+ functions
3. **Consistency**: Manual implementation causes drift in behavior
4. **Maintenance**: Bug fixes must be applied to N places
5. **Testing**: Each resource needs separate test suites
6. **Complexity**: High cognitive load to understand resource-specific code
7. **Module Generation Gap**: Variable extraction/usage not unified

### Current Resource Processing Flow

```text
API Response (SDK type)
  ↓
convertXToJSON() - Manual marshaling
  ↓
XResponse struct - Custom definition
  ↓
ConvertXWithOptions() - Resource-specific converter
  ↓
generateXHCL() - Resource-specific HCL builder
  ↓
ExportXWithImports() - Resource-specific orchestration
  ↓
GetXVariableEligibleAttributes() - Resource-specific extraction
  ↓
HCL string + variables
```

Each step requires manual implementation per resource type.

---

## Proposed Generic Architecture

### Core Concept: Resource Handler

A single generic `ResourceHandler[T]` type that handles ALL resources through schema-driven behavior.

```go
// Generic resource handler for any Terraform resource
type ResourceHandler[T any] struct {
    Schema       *ResourceSchema
    APIClient    APIClientInterface
    Graph        *DependencyGraph
    ImportGen    *ImportBlockGenerator
    Logger       Logger
}

// All operations available on any resource
func (h *ResourceHandler[T]) Export(ctx context.Context, opts ExportOptions) (*ResourceExport, error)
func (h *ResourceHandler[T]) ConvertToHCL(data T, opts ConvertOptions) (string, error)
func (h *ResourceHandler[T]) ExtractVariables(data T) ([]VariableEligibleAttribute, error)
func (h *ResourceHandler[T]) GenerateImportBlock(data T) (string, error)
func (h *ResourceHandler[T]) RegisterDependencies(data T) error
func (h *ResourceHandler[T]) ValidateResource(data T) error
```

### ResourceSchema: Declarative Configuration

Instead of writing code, declare resource behavior:

```go
type ResourceSchema struct {
    // Basic metadata
    TerraformType    string            // "pingone_davinci_variable"
    TerraformVersion string            // ">=1.5.0"
    APIPath          string            // "/environments/{envId}/variables"
    SDKType          reflect.Type      // reflect.TypeOf(DaVinciVariable{})
    
    // Identity and naming
    IDField          string            // "id" - which field is the resource ID
    NameField        string            // "name" - which field is the resource name
    NameSanitizer    func(string) string // Custom name sanitization
    
    // Schema definition
    Attributes       []AttributeSchema // All resource attributes
    
    // Dependency configuration
    DependsOn        []DependencyRule  // Dependencies on other resource types
    ProvidedBy       string            // Which other resource creates this (e.g., "environment")
    
    // Variable extraction rules
    VariableRules    []VariableRule    // Which attributes become variables
    
    // API operations
    ListOperation    ListFunc          // How to fetch multiple from API
    GetOperation     GetFunc           // How to fetch single from API
    
    // Import configuration
    ImportIDFormat   string            // "{envId}/{resourceId}" or custom
    ImportIDBuilder  func(T) string    // Custom import ID builder
    
    // HCL generation
    HCLTemplate      string            // Optional: Custom HCL template
    HCLGenerator     func(T) string    // Optional: Custom HCL generation
}

type AttributeSchema struct {
    Name             string
    TerraformName    string           // HCL attribute name (default: snake_case of Name)
    Type             AttributeType    // String, Number, Bool, Object, List, Map
    Required         bool
    Computed         bool
    Sensitive        bool
    
    // Variable eligibility
    VariableEligible bool             // Can this become a module variable?
    VariableDefault  interface{}      // Default value for module variable
    
    // Dependencies
    ReferencesType   string           // Resource type this references
    ReferenceField   string           // Field in referenced resource
    
    // Nested attributes (for objects)
    Attributes       []AttributeSchema
    
    // Custom behavior
    CustomMarshaler  func(interface{}) (string, error)
    CustomValidator  func(interface{}) error
}

type VariableRule struct {
    AttributePath    string           // "value" or "properties.apiKey"
    VariablePrefix   string           // "davinci_variable_"
    IsSecret         bool             // Omit value in module.tf
    DefaultValue     interface{}      // Default for variables.tf
}

type DependencyRule struct {
    ResourceType     string           // "pingone_davinci_connector_instance"
    FieldPath        string           // "connector.id"
    ReferenceFormat  string           // "{resourceType}.{resourceName}.id"
}
```

### Resource Registry

Central registry of all supported resources:

```go
// Global registry of resource schemas
type ResourceRegistry struct {
    schemas map[string]*ResourceSchema
    mu      sync.RWMutex
}

var GlobalRegistry = NewResourceRegistry()

// Register a resource schema
func (r *ResourceRegistry) Register(schema *ResourceSchema) error {
    // Validation: ensure no duplicate terraform types
    // Build internal indexes for fast lookup
}

// Get handler for a resource type
func (r *ResourceRegistry) GetHandler[T any](terraformType string) (*ResourceHandler[T], error) {
    schema := r.schemas[terraformType]
    return NewResourceHandler[T](schema), nil
}
```

### Example: Registering a Variable Resource

**Before** (current): 300+ lines across multiple files

**After** (proposed): ~50 lines of declarative schema

```go
func init() {
    GlobalRegistry.Register(&ResourceSchema{
        TerraformType:    "pingone_davinci_variable",
        TerraformVersion: ">=1.5.0",
        APIPath:          "/environments/{envId}/variables",
        SDKType:          reflect.TypeOf(pingone.DaVinciVariable{}),
        
        IDField:   "id",
        NameField: "name",
        
        Attributes: []AttributeSchema{
            {Name: "environment_id", Type: String, Required: true},
            {Name: "name", Type: String, Required: true},
            {Name: "type", Type: String, Required: true},
            {Name: "context", Type: String, Required: true},
            {Name: "value", Type: String, VariableEligible: true},
            {Name: "display_name", Type: String},
            {Name: "mutable", Type: Bool},
            {Name: "min", Type: Number},
            {Name: "max", Type: Number},
            {
                Name: "flow",
                Type: Object,
                Attributes: []AttributeSchema{
                    {
                        Name: "id",
                        Type: String,
                        ReferencesType: "pingone_davinci_flow",
                    },
                },
            },
        },
        
        VariableRules: []VariableRule{
            {
                AttributePath:  "value",
                VariablePrefix: "davinci_variable_",
                IsSecret:       false,
            },
        },
        
        DependsOn: []DependencyRule{
            {
                ResourceType:    "pingone_davinci_flow",
                FieldPath:       "flow.id",
                ReferenceFormat: "pingone_davinci_flow.{resourceName}.id",
            },
        },
        
        ImportIDFormat: "{envId}/{resourceId}",
        
        ListOperation: func(ctx context.Context, client APIClient, envID string) ([]interface{}, error) {
            return client.ListVariables(ctx, envID)
        },
    })
}
```

### Generic Exporter

Single exporter replaces all resource-specific exporters:

```go
type GenericExporter struct {
    registry      *ResourceRegistry
    client        *api.Client
    graph         *DependencyGraph
    importGen     *ImportBlockGenerator
    logger        Logger
}

// Export any resource type by terraform type name
func (e *GenericExporter) Export(
    ctx context.Context,
    terraformType string,
    opts ExportOptions,
) (*ResourceExport, error) {
    // Get handler for this resource type
    handler, err := e.registry.GetHandler(terraformType)
    if err != nil {
        return nil, fmt.Errorf("unsupported resource type: %s", terraformType)
    }
    
    // Use handler's export method (generic implementation)
    return handler.Export(ctx, opts)
}

// Export result with all necessary data
type ResourceExport struct {
    HCL                string
    ImportBlocks       string
    ExtractedVariables []VariableEligibleAttribute
    ResourceIDs        []string
    ResourceNames      []string
}
```

### Generic HCL Generator

Schema-driven HCL generation replaces manual string building:

```go
type HCLGenerator struct {
    schema *ResourceSchema
}

func (g *HCLGenerator) Generate(data interface{}, opts ConvertOptions) (string, error) {
    var hcl strings.Builder
    
    // Get resource name
    resourceName := g.getResourceName(data)
    
    // Resource block header
    hcl.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n",
        g.schema.TerraformType, resourceName))
    
    // Generate attributes from schema
    for _, attr := range g.schema.Attributes {
        value := g.getFieldValue(data, attr.Name)
        
        if attr.VariableEligible && opts.UseVariables {
            // Use variable reference
            varName := g.buildVariableName(resourceName, attr.Name)
            hcl.WriteString(fmt.Sprintf("  %s = var.%s\n", attr.TerraformName, varName))
        } else if attr.ReferencesType != "" && !opts.SkipDependencies {
            // Use terraform reference
            ref := g.buildReference(value, attr)
            hcl.WriteString(fmt.Sprintf("  %s = %s\n", attr.TerraformName, ref))
        } else {
            // Literal value
            hclValue := g.formatValue(value, attr.Type)
            hcl.WriteString(fmt.Sprintf("  %s = %s\n", attr.TerraformName, hclValue))
        }
    }
    
    hcl.WriteString("}\n")
    return hcl.String(), nil
}
```

### Generic Variable Extractor

Schema drives variable extraction automatically:

```go
type VariableExtractor struct {
    schema *ResourceSchema
}

func (e *VariableExtractor) Extract(data interface{}, resourceName string) ([]VariableEligibleAttribute, error) {
    var variables []VariableEligibleAttribute
    
    // Iterate over variable rules in schema
    for _, rule := range e.schema.VariableRules {
        value := e.getValueAtPath(data, rule.AttributePath)
        
        variables = append(variables, VariableEligibleAttribute{
            ResourceType:     e.schema.TerraformType,
            ResourceName:     resourceName,
            AttributePath:    rule.AttributePath,
            TerraformVarName: e.buildVarName(resourceName, rule),
            Value:            value,
            IsSecret:         rule.IsSecret,
            VariableType:     e.inferType(value),
            DefaultValue:     rule.DefaultValue,
        })
    }
    
    return variables, nil
}
```

---

## Usage Examples

### Adding a New Resource Type

**Before**: Write 7 functions across 4 files, ~500 lines of code

**After**: Register schema, ~80 lines of declarative configuration

```go
// File: internal/resources/application.go

func init() {
    GlobalRegistry.Register(&ResourceSchema{
        TerraformType: "pingone_davinci_application",
        APIPath:       "/environments/{envId}/applications",
        SDKType:       reflect.TypeOf(pingone.DaVinciApplication{}),
        
        IDField:   "id",
        NameField: "name",
        
        Attributes: []AttributeSchema{
            {Name: "environment_id", Type: String, Required: true},
            {Name: "name", Type: String, Required: true},
            {Name: "api_key_enabled", Type: Bool},
            {Name: "oauth", Type: Object, Attributes: []AttributeSchema{
                {Name: "enabled", Type: Bool},
                {Name: "values", Type: List, Attributes: []AttributeSchema{
                    {Name: "client_secret", Type: String, Sensitive: true, VariableEligible: true},
                }},
            }},
        },
        
        VariableRules: []VariableRule{
            {
                AttributePath:  "oauth.values[*].client_secret",
                VariablePrefix: "davinci_app_",
                IsSecret:       true,
            },
        },
        
        ImportIDFormat: "{envId}/{appId}",
        
        ListOperation: func(ctx context.Context, client APIClient, envID string) ([]interface{}, error) {
            return client.ListApplications(ctx, envID)
        },
    })
}
```

### Exporting All Resources

**Before**: Call each exporter manually in specific order

**After**: Use generic orchestrator with dependency-based ordering

```go
func ExportEnvironment(ctx context.Context, client *api.Client, opts ExportOptions) (string, error) {
    exporter := NewGenericExporter(GlobalRegistry, client, logger)
    
    // Define resource types to export (in dependency order, or let graph sort)
    resourceTypes := []string{
        "pingone_davinci_variable",
        "pingone_davinci_connector_instance",
        "pingone_davinci_flow",
        "pingone_davinci_application",
        "pingone_davinci_application_flow_policy",
    }
    
    // Export all resources generically
    var results []*ResourceExport
    for _, resourceType := range resourceTypes {
        result, err := exporter.Export(ctx, resourceType, opts)
        if err != nil {
            return "", err
        }
        results = append(results, result)
    }
    
    // Combine results
    return exporter.CombineExports(results), nil
}
```

### Module Generation

**Before**: Custom logic in module_export.go with manual HCL regeneration

**After**: Generic module generator with schema-driven variable handling

```go
func GenerateModule(exportData *ExportedData, opts ModuleOptions) (*Module, error) {
    generator := NewModuleGenerator(GlobalRegistry)
    
    // Extract all variables from all resources (schema-driven)
    allVariables := generator.ExtractAllVariables(exportData)
    
    // Build variable map
    variableMap := generator.BuildVariableMap(allVariables)
    
    // Regenerate HCL with variable references (schema-driven)
    childModuleHCL := generator.RegenerateWithVariables(exportData, variableMap)
    
    // Generate module files
    return &Module{
        ChildModule: ModuleDefinition{
            VariablesTF:  generator.GenerateVariablesTF(allVariables),
            VersionsTF:   generator.GenerateVersionsTF(),
            ResourcesTF:  childModuleHCL,
            OutputsTF:    generator.GenerateOutputsTF(exportData),
        },
        RootModule: ModuleDefinition{
            ModuleTF:     generator.GenerateModuleTF(allVariables, opts),
            ImportsTF:    generator.GenerateImportsTF(exportData),
        },
    }, nil
}
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Goal**: Establish generic infrastructure without breaking existing code

```text
Tasks:
1. Create ResourceSchema types and structures
2. Create ResourceRegistry with registration system
3. Create generic ResourceHandler[T] type
4. Create schema-driven AttributeSchema system
5. Build reflection utilities for schema introspection
6. Create comprehensive unit tests for foundation

Deliverables:
- internal/schema/resource_schema.go
- internal/schema/attribute_schema.go
- internal/schema/registry.go
- internal/schema/handler.go
- internal/schema/reflection.go
- Test coverage >90%

Success Criteria:
- Can register a schema
- Can instantiate a handler from registry
- Can introspect schema attributes
- All tests pass
```

### Phase 2: Generic HCL Generation (Week 2)

**Goal**: Replace resource-specific HCL generators with schema-driven generator

```text
Tasks:
1. Create generic HCLGenerator
2. Implement schema-driven attribute iteration
3. Handle primitive types (string, number, bool)
4. Handle complex types (object, list, map)
5. Handle variable references via schema
6. Handle terraform references via schema
7. Migrate one resource (variable) to new system
8. Parallel test: old vs new output identical

Deliverables:
- internal/hclgen/generator.go
- internal/hclgen/types.go
- internal/hclgen/formatter.go
- Side-by-side comparison tests

Success Criteria:
- Variable resource generates identical HCL
- Both generators pass same test suite
- Performance comparable or better
```

### Phase 3: Generic Variable Extraction (Week 2)

**Goal**: Replace resource-specific extraction with schema-driven extraction

```text
Tasks:
1. Create generic VariableExtractor
2. Implement VariableRule evaluation
3. Handle nested attribute paths
4. Handle array/list attribute paths
5. Integrate with module generation
6. Migrate variable + connector resources

Deliverables:
- internal/varextract/extractor.go
- internal/varextract/rules.go
- Updated module generation logic

Success Criteria:
- Extracts same variables as current implementation
- Module generation tests pass
- Variable references work in child modules
```

### Phase 4: Generic Exporter (Week 3)

**Goal**: Replace resource-specific exporters with generic exporter

```text
Tasks:
1. Create GenericExporter
2. Implement schema-driven API calls
3. Implement schema-driven dependency registration
4. Implement schema-driven import block generation
5. Migrate all 5 resource types to generic system
6. Remove old resource-specific exporters

Deliverables:
- internal/exporter/generic_exporter.go
- Updated orchestrator.go using generic exporter
- Deprecated old exporters (mark for removal)

Success Criteria:
- Export output identical for all resources
- All integration tests pass
- Performance acceptable
```

### Phase 5: Resource Schema Definitions (Week 3-4)

**Goal**: Create schema definitions for all existing resources

```text
Tasks:
1. Define schema for pingone_davinci_variable
2. Define schema for pingone_davinci_connector_instance
3. Define schema for pingone_davinci_flow
4. Define schema for pingone_davinci_application
5. Define schema for pingone_davinci_application_flow_policy
6. Register all schemas in init functions
7. Comprehensive integration testing

Deliverables:
- internal/resources/variable.go
- internal/resources/connector.go
- internal/resources/flow.go
- internal/resources/application.go
- internal/resources/policy.go
- internal/resources/registry.go

Success Criteria:
- All 5 resources fully defined
- Export works for entire environment
- Module generation works
- All tests pass
```

### Phase 6: Remove Legacy Code (Week 4)

**Goal**: Delete old resource-specific implementations

```text
Tasks:
1. Run comprehensive test suite
2. Verify backwards compatibility
3. Remove old converter files
4. Remove old exporter files
5. Update documentation
6. Update ARCHITECTURE.md

Deliverables:
- Deleted: variable_converter.go, connector_instance_converter.go, etc.
- Deleted: variable_exporter.go, connector_exporter.go, etc.
- Updated: ARCHITECTURE.md, README.md
- Migration guide for contributors

Success Criteria:
- Codebase reduced by ~2000 lines
- All tests passing
- Documentation updated
```

### Phase 7: OpenAPI Code Generation (Week 5)

**Goal**: Generate resource schemas from OpenAPI spec

```text
Tasks:
1. Create OpenAPI spec parser
2. Map OpenAPI types to AttributeSchema
3. Generate ResourceSchema from OpenAPI operations
4. Create code generator for resource files
5. Document schema generation workflow

Deliverables:
- cmd/generate-schemas/main.go
- internal/codegen/openapi.go
- internal/codegen/templates/
- docs/CODE_GENERATION.md

Success Criteria:
- Can generate schema from OpenAPI spec
- Generated schema matches manual schema
- Can add new resource in <5 minutes
```

---

## Benefits Analysis

### Code Reduction

| Metric | Current | After Redesign | Reduction |
|--------|---------|----------------|-----------|
| Lines per resource | ~500 | ~80 | 84% |
| Files per resource | 4-5 | 1 | 80% |
| Functions per resource | 7 | 0 | 100% |
| Test files per resource | 2 | Shared | 50% |

**For 50 resources:**

- Current: 25,000 lines, 200 files, 350 functions
- Redesign: 4,000 lines, 50 files, 0 new functions

### Maintenance Benefits

1. **Bug Fixes**: Fix once in generic handler vs N times
2. **Features**: Add once for all resources (e.g., validation, formatting)
3. **Testing**: Test generic behavior once, not per resource
4. **Onboarding**: New developers learn one pattern, not N patterns
5. **Consistency**: Schema enforcement prevents drift

### Performance Benefits

1. **Compile Time**: Fewer files, faster builds
2. **Runtime**: Minimal overhead from generics/reflection
3. **Memory**: Shared handlers vs duplicate implementations

### Scalability Benefits

1. **50+ Resources**: Add schema, not code
2. **Multiple Providers**: Generic foundation supports PingFederate, PingAccess, etc.
3. **OpenAPI Generation**: Automated schema creation
4. **Community Contributions**: Lower barrier to adding resources

---

## Risk Analysis

### Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Complex resources need custom logic | High | Medium | Schema allows custom HCL generators, validators |
| Performance degradation | Low | High | Benchmark suite, optimization pass |
| Breaking changes | Low | High | Parallel implementation, side-by-side testing |
| Learning curve | Medium | Low | Comprehensive docs, examples |
| Over-engineering | Medium | Medium | Start simple, add complexity only when needed |

### Rollback Strategy

Each phase maintains backwards compatibility:

- Phase 1-3: New code alongside old code
- Phase 4-5: Generic system proven, start migration
- Phase 6: Delete old code only after full validation

If issues arise, revert to previous phase and reassess.

---

## Alternative Approaches Considered

### 1. Interface-Based Approach

```go
type ResourceConverter interface {
    ConvertToHCL(data interface{}) (string, error)
    ExtractVariables(data interface{}) ([]Variable, error)
    // ...
}
```

**Rejected**: Still requires implementing interface per resource, no code reduction.

### 2. Code Generation Only

Generate resource-specific code from schemas, keep separate files.

**Rejected**: Doesn't reduce maintenance burden, just automates duplication.

### 3. Reflection-Only (No Generics)

Use pure reflection without Go generics.

**Rejected**: Loses type safety, harder to debug, worse IDE support.

### 4. Terraform Schema Import

Import Terraform provider schemas directly.

**Rejected**: Requires running provider, circular dependency, overkill for our use case.

---

## Success Metrics

### Quantitative Metrics

- [ ] Code reduction: >80% fewer lines for resource handling
- [ ] Time to add resource: <1 hour (vs 4-8 hours currently)
- [ ] Test coverage: >85% for generic components
- [ ] Performance: <10% overhead vs current implementation
- [ ] Build time: <5% increase (acceptable for maintainability gain)

### Qualitative Metrics

- [ ] New contributor can add resource in single session
- [ ] Bug fixes apply to all resources automatically
- [ ] Module generation works for all resources
- [ ] OpenAPI-driven workflow documented and tested
- [ ] Architecture praised as best-in-class for Terraform converters

---

## Next Steps

1. **Review and Approval**: Get team feedback on this design
2. **Prototype**: Build Phase 1 foundation with one resource
3. **Validate**: Ensure generic approach works for complex resources (flows)
4. **Commit**: Proceed with full implementation if prototype successful
5. **Document**: Create contributor guide for adding resources via schemas

---

## Appendix A: Complex Resource Handling

### Flow Resources (Most Complex Case)

Flows have nested graph_data with dynamic node structures. Schema approach:

```go
GlobalRegistry.Register(&ResourceSchema{
    TerraformType: "pingone_davinci_flow",
    
    Attributes: []AttributeSchema{
        // Standard attributes
        {Name: "environment_id", Type: String, Required: true},
        {Name: "name", Type: String, Required: true},
        
        // Complex nested attribute with custom handling
        {
            Name: "graph_data",
            Type: Object,
            CustomMarshaler: func(data interface{}) (string, error) {
                // Custom logic for graph_data complexity
                return marshalGraphData(data)
            },
        },
    },
    
    // Custom HCL generator for complex parts
    HCLGenerator: func(flow Flow) string {
        // Use generic generator for simple attributes
        // Use custom logic for graph_data
        return generateFlowHCL(flow)
    },
})
```

**Key**: Schema supports custom handlers for exceptions, defaults to generic behavior.

---

## Appendix B: OpenAPI Schema Example

Example OpenAPI operation → ResourceSchema:

```yaml
# OpenAPI spec
paths:
  /environments/{envId}/variables:
    get:
      operationId: listVariables
      responses:
        '200':
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Variable'

components:
  schemas:
    Variable:
      type: object
      required: [name, type, context]
      properties:
        id: {type: string, format: uuid}
        name: {type: string}
        type: {type: string, enum: [string, number, boolean, object]}
        context: {type: string, enum: [company, flowInstance, user]}
        value: {type: string}
```

Generated schema:

```go
// Auto-generated from OpenAPI spec
func init() {
    GlobalRegistry.Register(&ResourceSchema{
        TerraformType:    "pingone_davinci_variable",
        APIPath:          "/environments/{envId}/variables",
        IDField:          "id",
        NameField:        "name",
        
        Attributes: []AttributeSchema{
            {Name: "id", Type: String, Computed: true},
            {Name: "name", Type: String, Required: true},
            {Name: "type", Type: String, Required: true},
            {Name: "context", Type: String, Required: true},
            {Name: "value", Type: String, VariableEligible: true},
        },
        
        // Generated from OpenAPI operation
        ListOperation: func(ctx context.Context, client APIClient, envID string) ([]interface{}, error) {
            return client.ListVariables(ctx, envID)
        },
    })
}
```

---

## Conclusion

This redesign transforms the project from **resource-specific implementations** to a **generic, schema-driven foundation**. The architecture is:

- **Scalable**: 50+ resources with minimal code
- **Maintainable**: Fix once, apply everywhere
- **Consistent**: Schema enforcement prevents drift
- **Extensible**: Custom handlers for exceptions
- **Automatable**: OpenAPI code generation
- **Type-Safe**: Go generics preserve compile-time checks

**Estimated effort**: 4-5 weeks full implementation, 3-6 months ROI for 50 resources.

**Recommendation**: Proceed with Phase 1 prototype to validate approach before committing to full redesign.
