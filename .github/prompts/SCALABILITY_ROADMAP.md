# Scalability Roadmap: Multi-Resource Architecture

## Current State Assessment

### Architecture Overview
The converter currently handles DaVinci-specific resources with hardcoded logic per resource type:
- `flow_converter.go` - Flow conversion
- `application_converter.go` - Application conversion
- `connector_instance_converter.go` - Connector conversion
- `variable_converter.go` - Variable conversion
- `flow_policy_converter.go` - Flow policy conversion

### Current Pain Points

1. **Duplication**: Each exporter (`flow_exporter.go`, `application_exporter.go`, etc.) implements similar patterns
2. **Manual Orchestration**: `orchestrator.go` hardcodes resource ordering
3. **Tight Coupling**: Converters directly write HCL strings
4. **Limited Extensibility**: Adding new PingOne resources requires changes in multiple locations
5. **Naming Inconsistency**: Multiple `ensureUnique*` functions across exporters

## Proposed Architecture: Plugin-Based Resource Registry

### Design Principles

1. **Convention over Configuration**: Resources follow standard patterns
2. **Declarative Schemas**: Define resource structure once, generate conversion logic
3. **Dependency Graph First**: All resources register in graph before conversion
4. **Unified Interfaces**: Common abstraction for all PingOne resources

---

## Phase 1: Resource Registry Pattern

### 1.1 Create Resource Interface

**File**: `internal/registry/resource.go`

```go
package registry

import (
	"context"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

// ResourceConverter defines how a resource type is converted to Terraform
type ResourceConverter interface {
	// Type returns the internal resource type (e.g., "flow", "environment", "application")
	Type() string
	
	// TerraformType returns the Terraform resource type (e.g., "pingone_davinci_flow")
	TerraformType() string
	
	// Fetch retrieves all resources of this type from the API
	Fetch(ctx context.Context, client interface{}) ([]Resource, error)
	
	// Convert transforms a resource to HCL
	Convert(resource Resource, graph *resolver.DependencyGraph, skipDeps bool) (string, error)
	
	// Dependencies returns the list of resource types this depends on
	Dependencies() []string
	
	// ExtractDependencies parses resource data to find dependency IDs
	ExtractDependencies(resource Resource) ([]resolver.Dependency, error)
}

// Resource represents a single resource instance
type Resource struct {
	ID       string
	Name     string
	Type     string
	RawData  map[string]interface{}
}

// Registry manages all resource converters
type Registry struct {
	converters map[string]ResourceConverter
	graph      *resolver.DependencyGraph
}

func NewRegistry() *Registry {
	return &Registry{
		converters: make(map[string]ResourceConverter),
		graph:      resolver.NewDependencyGraph(),
	}
}

func (r *Registry) Register(converter ResourceConverter) {
	r.converters[converter.Type()] = converter
}

func (r *Registry) Get(resourceType string) (ResourceConverter, bool) {
	conv, exists := r.converters[resourceType]
	return conv, exists
}

// TopologicalSort returns converters in dependency order
func (r *Registry) TopologicalSort() ([]ResourceConverter, error) {
	// Implement Kahn's algorithm for topological sort
	// Returns converters ordered so dependencies come first
	// (e.g., connectors before flows, flows before applications)
}
```

### 1.2 Implement Resource Converters

**Example**: `internal/registry/converters/flow_converter.go`

```go
package converters

import (
	"context"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/converter"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/registry"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

type FlowConverter struct{}

func (fc *FlowConverter) Type() string {
	return "flow"
}

func (fc *FlowConverter) TerraformType() string {
	return "pingone_davinci_flow"
}

func (fc *FlowConverter) Dependencies() []string {
	return []string{"connector_instance", "variable"} // Flows depend on connectors and variables
}

func (fc *FlowConverter) Fetch(ctx context.Context, client interface{}) ([]registry.Resource, error) {
	apiClient := client.(*api.Client)
	flows, err := apiClient.ListFlows(ctx)
	if err != nil {
		return nil, err
	}
	
	resources := make([]registry.Resource, len(flows))
	for i, flow := range flows {
		flowDetail, err := apiClient.GetFlow(ctx, flow.FlowID)
		if err != nil {
			return nil, err
		}
		
		flowMap, err := convertFlowDetailToMap(flowDetail)
		if err != nil {
			return nil, err
		}
		
		resources[i] = registry.Resource{
			ID:      flow.FlowID,
			Name:    flow.Name,
			Type:    "flow",
			RawData: flowMap,
		}
	}
	return resources, nil
}

func (fc *FlowConverter) Convert(resource registry.Resource, graph *resolver.DependencyGraph, skipDeps bool) (string, error) {
	// Delegate to existing converter logic
	return converter.ConvertFlowToHCL(resource.RawData, "var.environment_id", skipDeps, graph)
}

func (fc *FlowConverter) ExtractDependencies(resource registry.Resource) ([]resolver.Dependency, error) {
	// Use resolver schema to extract dependencies
	schema := resolver.GetResourceSchema("flow")
	deps, err := resolver.ParseResourceDependencies(schema, resource.RawData)
	if err != nil {
		return nil, err
	}
	return deps, nil
}
```

### 1.3 Universal Exporter

**File**: `internal/exporter/universal_exporter.go`

```go
package exporter

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/registry"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

type UniversalExporter struct {
	registry *registry.Registry
	client   *api.Client
}

func NewUniversalExporter(client *api.Client) *UniversalExporter {
	reg := registry.NewRegistry()
	
	// Auto-register all converters
	registry.RegisterDefaultConverters(reg)
	
	return &UniversalExporter{
		registry: reg,
		client:   client,
	}
}

func (e *UniversalExporter) ExportAll(ctx context.Context, skipDeps bool) (string, error) {
	var hcl strings.Builder
	
	// 1. Get converters in dependency order
	converters, err := e.registry.TopologicalSort()
	if err != nil {
		return "", fmt.Errorf("failed to sort dependencies: %w", err)
	}
	
	// 2. Fetch all resources and register in dependency graph
	allResources := make(map[string][]registry.Resource)
	for _, conv := range converters {
		resources, err := conv.Fetch(ctx, e.client)
		if err != nil {
			return "", fmt.Errorf("failed to fetch %s: %w", conv.Type(), err)
		}
		allResources[conv.Type()] = resources
		
		// Register resources in dependency graph
		for _, res := range resources {
			sanitizedName := resolver.SanitizeName(res.Name, nil)
			e.registry.GetGraph().AddResource(conv.Type(), res.ID, sanitizedName)
		}
	}
	
	// 3. Parse dependencies for all resources
	for _, conv := range converters {
		for _, res := range allResources[conv.Type()] {
			deps, err := conv.ExtractDependencies(res)
			if err != nil {
				// Log warning but continue
				continue
			}
			
			for _, dep := range deps {
				e.registry.GetGraph().AddDependency(
					resolver.ResourceRef{Type: conv.Type(), ID: res.ID},
					dep.To,
					dep.Field,
					dep.Location,
				)
			}
		}
	}
	
	// 4. Generate HCL in dependency order
	for _, conv := range converters {
		hcl.WriteString(fmt.Sprintf("\n# %s Resources\n", strings.Title(conv.Type())))
		
		for _, res := range allResources[conv.Type()] {
			resHCL, err := conv.Convert(res, e.registry.GetGraph(), skipDeps)
			if err != nil {
				return "", fmt.Errorf("failed to convert %s %s: %w", conv.Type(), res.ID, err)
			}
			hcl.WriteString(resHCL)
			hcl.WriteString("\n\n")
		}
	}
	
	return hcl.String(), nil
}
```

---

## Phase 2: Expand to PingOne Resources

### 2.1 PingOne Resource Categories

#### Identity Management
- **Environments**: `pingone_environment`
- **Populations**: `pingone_population`
- **Groups**: `pingone_group`
- **Users**: `pingone_user`
- **Schemas**: `pingone_schema`, `pingone_schema_attribute`

#### Authentication & Authorization
- **Applications**: `pingone_application` (OAuth, SAML, OpenID)
- **Resource Servers**: `pingone_application_resource_grant`
- **Sign-On Policies**: `pingone_sign_on_policy`, `pingone_sign_on_policy_action`
- **MFA Policies**: `pingone_mfa_policy`, `pingone_mfa_fido_policy`

#### DaVinci (Current)
- **Flows**: `pingone_davinci_flow`
- **Applications**: `pingone_davinci_application`
- **Connections**: `pingone_davinci_connector`
- **Variables**: `pingone_davinci_variable`
- **Flow Policies**: `pingone_davinci_flow_policy`

#### Protect
- **Risk Policies**: `pingone_risk_policy`, `pingone_risk_predictor`

#### Verify
- **Verify Policies**: `pingone_verify_policy`

#### Credentials
- **Digital Wallets**: `pingone_credential_type`, `pingone_credential_issuance_rule`

### 2.2 Resource Converter Template

**File**: `internal/registry/converters/template.go`

```go
// Template for adding new PingOne resource converters
//
// To add a new resource:
// 1. Copy this template to <resource_type>_converter.go
// 2. Implement all interface methods
// 3. Register in internal/registry/register.go
// 4. Add schema to internal/resolver/schema.go
// 5. Add tests in internal/registry/converters/<resource_type>_converter_test.go

package converters

import (
	"context"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/api"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/registry"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

type TemplateConverter struct{}

func (tc *TemplateConverter) Type() string {
	return "RESOURCE_TYPE" // e.g., "environment", "population", "application"
}

func (tc *TemplateConverter) TerraformType() string {
	return "pingone_RESOURCE_TYPE" // e.g., "pingone_environment"
}

func (tc *TemplateConverter) Dependencies() []string {
	// List resource types this depends on
	// e.g., applications depend on environments and populations
	return []string{}
}

func (tc *TemplateConverter) Fetch(ctx context.Context, client interface{}) ([]registry.Resource, error) {
	// 1. Type assert client
	// 2. Call appropriate API method
	// 3. Convert to []registry.Resource
	return nil, nil
}

func (tc *TemplateConverter) Convert(resource registry.Resource, graph *resolver.DependencyGraph, skipDeps bool) (string, error) {
	// 1. Build HCL string
	// 2. Use graph.GetReferenceName() for dependency references
	// 3. Handle missing dependencies with resolver.GenerateTODOPlaceholder()
	return "", nil
}

func (tc *TemplateConverter) ExtractDependencies(resource registry.Resource) ([]resolver.Dependency, error) {
	// 1. Get schema: schema := resolver.GetResourceSchema(tc.Type())
	// 2. Parse: deps, err := resolver.ParseResourceDependencies(schema, resource.RawData)
	// 3. Return dependencies
	return nil, nil
}
```

### 2.3 Auto-Registration

**File**: `internal/registry/register.go`

```go
package registry

import (
	"github.com/samir-gandhi/davinci-terraform-converter/internal/registry/converters"
)

// RegisterDefaultConverters registers all built-in resource converters
func RegisterDefaultConverters(reg *Registry) {
	// DaVinci Resources
	reg.Register(&converters.VariableConverter{})
	reg.Register(&converters.ConnectorConverter{})
	reg.Register(&converters.FlowConverter{})
	reg.Register(&converters.DaVinciApplicationConverter{})
	reg.Register(&converters.FlowPolicyConverter{})
	
	// PingOne Platform Resources (Phase 2)
	// reg.Register(&converters.EnvironmentConverter{})
	// reg.Register(&converters.PopulationConverter{})
	// reg.Register(&converters.ApplicationConverter{})
	// reg.Register(&converters.SignOnPolicyConverter{})
	// ... add more as implemented
}

// RegisterConverter allows external packages to add custom converters
func RegisterConverter(reg *Registry, converter ResourceConverter) {
	reg.Register(converter)
}
```

---

## Phase 3: Configuration-Driven Conversion

### 3.1 Resource Metadata

**File**: `internal/registry/metadata/resources.yaml`

```yaml
# Declarative resource definitions
# Reduces boilerplate for simple resources

resources:
  - type: population
    terraform_type: pingone_population
    api_endpoint: /environments/{env_id}/populations
    dependencies: []
    fields:
      - name: name
        required: true
        type: string
      - name: description
        type: string
      - name: environment_id
        type: reference
        reference_type: environment
        
  - type: group
    terraform_type: pingone_group
    api_endpoint: /environments/{env_id}/groups
    dependencies:
      - population
    fields:
      - name: name
        required: true
        type: string
      - name: description
        type: string
      - name: population_id
        type: reference
        reference_type: population
        attribute: id
```

### 3.2 Code Generator

**File**: `cmd/generate/main.go`

```go
package main

import (
	"fmt"
	"os"
	"text/template"
	
	"gopkg.in/yaml.v3"
)

// Generate converter boilerplate from YAML definitions
// Usage: go run cmd/generate/main.go

type ResourceDef struct {
	Type          string   `yaml:"type"`
	TerraformType string   `yaml:"terraform_type"`
	APIEndpoint   string   `yaml:"api_endpoint"`
	Dependencies  []string `yaml:"dependencies"`
	Fields        []Field  `yaml:"fields"`
}

type Field struct {
	Name          string `yaml:"name"`
	Required      bool   `yaml:"required"`
	Type          string `yaml:"type"`
	ReferenceType string `yaml:"reference_type"`
	Attribute     string `yaml:"attribute"`
}

func main() {
	// 1. Read resources.yaml
	// 2. Generate converter files
	// 3. Generate schema definitions
	// 4. Generate test templates
}
```

---

## Phase 4: Advanced Features

### 4.1 Selective Export

```go
// Allow filtering by resource type
type ExportOptions struct {
	ResourceTypes []string          // Only export these types
	ExcludeTypes  []string          // Skip these types
	Filter        map[string]string // Filter by attributes (e.g., name contains "prod")
	SkipDeps      bool
}

func (e *UniversalExporter) ExportWithOptions(ctx context.Context, opts ExportOptions) (string, error) {
	// Filtered export logic
}
```

### 4.2 Import Support

```go
// Generate terraform import commands
func (e *UniversalExporter) GenerateImportCommands(ctx context.Context) ([]string, error) {
	// For each resource:
	// terraform import pingone_davinci_flow.my_flow <environment_id>/<flow_id>
}
```

### 4.3 State File Generation

```go
// Generate .tfstate file for existing resources
func (e *UniversalExporter) GenerateState(ctx context.Context) (*terraform.State, error) {
	// Construct Terraform state JSON
}
```

### 4.4 Diff Detection

```go
// Compare live resources vs Terraform configuration
func (e *UniversalExporter) DetectDrift(ctx context.Context, configPath string) (*DriftReport, error) {
	// Parse .tf files
	// Compare with live API data
	// Report differences
}
```

---

## Phase 5: Performance & Optimization

### 5.1 Parallel Fetching

```go
// Fetch independent resources in parallel
func (e *UniversalExporter) parallelFetch(ctx context.Context, converters []ResourceConverter) error {
	errChan := make(chan error, len(converters))
	
	for _, conv := range converters {
		go func(c ResourceConverter) {
			_, err := c.Fetch(ctx, e.client)
			errChan <- err
		}(conv)
	}
	
	// Collect errors
}
```

### 5.2 Incremental Export

```go
// Cache previously exported resources
type ExportCache struct {
	Resources map[string]CachedResource
	LastSync  time.Time
}

func (e *UniversalExporter) IncrementalExport(cache *ExportCache) (string, error) {
	// Only fetch resources changed since LastSync
}
```

### 5.3 Streaming Export

```go
// Stream HCL to writer instead of building in memory
func (e *UniversalExporter) StreamExport(ctx context.Context, w io.Writer) error {
	// Write resources as they're converted
	// Reduces memory footprint for large environments
}
```

---

## Migration Path

### Step 1: Refactor Existing (No Breaking Changes)
1. Create `internal/registry` package with interfaces
2. Wrap existing converters in `ResourceConverter` interface
3. Create `UniversalExporter` that delegates to existing exporters
4. Add tests to ensure output parity

### Step 2: Migrate DaVinci Resources
1. Move `flow_converter.go` → `internal/registry/converters/flow_converter.go`
2. Implement `ResourceConverter` interface
3. Update `orchestrator.go` to use `UniversalExporter`
4. Deprecate old exporter files

### Step 3: Add PingOne Platform Resources
1. Start with simple resources (populations, groups)
2. Add SDK client methods for new APIs
3. Register converters
4. Expand test coverage

### Step 4: Configuration-Driven Generation
1. Extract common patterns to YAML
2. Build code generator
3. Regenerate converters from definitions

### Step 5: Advanced Features
1. Add import command generation
2. Implement state file export
3. Build diff detection

---

## File Structure (Target State)

```
davinci-terraform-converter/
├── cmd/
│   ├── convert/           # CLI entrypoint
│   └── generate/          # Code generator from YAML
├── internal/
│   ├── api/              # API clients (existing)
│   ├── converter/        # Legacy converters (deprecated after migration)
│   ├── exporter/
│   │   ├── orchestrator.go          # Legacy (deprecated)
│   │   └── universal_exporter.go    # New unified exporter
│   ├── registry/
│   │   ├── resource.go              # Core interfaces
│   │   ├── registry.go              # Converter registry
│   │   ├── register.go              # Auto-registration
│   │   ├── metadata/
│   │   │   └── resources.yaml       # Declarative definitions
│   │   └── converters/
│   │       ├── template.go          # Template for new resources
│   │       ├── flow_converter.go
│   │       ├── application_converter.go
│   │       ├── environment_converter.go  # Phase 2
│   │       ├── population_converter.go   # Phase 2
│   │       └── ... (50+ converters)
│   ├── resolver/         # Dependency resolution (existing, enhanced)
│   │   ├── resolver.go
│   │   ├── schema.go                # Expand for all resources
│   │   ├── parser.go
│   │   └── reference.go
│   └── utils/            # Shared utilities (existing)
└── templates/
    ├── converter.go.tmpl  # Code gen templates
    └── test.go.tmpl
```

---

## Benefits of This Architecture

### For Developers
1. **Add new resources in <30 minutes**: Implement interface, register, done
2. **Consistent patterns**: All resources follow same structure
3. **Testable**: Each converter is independently testable
4. **Type-safe**: Go interfaces enforce contracts

### For Maintainability  
1. **Single source of truth**: Resource metadata in one place
2. **DRY**: No duplicated orchestration logic
3. **Extensible**: Plugin architecture allows external converters
4. **Backward compatible**: Can coexist with existing code during migration

### For Users
1. **More resources**: Easy to add entire PingOne platform
2. **Faster exports**: Parallel fetching
3. **Better error handling**: Per-resource error isolation
4. **Import support**: Generate import commands automatically

---

## Estimated Effort

| Phase | Description | Effort | Priority |
|-------|-------------|--------|----------|
| 1 | Resource Registry Pattern | 2-3 weeks | **HIGH** |
| 2 | Migrate DaVinci Resources | 1-2 weeks | **HIGH** |
| 3 | Add 10 PingOne Resources | 2-3 weeks | **MEDIUM** |
| 4 | Configuration-Driven | 2-3 weeks | **MEDIUM** |
| 5 | Advanced Features | 3-4 weeks | **LOW** |
| 6 | Performance Optimization | 1-2 weeks | **LOW** |

**Total**: 11-17 weeks for complete transformation

---

## Success Metrics

1. **Time to add new resource** < 30 minutes (vs current ~4 hours)
2. **Code duplication** < 10% (vs current ~40%)
3. **Test coverage** > 80% for all converters
4. **Export performance** < 30 seconds for 100 resources
5. **Resource support** 50+ PingOne resource types

---

## Next Steps (Immediate)

1. **Create PoC**: Implement Phase 1 for one resource (e.g., `variable`)
2. **Validate**: Ensure HCL output matches existing converter
3. **Document**: Add examples and migration guide
4. **Socialize**: Get feedback from team
5. **Iterate**: Refine interfaces based on PoC learnings
6. **Execute**: Begin full migration

---

## References

- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Go Plugin Architecture](https://go.dev/doc/effective_go#embedding)
- [PingOne Platform API](https://apidocs.pingidentity.com/pingone/platform/v1/api/)
- [Registry Pattern](https://en.wikipedia.org/wiki/Service_locator_pattern)
