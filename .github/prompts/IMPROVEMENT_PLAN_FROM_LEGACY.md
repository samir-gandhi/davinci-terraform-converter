# Improvement Plan: Legacy dvtf-pingctl vs New davinci-terraform-converter

## Executive Summary

This document analyzes the differences between the legacy `dvtf-pingctl` (for legacy davinci provider) and the new `davinci-terraform-converter` (for pingone provider), identifying critical improvements needed to minimize consumer migration effort.

## Schema Differences: Legacy vs New Provider

### Legacy Provider Schema (`davinci_flow`)
```hcl
resource "davinci_flow" "example" {
  environment_id = var.environment_id
  flow_json      = file("./flow.json")
  
  # Explicit mapping blocks required
  connection_link {
    id                           = davinci_connection.http.id
    name                         = davinci_connection.http.name
    replace_import_connection_id = "867ed4363b2bc21c860085ad2baa817d"
  }
  
  subflow_link {
    id                        = davinci_flow.subflow.id
    name                      = davinci_flow.subflow.name
    replace_import_subflow_id = "07503fed5c02849dbbd5ee932da654b2"
  }
}
```

### New Provider Schema (`pingone_davinci_flow`)
```hcl
resource "pingone_davinci_flow" "example" {
  environment_id = var.environment_id
  name           = "Example Flow"
  description    = "Flow description"
  
  # Inline graph_data structure - NO external file reference
  graph_data {
    elements {
      nodes = [
        {
          data = {
            id            = "node1"
            connection_id = pingone_davinci_connector_instance.http.id  # Direct reference
            # ... full node definition inline
          }
        }
      ]
    }
  }
  
  # NO connection_link or subflow_link blocks
  # References are INLINE in graph_data
}
```

**Critical Difference**: New provider embeds entire flow structure inline with direct Terraform references, while legacy provider uses external JSON file with mapping blocks.

---

## Comparison Analysis

### 1. Dependency Identification & Mapping

#### Legacy CLI (`dvtf-pingctl`)

**How it works:**
```go
// Scans graphData.Elements.Nodes for connections
for _, node := range nodes {
    if nodeData.ConnectorID != nil && nodeData.ConnectionID != nil {
        // Creates connection_link mapping
        flowConnectionLinks = append(flowConnectionLinks, flowConnectionLink{
            ConnectionRefID:     resourceName,
            ConnectorID:         *nodeData.ConnectorID,
            ReplaceConnectionID: *nodeData.ConnectionID,  // Original UUID
        })
    }
    
    // Scans for subflow references
    if nodeData.Properties.SubFlowID != nil {
        subflowLinks = append(subflowLinks, flowSubflowLink{
            FlowRefID:        sanitizeResourceName(*subflowID.Label),
            SubFlowName:      *subflowID.Label,
            ReplaceSubflowID: *subflowID.Value,  // Original UUID
        })
    }
}
```

**Output:**
```hcl
resource "davinci_flow" "main_flow" {
  flow_json = file("assets/flows/main_flow.json")
  
  connection_link {
    id                           = davinci_connection.httpconnector__conn123.id
    name                         = davinci_connection.httpconnector__conn123.name
    replace_import_connection_id = "conn123"  # UUID from JSON
  }
  
  subflow_link {
    id                        = davinci_flow.my_subflow.id
    name                      = davinci_flow.my_subflow.name
    replace_import_subflow_id = "subflow456"  # UUID from JSON
  }
}
```

**Key Features:**
- ✅ Keeps flow JSON **external** in `assets/flows/` directory
- ✅ Creates explicit mapping blocks with original UUIDs
- ✅ Generates separate resources for all dependencies first
- ✅ Uses `replace_import_*_id` to map old UUIDs to new resources

#### New CLI (`davinci-terraform-converter`)

**How it works:**
```go
// Converts entire flow structure to inline HCL
func ConvertFlowToHCL(flowData map[string]interface{}, ...) {
    // Generates inline graph_data block
    writeGraphDataBlock(&hcl, graphData, skipDependencies, graph)
    
    // Inside node generation:
    if connectionID := node["connectionId"]; connectionID != "" {
        if graph != nil {
            ref := graph.GetReference("pingone_davinci_connector_instance", connectionID)
            hcl.WriteString(fmt.Sprintf("connection_id = %s\n", ref))
        } else {
            hcl.WriteString(fmt.Sprintf("connection_id = %q\n", connectionID))
        }
    }
}
```

**Output:**
```hcl
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id
  name           = "Main Flow"
  
  graph_data {
    elements {
      nodes = [
        {
          data = {
            id            = "node1"
            connection_id = pingone_davinci_connector_instance.httpconnector__conn123.id
            # Full inline structure - 500+ lines
          }
        }
      ]
    }
  }
  # NO connection_link blocks - references are inline
}
```

**Key Issues:**
- ❌ Embeds **entire** flow JSON inline (can be 1000+ lines per flow)
- ❌ No explicit mapping of original UUIDs to new resource references
- ⚠️ Harder to review/debug dependency mappings
- ❌ Cannot use external file reference like `file("flow.json")`

---

### 2. Multi-Flow Handling

#### Legacy CLI (`dvtf-pingctl`)

**Process:**
1. **Parses multi-flow JSON** with flows array
2. **Splits flows** into separate files
3. **Generates dependency graph** across all flows
4. **Creates ordered resources** (dependencies first, then flows that use them)

```go
// Handles multi-flow exports
for _, flow := range parsedIntf["flows"].([]interface{}) {
    flowName := flow.(map[string]interface{})["name"].(string)
    
    // Create asset file for each flow
    pathVar := fmt.Sprintf("assets/flows/%s.json", sanitizeResourceName(flowName))
    
    // Generate HCL with references to other flows
    buildDataSingleFlow(flow, pathVar)
}
```

**Output Structure:**
```
generated/
├── assets/
│   └── flows/
│       ├── main_flow.json       # Original JSON preserved
│       ├── subflow_1.json
│       └── subflow_2.json
├── davinci_variables.tf         # Shared variables
├── davinci_connectors.tf        # Shared connectors
└── davinci_flows.tf             # All flows with proper ordering
```

**Generated Flow with Dependencies:**
```hcl
# Variables first (no dependencies)
resource "davinci_variable" "language" {
  context = "flowInstance"
  name    = "language"
  value   = "en"
}

# Connectors (no dependencies)
resource "davinci_connection" "httpconnector__abc123" {
  connector_id = "httpConnector"
  name         = "HTTP"
}

# Subflows (depend on connectors, not on main flow)
resource "davinci_flow" "subflow_1" {
  depends_on = [
    davinci_variable.language,  # Explicit dependency
  ]
  
  flow_json = file("assets/flows/subflow_1.json")
  
  connection_link {
    id   = davinci_connection.httpconnector__abc123.id
    name = davinci_connection.httpconnector__abc123.name
    replace_import_connection_id = "abc123"
  }
}

# Main flow (depends on subflows)
resource "davinci_flow" "main_flow" {
  depends_on = [
    davinci_variable.language,
  ]
  
  flow_json = file("assets/flows/main_flow.json")
  
  subflow_link {
    id                        = davinci_flow.subflow_1.id
    name                      = davinci_flow.subflow_1.name
    replace_import_subflow_id = "subflow123"
  }
}
```

#### New CLI (`davinci-terraform-converter`)

**Current Status:**
```go
// ConvertMultiFlowWithOptions exists but limited
func ConvertMultiFlowWithOptions(multiFlowJSON []byte, skipDeps bool) ([]string, error) {
    // Returns array of HCL strings
    // Each flow converted independently
    // NO cross-flow dependency resolution
}
```

**Current Output:**
```hcl
# Flow 1 - all inline, no awareness of Flow 2
resource "pingone_davinci_flow" "subflow_1" {
  environment_id = var.environment_id
  name           = "Subflow 1"
  
  graph_data {
    elements {
      nodes = [
        # 500 lines of inline structure
      ]
    }
  }
}

# Flow 2 - references subflow but no ordering
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id
  name           = "Main Flow"
  
  graph_data {
    elements {
      nodes = [
        {
          data = {
            # Subflow reference - but which resource?
            # Could be hardcoded UUID or missing reference
          }
        }
      ]
    }
  }
}
```

**Key Issues:**
- ❌ Returns array of strings, not cohesive multi-file output
- ❌ No explicit ordering of flow dependencies
- ❌ No shared resource deduplication
- ❌ Cannot generate single `.tf` file with proper `depends_on`

---

### 3. Terraform Variable Generation (Module Support)

#### Legacy CLI (`dvtf-pingctl`)

**Generated Structure:**
```hcl
# davinci_variables.tf - DaVinci variables
resource "davinci_variable" "company_name" {
  environment_id = local.pingone_environment_id  # References local
  context        = "company"
  name           = "companyName"
  value          = "Acme Corp"
}

# davinci_connectors.tf - Stub with TODOs
resource "davinci_connection" "httpconnector__abc123" {
  environment_id = local.pingone_environment_id
  connector_id   = "httpConnector"
  name           = "HTTP"

  // properties based on the connector type
  // Visit the DaVinci Connector Parameter Reference for details:
  // https://registry.terraform.io/providers/pingidentity/davinci/latest/docs/guides/connector-reference
}

# Module interface variables (NOT generated but documented in README)
# User must create:
variable "pingone_environment_id" {
  type        = string
  description = "PingOne Environment ID"
}

locals {
  pingone_environment_id = var.pingone_environment_id
}
```

**Module Usage Pattern:**
```hcl
# Root module
module "davinci_flows" {
  source = "./generated"
  
  pingone_environment_id = pingone_environment.prod.id
}
```

**Key Features:**
- ✅ Uses `local.pingone_environment_id` consistently
- ✅ Documentation explains module pattern
- ✅ Generated code ready for module usage
- ⚠️ Requires manual creation of `variables.tf` and `locals.tf`

#### New CLI (`davinci-terraform-converter`)

**Current Output:**
```hcl
# Single file output - NOT module-ready
resource "pingone_davinci_variable" "company_name" {
  environment_id = var.environment_id  # Hardcoded var name
  context        = "company"
  name           = "companyName"
  value          = "Acme Corp"
}

resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id  # Same hardcoded var
  name           = "Main Flow"
  # ...
}
```

**Issues:**
- ❌ No `variables.tf` generation
- ❌ No `locals.tf` generation
- ❌ No module structure guidance
- ❌ Hardcoded `var.environment_id` without variable definition
- ❌ Single-file output not suitable for large exports

---

### 4. Shell Pipe Support

#### Legacy CLI (`dvtf-pingctl`)

**Implementation:**
```go
// cmd/root.go
var jsonContents string

func init() {
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) == 0 {
        // Stdin is pipe
        bytes, _ := io.ReadAll(os.Stdin)
        jsonContents = string(bytes)
    }
}

// cmd/generate.go
func Run() {
    if jsonContents != "" {
        dvFlow, _ := flow.NewFromPipe(string(jsonContents))
    } else {
        dvFlow, _ := flow.NewFromPaths(jsonFilePath)
    }
}
```

**Usage:**
```bash
# Pipe support
cat flow.json | dvtf-pingctl generate -o ./output

# File-based
dvtf-pingctl generate -e flow.json -o ./output

# Multi-file
dvtf-pingctl generate -e flow1.json -e flow2.json -o ./output
```

#### New CLI (`davinci-terraform-converter`)

**Current Implementation:**
```go
// cmd/davinci_to_hcl.go
func runConvert() {
    // Only supports --flow-json flag
    flowJSON := flags.String("flow-json", "", "Path to input file")
    
    flowJSONBytes, _ := os.ReadFile(*flowJSON)
    hcl, _ := converter.ConvertWithOptions(flowJSONBytes, skipDeps)
}
```

**Current Limitations:**
- ❌ No pipe support
- ❌ Must use file path
- ❌ No stdin detection
- ⚠️ Cannot chain with `jq` for preprocessing

---

## Improvement Plan

### Priority 1: Critical Schema Alignment

#### 1.1 External File Reference Support (HIGH)

**Problem**: New provider embeds entire flow inline (1000+ lines), legacy kept flows external.

**Solution Options:**

**Option A: Generate External JSON + Inline References (Recommended)**
```hcl
# Generated: flows/main_flow_export.json (original export preserved)
# Generated: main_flow.tf
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id
  name           = "Main Flow"
  
  # Use graph_data with references to external file data
  # This may require provider enhancement
  graph_data = jsondecode(file("flows/main_flow_export.json")).graphData
}
```

**Option B: Hybrid Approach - Preserve Original Export as Documentation**
```hcl
# Generated: flows/main_flow_original.json (reference only)
# Generated: main_flow.tf
# Original export: flows/main_flow_original.json
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id
  name           = "Main Flow"
  
  # Inline graph_data (required by provider)
  graph_data {
    elements {
      nodes = [
        # ... full structure
      ]
    }
  }
}
```

**Option C: Comment-Based Reference (Minimal Change)**
```hcl
# Source: flows/main_flow_export.json
# SHA256: abc123...
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id
  
  # Inline structure with reference comment
  graph_data { /* ... */ }
}
```

**Recommendation**: Start with Option B (preserve originals as docs), then explore Option A with PingOne provider team.

**Implementation:**
```go
// New: ExportOptions struct
type ExportOptions struct {
    PreserveOriginalExports bool  // Save original JSON files
    InlineGraphData         bool  // True for new provider
    UseExternalFileRefs     bool  // Future: if provider supports
}

// Updated converter
func ConvertFlowWithExportPreservation(flowData map[string]interface{}, opts ExportOptions) (hcl string, assets []Asset, error) {
    assets := []Asset{}
    
    if opts.PreserveOriginalExports {
        // Save original JSON
        assets = append(assets, Asset{
            Path:    "flows/" + sanitizeName(flowName) + "_export.json",
            Content: originalJSON,
        })
    }
    
    // Generate HCL
    if opts.InlineGraphData {
        // Current behavior
        hcl = generateInlineGraphData(flowData)
    } else {
        // Legacy behavior
        hcl = generateFileReference(flowData)
    }
    
    return hcl, assets, nil
}
```

#### 1.2 Connection/Subflow Reference Mapping (CRITICAL)

**Problem**: New provider has no `connection_link`/`subflow_link` blocks - references are inline in nodes.

**Current Legacy Approach:**
```hcl
connection_link {
  id                           = davinci_connection.http.id
  name                         = davinci_connection.http.name
  replace_import_connection_id = "867ed...baa817d"  # Track original UUID
}
```

**New Provider Requirement:**
```hcl
graph_data {
  elements {
    nodes = [
      {
        data = {
          connection_id = pingone_davinci_connector_instance.http.id  # Direct inline
        }
      }
    ]
  }
}
```

**Solution: Generate Reference Mapping Document**

Create a sidecar mapping file to help with migration:

```hcl
# Generated: MIGRATION_GUIDE.md
# Connection UUID to Terraform Resource Mapping

| Original UUID | DaVinci Export | New Resource Reference |
|---------------|----------------|------------------------|
| 867ed...817d  | HTTP Connector | pingone_davinci_connector_instance.httpconnector__867ed_817d.id |
| 33329...1ea3  | Flow Connector | pingone_davinci_connector_instance.flowconnector__33329_1ea3.id |

| Original Subflow UUID | DaVinci Export | New Resource Reference |
|-----------------------|----------------|------------------------|
| 07503...654b2         | Subflow 1      | pingone_davinci_flow.subflow_1.id |
```

**Implementation:**
```go
// Add to export command output
type DependencyMapping struct {
    OriginalUUID     string
    ResourceType     string // "connector_instance", "flow", "variable"
    ResourceName     string // Terraform resource name
    TerraformRef     string // Full reference path
    SourceFlowName   string
    NodeLabel        string // Human-readable name from flow
}

func GenerateMigrationGuide(graph *resolver.DependencyGraph, outputPath string) error {
    mappings := []DependencyMapping{}
    
    // Extract all mappings from graph
    for resourceType, instances := range graph.GetAllMappings() {
        for uuid, name := range instances {
            mappings = append(mappings, DependencyMapping{
                OriginalUUID: uuid,
                ResourceType: resourceType,
                ResourceName: name,
                TerraformRef: fmt.Sprintf("%s.%s.id", resourceType, name),
            })
        }
    }
    
    // Write markdown table
    return writeMigrationGuide(mappings, outputPath)
}
```

**Usage in Export:**
```go
// cmd/export.go
func runExport() {
    // ... existing export logic
    
    // Generate migration guide
    if err := converter.GenerateMigrationGuide(graph, outputPath+"/MIGRATION_GUIDE.md"); err != nil {
        return err
    }
    
    logger.Success("Generated migration guide: MIGRATION_GUIDE.md")
}
```

---

### Priority 2: Multi-Flow Enhancement

#### 2.1 Unified Multi-Flow Output (HIGH)

**Problem**: Current implementation returns `[]string`, not cohesive multi-file structure.

**Solution: Multi-File Generator**

```go
// New: MultiFlowOutput structure
type MultiFlowOutput struct {
    MainFile      string            // main.tf with all resources
    FlowFiles     map[string]string // flow_name.tf -> HCL
    VariableFiles map[string]string // variables.tf, outputs.tf
    Assets        map[string][]byte // Original JSON exports
    Metadata      ExportMetadata    // Dependency graph, ordering
}

type ExportMetadata struct {
    FlowOrder         []string                    // Dependency-sorted flow names
    SharedConnectors  []string                    // Connector resource names
    SharedVariables   []string                    // Variable resource names
    DependencyGraph   *resolver.DependencyGraph
}

// Enhanced multi-flow converter
func ConvertMultiFlowToFiles(multiFlowJSON []byte, opts ExportOptions) (*MultiFlowOutput, error) {
    output := &MultiFlowOutput{
        FlowFiles:     make(map[string]string),
        VariableFiles: make(map[string]string),
        Assets:        make(map[string][]byte),
    }
    
    // Parse multi-flow export
    var multiFlow struct {
        Flows []map[string]interface{} `json:"flows"`
    }
    json.Unmarshal(multiFlowJSON, &multiFlow)
    
    // Build dependency graph across ALL flows
    graph := resolver.NewDependencyGraph()
    for _, flowData := range multiFlow.Flows {
        RegisterFlowDependencies(graph, flowData)
    }
    
    // Topological sort - subflows before parent flows
    sortedFlows := graph.TopologicalSort()
    output.Metadata.FlowOrder = sortedFlows
    
    // Generate shared resources FIRST
    output.VariableFiles["variables.tf"] = GenerateSharedVariables(graph)
    output.MainFile += GenerateSharedConnectors(graph)
    
    // Generate flows in dependency order
    for _, flowName := range sortedFlows {
        flowData := getFlowByName(multiFlow.Flows, flowName)
        
        // Generate flow HCL
        flowHCL, err := ConvertFlowToHCL(flowData, opts, graph)
        if err != nil {
            return nil, err
        }
        
        // Store in separate file or main file
        if opts.SeparateFlowFiles {
            output.FlowFiles[sanitizeName(flowName)+".tf"] = flowHCL
        } else {
            output.MainFile += "\n" + flowHCL
        }
        
        // Preserve original export
        if opts.PreserveOriginalExports {
            originalJSON, _ := json.Marshal(flowData)
            output.Assets["flows/"+sanitizeName(flowName)+"_export.json"] = originalJSON
        }
    }
    
    // Generate migration guide
    output.VariableFiles["MIGRATION_GUIDE.md"] = GenerateMigrationGuide(graph)
    
    return output, nil
}
```

**Usage:**
```go
// cmd/export.go or cmd/davinci_to_hcl.go
func runMultiFlowExport() {
    multiFlowJSON, _ := os.ReadFile("export.json")
    
    opts := converter.ExportOptions{
        PreserveOriginalExports: true,
        SeparateFlowFiles:       true,
        InlineGraphData:         true,  // Required for new provider
    }
    
    output, err := converter.ConvertMultiFlowToFiles(multiFlowJSON, opts)
    
    // Write all files
    os.WriteFile(outputPath+"/main.tf", []byte(output.MainFile), 0644)
    
    for name, content := range output.FlowFiles {
        os.WriteFile(outputPath+"/"+name, []byte(content), 0644)
    }
    
    for name, content := range output.VariableFiles {
        os.WriteFile(outputPath+"/"+name, []byte(content), 0644)
    }
    
    // Write assets (original exports)
    os.MkdirAll(outputPath+"/flows", 0755)
    for path, data := range output.Assets {
        os.WriteFile(outputPath+"/"+path, data, 0644)
    }
}
```

**Generated Structure:**
```
output/
├── main.tf                          # Shared connectors
├── variables.tf                     # Input variables
├── outputs.tf                       # Output values
├── MIGRATION_GUIDE.md              # UUID -> Resource mapping
├── flows/
│   ├── main_flow_export.json       # Original export
│   ├── subflow_1_export.json
│   └── subflow_2_export.json
├── flow_main_flow.tf               # Main flow HCL
├── flow_subflow_1.tf               # Subflow 1 HCL
└── flow_subflow_2.tf               # Subflow 2 HCL
```

**Key Improvements:**
- ✅ Dependency-ordered flow generation
- ✅ Shared resource deduplication
- ✅ Separate files for maintainability
- ✅ Original exports preserved
- ✅ Migration guide for UUID tracking

#### 2.2 Dependency Graph Enhancement

**Current Status:**
```go
// internal/resolver/dependency_graph.go
type DependencyGraph struct {
    resources map[string]map[string]string  // type -> uuid -> name
}
```

**Enhancement: Cross-Flow Dependency Tracking**

```go
// Enhanced graph with relationships
type DependencyGraph struct {
    resources     map[string]map[string]string  // type -> uuid -> name
    dependencies  map[string][]Dependency       // resource -> dependencies
    flowRelations map[string]FlowRelation       // flowID -> parent/child relationships
}

type Dependency struct {
    FromResource   string  // "pingone_davinci_flow.main_flow"
    ToResource     string  // "pingone_davinci_flow.subflow_1"
    DependencyType string  // "subflow", "connector", "variable"
    OriginalUUID   string  // UUID from export
}

type FlowRelation struct {
    FlowID       string
    FlowName     string
    ParentFlows  []string  // Flows that call this as subflow
    ChildFlows   []string  // Subflows this flow calls
    Depth        int       // Nesting level (0 = leaf subflow, 1+ = parent)
}

// Build flow hierarchy
func (g *DependencyGraph) BuildFlowHierarchy() error {
    // Analyze all subflow relationships
    // Detect circular dependencies
    // Calculate depth for ordering
}

// Topological sort for proper ordering
func (g *DependencyGraph) TopologicalSort() []string {
    sorted := []string{}
    visited := make(map[string]bool)
    
    // Sort by depth first (leaf subflows first)
    // Then alphabetically within same depth
    
    return sorted
}

// Generate Terraform depends_on blocks
func (g *DependencyGraph) GetDependsOn(resourceName string) []string {
    deps := []string{}
    
    for _, dep := range g.dependencies[resourceName] {
        if dep.DependencyType == "subflow" || dep.DependencyType == "variable" {
            deps = append(deps, dep.ToResource)
        }
    }
    
    return deps
}
```

**Usage in Generation:**
```go
func ConvertFlowToHCL(flowData map[string]interface{}, graph *DependencyGraph) (string, error) {
    // ... existing conversion
    
    // Add depends_on if flow has dependencies
    dependsOn := graph.GetDependsOn(flowResourceName)
    if len(dependsOn) > 0 {
        hcl.WriteString("\n")
        hcl.WriteString("  depends_on = [\n")
        for _, dep := range dependsOn {
            hcl.WriteString(fmt.Sprintf("    %s,\n", dep))
        }
        hcl.WriteString("  ]\n")
    }
}
```

---

### Priority 3: Module Support

#### 3.1 Terraform Variable File Generation (MEDIUM)

**Problem**: No `variables.tf` or `locals.tf` generated - output not module-ready.

**Solution: Auto-Generate Module Structure**

```go
// New: ModuleStructure generator
type ModuleStructure struct {
    Variables    []TerraformVariable
    Locals       map[string]string
    Outputs      []TerraformOutput
    RequiredProviders string
}

type TerraformVariable struct {
    Name        string
    Type        string
    Description string
    Default     interface{}
    Sensitive   bool
}

type TerraformOutput struct {
    Name        string
    Description string
    Value       string
    Sensitive   bool
}

func GenerateModuleStructure(graph *DependencyGraph, opts ExportOptions) *ModuleStructure {
    module := &ModuleStructure{
        Variables: []TerraformVariable{},
        Locals:    make(map[string]string),
        Outputs:   []TerraformOutput{},
    }
    
    // Core variable: environment_id
    module.Variables = append(module.Variables, TerraformVariable{
        Name:        "environment_id",
        Type:        "string",
        Description: "PingOne Environment ID for DaVinci resources",
    })
    
    // Optional: region_code for multi-region
    if opts.IncludeRegionCode {
        module.Variables = append(module.Variables, TerraformVariable{
            Name:        "region_code",
            Type:        "string",
            Description: "PingOne region code (NA, EU, AP, CA, AU)",
            Default:     "NA",
        })
    }
    
    // Generate outputs for flows
    for flowName, flowID := range graph.GetFlowsByType("pingone_davinci_flow") {
        module.Outputs = append(module.Outputs, TerraformOutput{
            Name:        sanitizeName(flowName) + "_id",
            Description: fmt.Sprintf("ID of the %s flow", flowName),
            Value:       fmt.Sprintf("pingone_davinci_flow.%s.id", sanitizeName(flowName)),
        })
    }
    
    // Required providers
    module.RequiredProviders = `
terraform {
  required_version = ">= 1.5"
  
  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = ">= 1.1.0"
    }
  }
}
`
    
    return module
}

// Generate variables.tf content
func (m *ModuleStructure) GenerateVariablesFile() string {
    var hcl strings.Builder
    
    hcl.WriteString("# Generated by davinci-terraform-converter\n")
    hcl.WriteString("# Input variables for DaVinci flows module\n\n")
    
    for _, v := range m.Variables {
        hcl.WriteString(fmt.Sprintf("variable \"%s\" {\n", v.Name))
        hcl.WriteString(fmt.Sprintf("  type        = %s\n", v.Type))
        hcl.WriteString(fmt.Sprintf("  description = \"%s\"\n", v.Description))
        
        if v.Default != nil {
            hcl.WriteString(fmt.Sprintf("  default     = \"%v\"\n", v.Default))
        }
        
        if v.Sensitive {
            hcl.WriteString("  sensitive   = true\n")
        }
        
        hcl.WriteString("}\n\n")
    }
    
    return hcl.String()
}

// Generate outputs.tf content
func (m *ModuleStructure) GenerateOutputsFile() string {
    var hcl strings.Builder
    
    hcl.WriteString("# Generated by davinci-terraform-converter\n")
    hcl.WriteString("# Output values from DaVinci flows module\n\n")
    
    for _, o := range m.Outputs {
        hcl.WriteString(fmt.Sprintf("output \"%s\" {\n", o.Name))
        hcl.WriteString(fmt.Sprintf("  description = \"%s\"\n", o.Description))
        hcl.WriteString(fmt.Sprintf("  value       = %s\n", o.Value))
        
        if o.Sensitive {
            hcl.WriteString("  sensitive   = true\n")
        }
        
        hcl.WriteString("}\n\n")
    }
    
    return hcl.String()
}

// Generate versions.tf content
func (m *ModuleStructure) GenerateVersionsFile() string {
    return m.RequiredProviders
}

// Generate README.md content
func (m *ModuleStructure) GenerateReadme(metadata ExportMetadata) string {
    return fmt.Sprintf(`# DaVinci Flows Module

Generated by davinci-terraform-converter

## Resources Created

- %d flows
- %d connector instances
- %d variables

## Usage

` + "```" + `hcl
module "davinci_flows" {
  source = "./path/to/this/module"
  
  environment_id = pingone_environment.my_env.id
}

# Access flow IDs via outputs
resource "example" "use_flow" {
  flow_id = module.davinci_flows.main_flow_id
}
` + "```" + `

## Requirements

- Terraform >= 1.5
- PingOne provider >= 1.1.0

## Flow Dependency Order

Generated flows are ordered to respect dependencies:

%s

## Original Exports

Original DaVinci export JSON files are preserved in ` + "`flows/`" + ` directory for reference.
`,
        len(metadata.FlowOrder),
        len(metadata.SharedConnectors),
        len(metadata.SharedVariables),
        strings.Join(metadata.FlowOrder, " → "),
    )
}
```

**Integration with Export:**
```go
// cmd/export.go
func runExport() {
    // ... existing export logic
    
    // Generate module structure
    moduleStruct := converter.GenerateModuleStructure(graph, opts)
    
    // Write module files
    os.WriteFile(outputPath+"/variables.tf", []byte(moduleStruct.GenerateVariablesFile()), 0644)
    os.WriteFile(outputPath+"/outputs.tf", []byte(moduleStruct.GenerateOutputsFile()), 0644)
    os.WriteFile(outputPath+"/versions.tf", []byte(moduleStruct.GenerateVersionsFile()), 0644)
    os.WriteFile(outputPath+"/README.md", []byte(moduleStruct.GenerateReadme(metadata)), 0644)
    
    logger.Success("Generated Terraform module structure")
}
```

**Generated Module Structure:**
```
davinci-module/
├── versions.tf        # Terraform & provider versions
├── variables.tf       # Input variables
├── outputs.tf         # Output values
├── README.md          # Usage documentation
├── main.tf            # Shared resources (connectors)
├── flows/
│   ├── *.tf           # Flow definitions
│   └── *_export.json  # Original exports
└── MIGRATION_GUIDE.md # UUID mapping reference
```

**Example Generated Files:**

**variables.tf:**
```hcl
# Generated by davinci-terraform-converter
variable "environment_id" {
  type        = string
  description = "PingOne Environment ID for DaVinci resources"
}
```

**outputs.tf:**
```hcl
# Generated by davinci-terraform-converter
output "main_flow_id" {
  description = "ID of the Main Flow"
  value       = pingone_davinci_flow.main_flow.id
}

output "subflow_1_id" {
  description = "ID of the Subflow 1"
  value       = pingone_davinci_flow.subflow_1.id
}
```

**versions.tf:**
```hcl
terraform {
  required_version = ">= 1.5"
  
  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = ">= 1.1.0"
    }
  }
}
```

---

### Priority 4: Pipe Support

#### 4.1 Add stdin Detection (LOW - but high impact on UX)

**Solution:**
```go
// cmd/davinci_to_hcl.go
func (c *DaVinciToHclCommand) Run(args []string, logger grpc.Logger) error {
    flags := pflag.NewFlagSet("davinci-to-hcl", pflag.ContinueOnError)
    
    flowJSON := flags.String("flow-json", "", "Path to input file (or use stdin)")
    out := flags.String("out", "", "Output file (defaults to stdout)")
    skipDeps := flags.Bool("skip-dependencies", false, "Skip dependency references")
    
    flags.Parse(args)
    
    // Check if stdin has data (pipe)
    var flowJSONBytes []byte
    var err error
    
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) == 0 {
        // stdin is pipe
        logger.Message("Reading flow JSON from stdin (pipe)", nil)
        flowJSONBytes, err = io.ReadAll(os.Stdin)
        if err != nil {
            return fmt.Errorf("failed to read from stdin: %w", err)
        }
    } else if *flowJSON != "" {
        // File-based
        logger.Message(fmt.Sprintf("Reading flow JSON from file: %s", *flowJSON), nil)
        flowJSONBytes, err = os.ReadFile(*flowJSON)
        if err != nil {
            return fmt.Errorf("failed to read file: %w", err)
        }
    } else {
        return fmt.Errorf("no input provided: use --flow-json flag or pipe JSON to stdin")
    }
    
    // Convert
    hcl, err := converter.ConvertWithOptions(flowJSONBytes, *skipDeps)
    if err != nil {
        return fmt.Errorf("conversion failed: %w", err)
    }
    
    // Output
    if *out != "" {
        os.WriteFile(*out, []byte(hcl), 0644)
        logger.Success(fmt.Sprintf("Wrote HCL to: %s", *out), nil)
    } else {
        fmt.Println(hcl)
    }
    
    return nil
}
```

**Usage Examples:**
```bash
# Pipe from cat
cat flow.json | davinci-convert davinci-to-hcl

# Pipe from jq (preprocess)
jq '.flows[0]' multi-flow.json | davinci-convert davinci-to-hcl

# Pipe from curl (direct API export)
curl -H "Authorization: Bearer $TOKEN" \
  https://api.pingone.com/.../flows/abc123 | \
  davinci-convert davinci-to-hcl -o flow.tf

# File-based (existing)
davinci-convert davinci-to-hcl --flow-json flow.json

# Output to file
cat flow.json | davinci-convert davinci-to-hcl -o flow.tf
```

**Export Command Enhancement:**
```go
// cmd/export.go - also support pipe for multi-flow exports
func (c *ExportCommand) Run(args []string, logger grpc.Logger) error {
    // ... existing flag parsing
    
    // Check for piped multi-flow JSON
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) == 0 {
        logger.Message("Reading multi-flow export from stdin", nil)
        multiFlowJSON, _ := io.ReadAll(os.Stdin)
        
        // Process piped data instead of API fetch
        return processMultiFlowExport(multiFlowJSON, outputPath, skipDeps)
    }
    
    // Otherwise, fetch from API (existing behavior)
    return runAPIExport(...)
}
```

---

## Implementation Roadmap

### Phase 1: Critical Alignment (Week 1-2)
- [ ] **1.1** External file preservation (Option B)
- [ ] **1.2** UUID mapping guide generation
- [ ] **2.1** Multi-flow unified output structure
- [ ] **4.1** Pipe support for davinci-to-hcl

**Deliverables:**
- Preserve original exports in `flows/` directory
- Generate `MIGRATION_GUIDE.md` with UUID mappings
- Multi-file output for multi-flow exports
- Pipe support: `cat flow.json | davinci-convert davinci-to-hcl`

### Phase 2: Module Support (Week 3)
- [ ] **3.1** Generate `variables.tf`, `outputs.tf`, `versions.tf`
- [ ] **3.1** Generate module `README.md`
- [ ] **2.2** Enhanced dependency graph with flow hierarchy

**Deliverables:**
- Complete Terraform module structure
- Module usage documentation
- Flow dependency visualization

### Phase 3: Advanced Features (Week 4+)
- [ ] **1.1** Explore external file reference (Option A) with provider team
- [ ] **2.2** Circular dependency detection
- [ ] **3.1** Generate example usage files
- [ ] Integration tests with real multi-flow exports

---

## Success Metrics

### Consumer Migration Effort Reduction

**Before Improvements:**
1. Run `davinci-convert export` → Get single file with all flows
2. Manually split into separate files
3. Manually track UUID → resource mappings
4. Manually add `depends_on` for flow ordering
5. Manually create `variables.tf`, `outputs.tf`
6. Manually debug dependency issues
7. **Estimated time: 4-8 hours for complex export**

**After Improvements:**
1. Run `davinci-convert export --out ./my-module`
2. Get complete module structure with:
   - Separate flow files
   - Dependency-ordered resources
   - UUID mapping guide
   - Module boilerplate
3. Review `MIGRATION_GUIDE.md` for mappings
4. Adjust variable values if needed
5. **Estimated time: 30 minutes for complex export**

**Target: 80% reduction in manual work**

### Compatibility Metrics

| Feature | Legacy CLI | Current | Target |
|---------|-----------|---------|--------|
| External flow files | ✅ | ❌ | ⚠️ (preserved as docs) |
| UUID mapping | ✅ | ❌ | ✅ |
| Multi-flow handling | ✅ | ⚠️ | ✅ |
| Module support | ⚠️ | ❌ | ✅ |
| Pipe support | ✅ | ❌ | ✅ |
| Dependency ordering | ✅ | ⚠️ | ✅ |

---

## Testing Strategy

### Test Cases

#### 1. Multi-Flow Export
```go
func TestMultiFlowExportWithDependencies(t *testing.T) {
    // Test data: main flow + 2 subflows
    multiFlowJSON := loadTestData("multi-flow-export.json")
    
    output, err := ConvertMultiFlowToFiles(multiFlowJSON, defaultOpts)
    require.NoError(t, err)
    
    // Verify flow ordering
    assert.Equal(t, []string{"subflow_1", "subflow_2", "main_flow"}, output.Metadata.FlowOrder)
    
    // Verify main.tf has shared connectors
    assert.Contains(t, output.MainFile, "resource \"pingone_davinci_connector_instance\"")
    
    // Verify flows reference subflows correctly
    mainFlowHCL := output.FlowFiles["flow_main_flow.tf"]
    assert.Contains(t, mainFlowHCL, "pingone_davinci_flow.subflow_1.id")
    
    // Verify depends_on present
    assert.Contains(t, mainFlowHCL, "depends_on")
    
    // Verify assets preserved
    assert.NotEmpty(t, output.Assets["flows/main_flow_export.json"])
}
```

#### 2. Pipe Support
```bash
# Test script
cat test-flow.json | davinci-convert davinci-to-hcl > output.tf
jq '.flows[0]' multi-flow.json | davinci-convert davinci-to-hcl > flow1.tf
```

#### 3. Module Structure
```go
func TestModuleStructureGeneration(t *testing.T) {
    graph := buildTestGraph()
    module := GenerateModuleStructure(graph, defaultOpts)
    
    // Verify variables
    assert.Contains(t, module.Variables, "environment_id")
    
    // Verify outputs
    assert.NotEmpty(t, module.Outputs)
    
    // Verify files
    varsHCL := module.GenerateVariablesFile()
    assert.Contains(t, varsHCL, "variable \"environment_id\"")
}
```

---

## Documentation Updates

### User-Facing Docs

#### README.md Updates
```markdown
## Migration from Legacy Provider

If migrating from the legacy `davinci` provider, this tool generates:

1. **Module Structure**: Complete Terraform module with variables/outputs
2. **UUID Mapping Guide**: `MIGRATION_GUIDE.md` tracks original UUIDs
3. **Original Exports**: Preserved in `flows/` for reference
4. **Dependency Ordering**: Flows ordered to respect subflow dependencies

See [docs/MIGRATION_FROM_LEGACY.md](docs/MIGRATION_FROM_LEGACY.md) for details.
```

#### New: docs/MIGRATION_FROM_LEGACY.md
```markdown
# Migrating from Legacy DaVinci Provider

## Key Differences

### Legacy Provider (davinci)
- Used external JSON files: `flow_json = file("flow.json")`
- Required `connection_link` and `subflow_link` blocks
- Tracked UUIDs with `replace_import_*_id`

### New Provider (pingone)
- Embeds flows inline in HCL
- Direct references: `connection_id = resource.id`
- No explicit mapping blocks

## Using the Migration Guide

Generated exports include `MIGRATION_GUIDE.md`:

| Original UUID | New Resource |
|---------------|--------------|
| 867ed...817d  | pingone_davinci_connector_instance.http.id |

Use this to:
1. Verify all connections mapped correctly
2. Debug "connection not found" errors
3. Update external references in other Terraform modules
```

---

## Conclusion

### Summary of Critical Improvements

1. **UUID Mapping Guide**: Bridge gap between external file + mapping blocks (legacy) and inline references (new)
2. **Multi-Flow Structure**: Generate cohesive multi-file module instead of array of strings
3. **Module Boilerplate**: Auto-generate variables.tf, outputs.tf, versions.tf for immediate use
4. **Pipe Support**: Enable preprocessing workflows with jq, curl, etc.

### Migration Effort Reduction

**Target**: 80% reduction in manual work for consumers migrating from legacy provider.

**Before**: 4-8 hours of manual file splitting, UUID tracking, dependency ordering.
**After**: 30 minutes reviewing generated module, adjusting variable values.

### Next Steps

1. Review improvement plan with team
2. Prioritize Phase 1 (critical alignment)
3. Create GitHub issues for each improvement
4. Begin implementation in 1-week sprints
5. Test with real-world legacy exports from customers
