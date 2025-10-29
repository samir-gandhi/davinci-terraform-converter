# Feature 1: Module-Ready Variable Scaffolding

## Objective

Transform the CLI from generating standalone resources to producing a **child module** with configurable variables. Enable two primary use cases:

1. **Shareable Module** - Generic module with variables, no hardcoded values, ready for reuse
2. **Environment Management** - Module with import blocks and actual values for managing existing environments

## Architecture Overview

### Directory Structure

```
output-directory/
├── module.tf                      # Root module - references child module
├── imports.tf                     # Import blocks (--include-imports only)
└── davinci-module/                # Child module (--module-dir name)
    ├── versions.tf                # Provider requirements
    ├── variables.tf               # Variable definitions
    ├── flows.tf                   # Flow resources
    ├── connections.tf             # Connection resources
    ├── variables_dv.tf            # DaVinci variable resources
    └── outputs.tf                 # Output values
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--module` | `true` | Generate module structure (default mode) |
| `--module-dir` | `davinci-module` | Name of child module directory |
| `--include-imports` | `false` | Generate import blocks in root module |
| `--include-values` | `false` | Populate variable inputs with actual values from export |

### Use Case 1: Shareable Module

**Command:**
```bash
davinci-terraform-converter export --module-dir ./davinci
```

**Result:**
- Child module with all resources
- Variables defined but no values in `module.tf`
- Consumer provides values when using module
- No import blocks

**module.tf:**
```hcl
module "davinci" {
  source = "./davinci-module"
  
  pingone_environment_id = var.target_environment_id
  
  # Variables without values - consumer fills in
  davinci_variable_company_name_value = ""
  davinci_variable_session_timeout_value = 0
  davinci_connection_http_recaptcha_secret = ""  # TODO: Provide secret value
}
```

### Use Case 2: Environment Management

**Command:**
```bash
davinci-terraform-converter export \
  --module-dir ./davinci-module \
  --include-imports \
  --include-values
```

**Result:**
- Child module with all resources
- Import blocks in `imports.tf` pointing to child module resources
- Variable inputs populated with actual values from API
- Secrets have comments instead of values

**module.tf:**
```hcl
module "davinci" {
  source = "./davinci-module"
  
  pingone_environment_id = "abc-123-def-456"
  
  # Actual values from export
  davinci_variable_company_name_value = "Acme Corp"
  davinci_variable_session_timeout_value = 5
  davinci_connection_http_recaptcha_secret = ""  # TODO: Provide secret value
}
```

**imports.tf:**
```hcl
import {
  to = module.davinci.pingone_davinci_flow.main_flow
  id = "abc-123-def-456:flow-id-789"
}

import {
  to = module.davinci.pingone_davinci_variable.company_name
  id = "abc-123-def-456:var-id-123"
}
```

## Current State

The CLI generates a single file with hardcoded values:

```hcl
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id
  name           = "Main Flow"
  description    = "Hardcoded description"
  # ...
}
```

**Problems:**

- Not reusable across environments
- No variable definitions
- No module structure
- Cannot manage existing resources (no imports)
- Hardcoded values throughout

## Target State

### Child Module Structure

Generate a complete Terraform child module with:

1. **versions.tf** - Provider requirements matching CLI's Terraform version
2. **variables.tf** - Variable definitions for all configurable attributes
3. **Resource files** - Resources referencing variables
4. **outputs.tf** - Output values for downstream use

### Variable-Ready Attributes

The converter schema needs flexibility to mark which resource attributes should be variables. Examples:

**DaVinci Variables:**
- `name` → hardcoded
- `value` → variable
- `type` → hardcoded

**Flows:**
- `environment_id` → variable (inherited from module input)
- `name` → hardcoded
- `description` → hardcoded

**connector_instance:**
- `environment_id` → variable (inherited)
- `name` → hardcoded
- `property.value` → variable (for configurable properties)

### Root Module Files

**module.tf** - Always generated:
```hcl
module "davinci" {
  source = "./davinci-module"
  
  # Core inputs
  pingone_environment_id = ""  # Empty without --include-values
  
  # Resource-specific inputs (empty by default)
  davinci_variable_company_name_value = ""
  davinci_flow_main_name = ""
}
```

**imports.tf** - Generated with `--include-imports`:
```hcl
import {
  to = module.davinci.pingone_davinci_flow.main_flow
  id = "env-id:flow-id"
}
```

## Implementation Approach

### Phase 1: Schema Enhancement

**Goal:** Make converter schemas flexible for variable assignment

1. **Update Resource Schemas:**
   - Add `variable_eligible: bool` field to attribute definitions
   - Mark which attributes should become variables
   - Define variable naming conventions

2. **Example Schema Change:**
   ```go
   type AttributeConfig struct {
       Name            string
       Type            string
       Required        bool
       VariableEligible bool    // NEW: Can this become a variable?
       VariableName    string   // NEW: Optional custom variable name
       IsSecret        bool     // NEW: Should value be omitted in module.tf?
   }
   ```

3. **Define Variable Eligibility Rules:**
   - Primitive types (string, number, bool) → eligible
   - Complex types (objects, lists) → not eligible
   - IDs and computed values → not eligible
   - Names, descriptions, values → eligible

### Phase 2: Module Generation Engine

**Goal:** Generate child module structure

1. **Create Module Generator:**
   ```go
   type ModuleGenerator struct {
       OutputDir      string
       ModuleDirName  string
       IncludeImports bool
       IncludeValues  bool
   }
   
   func (m *ModuleGenerator) Generate(graph *DependencyGraph) error {
       // Create directory structure
       // Generate versions.tf
       // Generate variables.tf
       // Generate resource files
       // Generate outputs.tf
       // Generate root module.tf
       // Generate imports.tf (if enabled)
   }
   ```

2. **versions.tf Generator:**
   - Detect Terraform version from runtime
   - Pin provider version to tested range
   - Include required_providers block

3. **variables.tf Generator:**
   - Extract all variable-eligible attributes
   - Generate variable blocks with:
     - Type
     - Description
     - Default (null or omitted)
     - Validation (where applicable)
   - Add sensitive = true for secrets

4. **Resource File Generator:**
   - Replace hardcoded values with variable references
   - Format: `var.{variable_name}`
   - Maintain resource structure otherwise

### Phase 3: Root Module Generation

**Goal:** Generate module.tf with variable inputs

1. **module.tf Generator:**
   - List all child module variables
   - Create input assignments
   - Without `--include-values`: empty strings/zeros
   - With `--include-values`: actual values from export
   - Secrets: empty with `# TODO: Provide value` comment

2. **Example Logic:**
   ```go
   func generateModuleInput(variable Variable, includeValues bool) string {
       if !includeValues {
           return fmt.Sprintf("%s = \"\"", variable.Name)
       }
       
       if variable.IsSecret {
           return fmt.Sprintf("%s = \"\"  # TODO: Provide secret value", variable.Name)
       }
       
       return fmt.Sprintf("%s = %s", variable.Name, formatValue(variable.Value))
   }
   ```

### Phase 4: Import Block Generation

**Goal:** Generate imports.tf for existing environments

1. **Import Block Generator:**
   - Only runs with `--include-imports` flag
   - Creates import block for each resource
   - Format: `module.{module_name}.{resource_type}.{resource_name}`
   - ID format: `{environment_id}:{resource_id}`

2. **Import ID Format per Resource Type:**
   - Flows: `{env_id}:{flow_id}`
   - Variables: `{env_id}:{variable_id}`
   - Connections: `{env_id}:{connection_id}`
   - Applications: `{env_id}:{application_id}`

### Phase 5: CLI Integration

**Goal:** Wire up new flags and orchestrate generation

1. **Add Flags:**
   ```go
   exportCmd.Flags().Bool("module", true, "Generate module structure")
   exportCmd.Flags().String("module-dir", "davinci-module", "Child module directory name")
   exportCmd.Flags().Bool("include-imports", false, "Generate import blocks")
   exportCmd.Flags().Bool("include-values", false, "Include actual values from export")
   ```

2. **Update Export Flow:**
   ```go
   func runExport(cmd *cobra.Command, args []string) error {
       // Existing: Fetch data from API
       // Existing: Build dependency graph
       
       // NEW: Generate module structure
       moduleGen := NewModuleGenerator(
           outputDir,
           moduleDirName,
           includeImports,
           includeValues,
       )
       
       return moduleGen.Generate(graph)
   }
   ```

## Success Criteria

- [ ] CLI generates child module directory structure
- [ ] versions.tf created with correct provider versions
- [ ] variables.tf contains all variable-eligible attributes
- [ ] Resources reference variables instead of hardcoded values
- [ ] module.tf created at root level
- [ ] module.tf inputs empty by default
- [ ] `--include-values` populates module.tf with actual values
- [ ] Secrets have TODO comments instead of values
- [ ] `--include-imports` generates imports.tf
- [ ] Import blocks point to child module resources
- [ ] Generated module passes `terraform init` and `terraform validate`
- [ ] Module can be used in separate root module

## Testing Strategy

### Unit Tests

1. **Schema Tests:**
   - Test variable eligibility detection
   - Test variable name generation
   - Test secret detection

2. **Generator Tests:**
   - Test versions.tf generation
   - Test variables.tf generation
   - Test module.tf generation with/without values
   - Test imports.tf generation

### Integration Tests

1. **Shareable Module Test:**
   ```bash
   # Generate without values
   davinci-terraform-converter export --env-id test-env --output ./test-output
   
   # Verify structure
   test -f ./test-output/module.tf
   test -d ./test-output/davinci-module
   test -f ./test-output/davinci-module/versions.tf
   test -f ./test-output/davinci-module/variables.tf
   
   # Verify no values in module.tf
   ! grep -q "= \"actual-value\"" ./test-output/module.tf
   
   # Verify Terraform works
   cd ./test-output
   terraform init
   terraform validate
   ```

2. **Environment Management Test:**
   ```bash
   # Generate with imports and values
   davinci-terraform-converter export \
     --env-id test-env \
     --output ./test-output \
     --include-imports \
     --include-values
   
   # Verify imports.tf exists
   test -f ./test-output/imports.tf
   
   # Verify values in module.tf
   grep -q "= \"actual-value\"" ./test-output/module.tf
   
   # Verify import blocks reference module
   grep -q "module.davinci" ./test-output/imports.tf
   ```

3. **Secret Handling Test:**
   - Generate module with secret values
   - Verify secrets have TODO comments
   - Verify secrets not exposed in module.tf

## Design Decisions

### Schema Design Questions

1. **Which attributes should be variable-eligible?**
   - All primitive types on all resources?
   - Only specific attributes (names, descriptions, values)?
   - User-configurable via schema annotations?

2. **How to handle nested attributes?**
   - Flatten nested attributes into variables?
   - Keep complex structures hardcoded?
   - Support selective variable exposure?

3. **Variable naming conventions:**
   - Format: `{resource_type}_{resource_name}_{attribute_name}`?
   - Should resource name be sanitized?
   - How to handle name collisions?

4. **Secret detection:**
   - Maintain list of known secret property names?
   - Use naming patterns (contains "secret", "key", "token")?
   - Schema annotation?

5. **Default values:**
   - Should variables.tf have defaults from export?
   - Or should defaults be null/empty?
   - Should module.tf defaults differ from variables.tf defaults?

### Open Questions

1. Should `--module` flag exist or is module mode always default?
2. Should standalone mode still be supported with a flag?
3. What provider version range to use in versions.tf?
4. Should outputs.tf be auto-generated for all resources?
5. How to handle resources that shouldn't be in child module?
6. Should there be a way to exclude certain variables from generation?
7. How to version the child module itself?
8. Should schema allow custom variable templates per resource type?

## References

- See `MISSING_FEATURES_FROM_LEGACY.md` for detailed comparison with legacy CLI
- Legacy implementation patterns for property extraction
- Connector schema handling approaches

### 1. Core Module Variables (`vars.tf`)## Design Decisions
