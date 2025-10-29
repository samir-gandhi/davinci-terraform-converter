# Missing Features: dvtf-pingctl vs davinci-terraform-converter

## Analysis Date
October 24, 2025

## Executive Summary

The legacy `dvtf-pingctl` tool generates a **comprehensive, production-ready Terraform module** with extensive variable scaffolding, validation, and documentation. The new `davinci-terraform-converter` currently outputs raw HCL without module structure. This document identifies critical missing features that significantly increase consumer post-generation work.

**Key Finding**: dvtf-pingctl reduces consumer manual work by generating **~15,000 lines of boilerplate** that consumers otherwise must create manually.

---

## Generated File Structure Comparison

### Legacy CLI Output (dvtf-pingctl)

```
generated/
├── README.md                               # 4,328 lines - Module documentation
├── versions.tf                             # Provider version constraints
├── vars.tf                                 # Core environment_id variable with validation
├── davinci_connection_property_vars.tf     # 312 lines - Connection property variables
├── davinci_variable_vars.tf                # Variable value overrides
├── davinci_flow_vars.tf                    # 11,701 lines - Flow configuration variables
├── davinci_flow_outputs.tf                 # 3,601 lines - Flow output values
├── davinci_connectors.tf                   # 407 lines - Connection resources
├── davinci_flows.tf                        # 18,076 lines - Flow resources
├── davinci_variables.tf                    # Variable resources
└── assets/
    └── flows/
        ├── flow1.json                      # Preserved original exports
        ├── flow2.json
        └── ...
```

**Total Generated**: ~38,000+ lines of production-ready code

### New CLI Output (davinci-terraform-converter)

```
output.tf                                    # Single file with inline HCL
```

**Total Generated**: Variable (depends on flow size, typically 500-5,000 lines)

**Missing**: Module structure, variables, outputs, documentation, validation

---

## Feature Gap Analysis

### 1. **Module-Ready Variable Scaffolding** (CRITICAL)

#### What dvtf-pingctl Generates

**A. Core Module Input (`vars.tf`)**:
```hcl
variable "pingone_environment_id" {
  description = "The PingOne environment ID to configure DaVinci resources in"
  type        = string

  validation {
    condition     = can(regex("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", var.pingone_environment_id))
    error_message = "The PingOne Environment ID must be a valid PingOne resource ID (UUID format)."
  }
}
```

**Benefits**:
- ✅ Runtime UUID validation
- ✅ Clear error messages for invalid input
- ✅ Self-documenting interface

**B. Connection Property Variables (`davinci_connection_property_vars.tf` - 312 lines)**:
```hcl
// Properties for the "Http" connector, with connector ID httpConnector.
// Terraform Resource: davinci_connection.httpconnector__867ed4363b2bc21c860085ad2baa817d

variable "davinci_connection_httpconnector__867ed4363b2bc21c860085ad2baa817d_recaptchaSecretKey" {
  type = string
  
  description = <<EOT
The 'reCAPTCHA v2 Secret Key' property for the connector named 'Http' with connector ID 'httpConnector'. 
The Secret Key from reCAPTCHA Admin dashboard.
EOT
  default = null
}

variable "davinci_connection_httpconnector__867ed4363b2bc21c860085ad2baa817d_recaptchaSiteKey" {
  type = string
  
  description = <<EOT
The 'reCAPTCHA v2 Site Key' property for the connector named 'Http' with connector ID 'httpConnector'. 
The Site Key from reCAPTCHA Admin dashboard.
EOT
  default = null
}

variable "davinci_connection_httpconnector__867ed4363b2bc21c860085ad2baa817d_whiteList" {
  type = string
  
  description = <<EOT
The 'Trusted Sites' property for the connector named 'Http' with connector ID 'httpConnector'. 
Enter the hostname for the trusted sites that host your HTML.
EOT
  default = null
}
```

**Benefits**:
- ✅ Every connector property exposed as variable
- ✅ Inline documentation from connector schema
- ✅ Type safety (string, bool, number)
- ✅ Optional (default = null) for flexibility
- ✅ Consumer can override without editing resources

**How It Works** (Implementation Deep Dive):

1. **Schema Embedding**: dvtf-pingctl embeds a complete connector schema JSON file at compile time:
   ```go
   // internal/generate/connector_schema.go
   //go:embed connector_schema/connector-schema.json
   var connectorSchemaBytes []byte
   ```

2. **Schema Structure**: The embedded JSON defines all connectors with property metadata:
   ```json
   [
     {
       "name": "Http",
       "connectorId": "httpConnector",
       "properties": {
         "recaptchaSecretKey": {
           "type": "string",
           "displayName": "reCAPTCHA v2 Secret Key",
           "preferredControlType": "textField",
           "info": "The Secret Key from reCAPTCHA Admin dashboard."
         },
         "customAuth": {
           "type": "array",
           "displayName": "Custom Parameters",
           "info": "Custom authentication parameters"
         }
       }
     }
   ]
   ```

3. **Property Extraction**: Function parses schema and extracts properties per connector:
   ```go
   // internal/generate/data_connection.go
   func getConnectionProperties(connectorID string) ([]connectionDataProperty, error) {
       // Unmarshal embedded schema
       var connectors []Connector
       json.Unmarshal(connectorSchemaBytes, &connectors)
       
       // Find matching connector
       for _, conn := range connectors {
           if conn.ConnectorID == connectorID {
               // Extract properties map
               for propName, propDef := range conn.Properties {
                   // Map DaVinci type → Terraform type
                   tfType := getConnectorGeneratorTypes(propDef.Type)
                   
                   properties = append(properties, connectionDataProperty{
                       Name:          propName,
                       DisplayName:   propDef.DisplayName,
                       TerraformType: tfType,  // bool, string, number, or string (for JSON)
                       Description:   propDef.Info,
                   })
               }
           }
       }
       return properties, nil
   }
   ```

4. **Type Mapping**: DaVinci types convert to Terraform types:
   - `boolean` → `bool`
   - `string` → `string`
   - `number` → `number`
   - `object`, `array` → `string` (JSON-encoded)

5. **Template Execution**: Properties flow into Go template:
   ```go
   // internal/generate/generate.go (line 263-271)
   connectionProperties, err := getConnectionProperties(*nodeData.ConnectorID)
   
   d.ConnectionsData = append(d.ConnectionsData, connectionData{
       ResourceName:       resourceName,
       ID:                 *nodeData.ConnectorID,
       Name:               *nodeData.Name,
       Properties:         connectionProperties,  // Array of connectionDataProperty
   })
   
   // Later: writeConnectionsPropertyVars() iterates d.ConnectionsData
   for _, connectionData := range d.ConnectionsData {
       hclTemplate.Execute(outputFile, connectionData)  // Generates variables
   }
   ```

6. **Generated Output**: Template produces variable blocks with metadata:
   ```gotmpl
   {{range $property := .Properties }}
   variable "davinci_connection_{{.ResourceName}}_{{$property.Name}}" {
     type = {{$property.TerraformType}}
     description = <<EOT
   The '{{$property.DisplayName}}' property for '{{.Name}}'. {{$property.Description}}
   EOT
     default = null
   }
   {{end}}
   ```

**Key Insight**: Property definitions come from **embedded static schema**, not runtime API calls. This enables offline generation with full metadata (displayName, description, type) without requiring DaVinci API access during code generation.

**C. Flow Configuration Variables (`davinci_flow_vars.tf` - 11,701 lines)**:
```hcl
// Variables for the "Main Flow" flow.
variable "davinci_flow_main_name" {
  type        = string
  description = "The name of the flow with resource name 'main'."
  default     = "Main Flow"
}

variable "davinci_flow_main_description" {
  type        = string
  description = "The description of the flow with resource name 'main'."
  default     = null
}

variable "davinci_flow_main_log_level" {
  type        = number
  description = "An integer that specifies the log level for the flow. Valid values are: `1` (no logging), `2` (info logging - the default), and `3` (debug logging)."
  default     = null

  validation {
    condition     = var.davinci_flow_main_log_level == null || contains([1, 2, 3], var.davinci_flow_main_log_level)
    error_message = "The value must be one of 1, 2, or 3."
  }
}

// Choice: file path vs raw JSON
variable "davinci_flow_main_json_file_path" {
  type        = string
  description = "The filesystem location of the flow JSON. Cannot be set with the `davinci_flow_main_json` variable."
  default     = "assets/flows/main_flow.json"

  validation {
    condition = (var.davinci_flow_main_json_file_path != null || var.davinci_flow_main_json != null) && 
                (var.davinci_flow_main_json_file_path == null || var.davinci_flow_main_json == null)
    error_message = "Must set either 'davinci_flow_main_json_file_path' or 'davinci_flow_main_json', but not both together."
  }
}

variable "davinci_flow_main_json" {
  type        = string
  description = "A string representing the raw DaVinci import JSON. Cannot be set with the `davinci_flow_main_json_file_path` variable."
  default     = null
}
```

**Benefits**:
- ✅ Flow name/description overridable
- ✅ Log level control per flow
- ✅ Flexible JSON source (file OR inline)
- ✅ Validation prevents configuration errors
- ✅ Enables environment-specific customization

**D. DaVinci Variable Value Variables (`davinci_variable_vars.tf`)**:
```hcl
// Company branding variable
variable "davinci_variable_ciam_companyname_value" {
  type        = string
  description = "The value of the variable with resource name 'ciam_companyname'. Set to null to ensure the variable value doesn't get tracked in Terraform state."
  default     = "Ping Identity"
}

variable "davinci_variable_ciam_logourl_value" {
  type        = string
  description = "URL of company logo"
  default     = "https://assets.pingone.com/ux/ui-library/5.0.2/images/logo-pingidentity.png"
}

variable "davinci_variable_ciam_sessionlengthinminute_value" {
  type        = number
  description = "Session timeout in minutes"
  default     = 5
}

// Feature toggles
variable "davinci_variable_ciam_magiclinkenabled_value" {
  type        = bool
  description = "Enable magic link authentication"
  default     = true
}

variable "davinci_variable_ciam_fidopasskeyenabled_value" {
  type        = bool
  description = "Enable FIDO passkey authentication"
  default     = true
}
```

**Benefits**:
- ✅ Environment-specific configuration
- ✅ Feature flag toggling without editing flows
- ✅ Branding customization (logo, company name)
- ✅ Security settings (session length, MFA toggles)
- ✅ Type-safe (bool, number, string)

#### What davinci-terraform-converter Generates

**Current Output**:
```hcl
resource "pingone_davinci_flow" "main_flow" {
  environment_id = var.environment_id  # Hardcoded variable name, no definition

  name        = "Main Flow"            # Hardcoded
  description = "..."                  # Hardcoded
  
  graph_data {
    # ... inline structure
  }
}
```

**Problems**:
- ❌ No `var.environment_id` definition → runtime error
- ❌ No validation → accepts invalid UUIDs
- ❌ Flow name hardcoded → cannot customize per environment
- ❌ No connector property variables → must edit resource blocks
- ❌ No DaVinci variable overrides → must edit resources
- ❌ Cannot toggle features without flow JSON changes

#### Consumer Impact

**Without Variable Scaffolding** (current state):
```hcl
# Consumer must manually create:
variable "environment_id" {
  type = string
  # Add validation? Documentation? Up to consumer.
}

# To customize company name, must find and edit inline:
resource "pingone_davinci_variable" "company_name" {
  value = "Acme Corp"  # Edit this directly - risky
}

# To change MFA settings, edit connector properties:
resource "pingone_davinci_connector_instance" "mfa" {
  property {
    name  = "policyId"
    value = "abc-123-..."  # Edit this directly
  }
}
```

**With Variable Scaffolding** (legacy approach):
```hcl
# Consumer uses clean module interface:
module "davinci_flows" {
  source = "./generated"
  
  pingone_environment_id = pingone_environment.prod.id
  
  # Customize company branding
  davinci_variable_ciam_companyname_value = "Acme Corp"
  davinci_variable_ciam_logourl_value     = "https://acme.com/logo.png"
  
  # Toggle features
  davinci_variable_ciam_magiclinkenabled_value = false
  
  # Override MFA policy
  davinci_connection_pingonemfaconnector_b72bd_policyId = var.acme_mfa_policy_id
}
```

**Time Saved**: 4-8 hours of manual variable creation for complex environments

---

### 2. ✅ **Comprehensive Output Values** (HIGH)

#### What dvtf-pingctl Generates

**Flow Outputs (`davinci_flow_outputs.tf` - 3,601 lines)**:
```hcl
// Outputs for the "Main Flow" flow.
output "davinci_flow_main" {
  description = "An object that represents the Main Flow, the 'davinci_flow.main' resource."
  
  value = {
    connection_link = davinci_flow.main.connection_link
    description     = davinci_flow.main.description
    environment_id  = davinci_flow.main.environment_id
    id              = davinci_flow.main.id
    log_level       = davinci_flow.main.log_level
    name            = davinci_flow.main.name
    subflow_link    = davinci_flow.main.subflow_link
  }
}

output "davinci_flow_main_id" {
  description = "The ID of the Main Flow"
  value       = davinci_flow.main.id
}

output "davinci_flow_main_name" {
  description = "The name of the Main Flow"
  value       = davinci_flow.main.name
}
```

**Benefits**:
- ✅ Every flow exposed as output
- ✅ Both object (full attributes) and scalar (ID only) outputs
- ✅ Enables downstream resource references
- ✅ Module composition support

#### Usage Example

```hcl
# Root module uses generated outputs
module "davinci_flows" {
  source = "./generated"
  pingone_environment_id = var.env_id
}

# Reference flow IDs in other resources
resource "pingone_application_flow_policy_assignment" "web_app" {
  environment_id = var.env_id
  application_id = pingone_application.web.id
  flow_policy_id = pingone_flow_policy.policy.id
  
  # Use module output
  flow_id = module.davinci_flows.davinci_flow_main_id
}

# Pass to another module
module "monitoring" {
  source = "./monitoring"
  
  # Pass entire flow object
  flows = {
    main     = module.davinci_flows.davinci_flow_main
    recovery = module.davinci_flows.davinci_flow_recovery
  }
}
```

#### What davinci-terraform-converter Generates

**Current Output**: None

**Problems**:
- ❌ Cannot reference flow IDs in other resources
- ❌ Cannot compose with other modules
- ❌ Must manually add outputs for every use case

#### Consumer Impact

**Manual Work Required**:
```hcl
# Consumer must create outputs.tf manually:
output "main_flow_id" {
  value = pingone_davinci_flow.main_flow.id
}

output "subflow_1_id" {
  value = pingone_davinci_flow.subflow_1.id
}

# ... repeat for 20-50 flows
```

**Time Saved**: 1-2 hours for large exports

---

### 3. ✅ **Dynamic Connector Property Blocks** (HIGH)

#### What dvtf-pingctl Generates

**Connector Resources with Dynamic Properties**:
```hcl
resource "davinci_connection" "httpconnector__867ed" {
  environment_id = var.pingone_environment_id
  connector_id   = "httpConnector"
  name           = "Http"
  
  # Dynamic block - only includes non-null properties
  dynamic "property" {
    for_each = concat(
      # reCAPTCHA Secret Key (optional)
      var.davinci_connection_httpconnector__867ed_recaptchaSecretKey != null ? [{
        name  = "recaptchaSecretKey"
        type  = "string"
        value = tostring(var.davinci_connection_httpconnector__867ed_recaptchaSecretKey)
      }] : [],
      
      # reCAPTCHA Site Key (optional)
      var.davinci_connection_httpconnector__867ed_recaptchaSiteKey != null ? [{
        name  = "recaptchaSiteKey"
        type  = "string"
        value = tostring(var.davinci_connection_httpconnector__867ed_recaptchaSiteKey)
      }] : [],
      
      # Trusted Sites (optional)
      var.davinci_connection_httpconnector__867ed_whiteList != null ? [{
        name  = "whiteList"
        type  = "string"
        value = tostring(var.davinci_connection_httpconnector__867ed_whiteList)
      }] : [],
    )
    
    content {
      name  = property.value.name
      type  = property.value.type
      value = property.value.value
    }
  }
}
```

**Benefits**:
- ✅ Variable-driven properties
- ✅ Only includes non-null values
- ✅ Type coercion handled automatically
- ✅ Scales to 20+ properties per connector
- ✅ Consumer overrides via variables only

#### What davinci-terraform-converter Generates

**Current Output**:
```hcl
resource "pingone_davinci_connector_instance" "http" {
  environment_id = var.environment_id
  connector = {
    id = "httpConnector"
  }
  name = "Http"
  
  # TODO: Add property blocks manually
  # See: https://registry.terraform.io/providers/pingidentity/pingone/latest/docs/guides/davinci-connector-reference
}
```

**Problems**:
- ❌ No property blocks generated
- ❌ Consumer must manually lookup connector schemas
- ❌ Hardcoded values in resource blocks
- ❌ No variable-driven configuration

#### Consumer Impact

**Manual Work Required**:
1. Visit connector reference docs
2. Find required/optional properties
3. Add property blocks manually
4. Repeat for 10-50 connectors

**Example Manual Work**:
```hcl
# Consumer must write this for EACH connector:
resource "pingone_davinci_connector_instance" "http" {
  # ... basic config
  
  property {
    name  = "recaptchaSecretKey"
    value = var.recaptcha_secret  # Must create variable manually
  }
  
  property {
    name  = "recaptchaSiteKey"
    value = var.recaptcha_site_key
  }
  
  property {
    name  = "whiteList"
    value = var.trusted_sites
  }
}

# And create variables:
variable "recaptcha_secret" {
  type = string
}
variable "recaptcha_site_key" {
  type = string
}
variable "trusted_sites" {
  type = string
}
```

**Time Saved**: 3-6 hours for connector-heavy flows

---

### 4. ✅ **Comprehensive Module Documentation** (CRITICAL)

#### What dvtf-pingctl Generates

**README.md (4,328 lines)**:

**A. Module Usage Example**:
````markdown
## Usage

```hcl
module "davinci_flows" {
  source = "./path/to/my_generated_module"

  pingone_environment_id = var.pingone_environment_id

  # Connection Properties (commented with defaults)
  # davinci_connection_httpconnector__867ed_recaptchaSecretKey = null
  # davinci_connection_httpconnector__867ed_recaptchaSiteKey = null
  
  # Variable Overrides
  # davinci_variable_ciam_companyname_value = "Ping Identity"
  # davinci_variable_ciam_sessionlengthinminute_value = 5
  
  # Flow Configuration
  # davinci_flow_main_name = "Main Flow"
  # davinci_flow_main_log_level = 2
}
```
````

**B. Variable Reference Tables** (auto-generated from schema):

| Variable | Description | Type | Default |
|----------|-------------|------|---------|
| `davinci_connection_httpconnector__867ed_recaptchaSecretKey` | The reCAPTCHA v2 Secret Key for Http connector. The Secret Key from reCAPTCHA Admin dashboard. | string | null |
| `davinci_variable_ciam_companyname_value` | Company name displayed in flows | string | "Ping Identity" |
| `davinci_flow_main_log_level` | Log level (1=none, 2=info, 3=debug) | number | null |

**C. Resource Inventory**:
```markdown
This module creates:
- 47 flows
- 15 connector instances
- 32 variables
- 5 flow policies
```

**D. Validation Requirements**:
```markdown
## Required Variables

- `pingone_environment_id` (string) - Must be valid UUID format

## Optional Variables

All other variables have sensible defaults extracted from the flow export.
```

#### What davinci-terraform-converter Generates

**Current Output**: None

**Problems**:
- ❌ No usage documentation
- ❌ No variable reference
- ❌ No guidance on customization
- ❌ Consumer must read HCL to understand interface

#### Consumer Impact

**Without Documentation**:
- Read through 10,000+ lines of HCL to find variable names
- Guess variable types and formats
- Trial-and-error to understand module interface
- 2-4 hours to understand what can be customized

**With Documentation** (legacy):
- Scan README table for variable
- Copy-paste variable name
- See description and type in table
- 5-10 minutes to find and use variable

**Time Saved**: 2-4 hours per environment setup

---

### 5. ✅ **File Path Flexibility** (MEDIUM)

#### What dvtf-pingctl Generates

**Smart Path Resolution**:
```hcl
resource "davinci_flow" "main" {
  # Supports both absolute and relative paths
  flow_json = var.davinci_flow_main_json_file_path != null ? 
    (fileexists(var.davinci_flow_main_json_file_path) ? 
      file(var.davinci_flow_main_json_file_path) : 
      file(format("%s/%s", path.module, var.davinci_flow_main_json_file_path))
    ) : var.davinci_flow_main_json
}
```

**Benefits**:
- ✅ Tries absolute path first
- ✅ Falls back to module-relative path
- ✅ Supports inline JSON alternative
- ✅ Works in CI/CD with different working directories

**Usage**:
```hcl
# Absolute path
davinci_flow_main_json_file_path = "/opt/terraform/flows/main.json"

# Relative to module
davinci_flow_main_json_file_path = "assets/flows/main.json"

# Inline JSON
davinci_flow_main_json = file("${path.root}/custom-location/main.json")
```

#### What davinci-terraform-converter Generates

**Current Output**: N/A (new provider doesn't support external files)

**Note**: This is a provider limitation, not a CLI limitation. However, legacy CLI demonstrates best practices for file handling that could apply to other scenarios.

---

### 6. ✅ **Provider Version Constraints** (LOW but important)

#### What dvtf-pingctl Generates

```hcl
terraform {
  required_version = ">= 1.12"

  required_providers {
    davinci = {
      source  = "pingidentity/davinci"
      version = ">= 0.5.0, < 1.0.0"
    }
  }
}
```

**Benefits**:
- ✅ Ensures compatible Terraform version
- ✅ Pins provider version to tested range
- ✅ Prevents breaking changes from provider updates
- ✅ Self-documenting requirements

#### What davinci-terraform-converter Generates

**Current Output**: None

**Problems**:
- ❌ Consumer must determine compatible versions
- ❌ Risk of version conflicts
- ❌ No protection from breaking changes

#### Consumer Impact

**Manual Work**: Create `versions.tf` with appropriate constraints

**Time Saved**: 5-10 minutes

---

### 7. ✅ **Flow Dependency Management** (HIGH)

#### What dvtf-pingctl Generates

**Explicit Dependencies**:
```hcl
resource "davinci_flow" "main" {
  depends_on = [
    davinci_variable.ciam_companyname,
    davinci_variable.ciam_logourl,
    davinci_variable.ciam_sessionlengthinminute,
    # ... all variables used by this flow
  ]
  
  # Connection links maintain dependency order
  connection_link {
    id   = davinci_connection.httpconnector__867ed.id
    name = davinci_connection.httpconnector__867ed.name
    replace_import_connection_id = "867ed4363b2bc21c860085ad2baa817d"
  }
  
  # Subflow links ensure proper ordering
  subflow_link {
    id   = davinci_flow.subflow_auth.id
    name = davinci_flow.subflow_auth.name
    replace_import_subflow_id = "abc123..."
  }
}
```

**Benefits**:
- ✅ Ensures variables exist before flows
- ✅ Ensures connectors exist before flows
- ✅ Ensures subflows exist before parent flows
- ✅ Prevents race conditions in parallel apply
- ✅ Deterministic apply order

#### What davinci-terraform-converter Generates

**Current Output**: Inline references only (new provider)

**Note**: New provider uses inline references which create implicit dependencies. However, complex scenarios (flow variables, cross-resource references) may still benefit from explicit `depends_on`.

---

## Summary: Missing Features Impact

| Feature | Lines Generated | Consumer Time Saved | Priority |
|---------|----------------|---------------------|----------|
| **Variable Scaffolding** | 12,000+ | 4-8 hours | CRITICAL |
| **Output Values** | 3,600+ | 1-2 hours | HIGH |
| **Connector Property Blocks** | Variable | 3-6 hours | HIGH |
| **Module Documentation** | 4,300+ | 2-4 hours | CRITICAL |
| **File Path Flexibility** | N/A | N/A | MEDIUM |
| **Version Constraints** | 10-15 | 5-10 min | LOW |
| **Dependency Management** | Variable | 1-2 hours | HIGH |
| **TOTAL** | **~20,000+** | **11-23 hours** | |

---

## Recommendations for davinci-terraform-converter

### Phase 1: Critical Module Structure (Week 1-2)

1. **Generate `variables.tf`**:
   - Core `environment_id` variable with UUID validation
   - One variable per DaVinci variable: `davinci_variable_{name}_value`
   - Flow configuration variables: `davinci_flow_{name}_name`, `_description`, `_color`

2. **Generate `outputs.tf`**:
   - Flow ID outputs: `davinci_flow_{name}_id`
   - Flow object outputs: `davinci_flow_{name}`
   - Variable ID outputs if needed

3. **Generate `versions.tf`**:
   - Terraform version constraint
   - PingOne provider version constraint

### Phase 2: Enhanced Variables (Week 3)

4. **Connector Property Variables**:
   - Generate variable for each connector property
   - Include property descriptions from connector schema
   - Use `sensitive = true` for secrets

5. **Variable Validation**:
   - UUID format validation
   - Log level validation (1, 2, 3)
   - Custom validations per property type

### Phase 3: Documentation & Polish (Week 4)

6. **Generate `README.md`**:
   - Module usage example
   - Variable reference table
   - Resource inventory
   - Customization guide

7. **Comment Resources**:
   - Add `# Source flow: X` comments
   - Document connector instances with original names
   - Reference variable origins

### Implementation Example

```go
// New: ModuleStructure with variables
type ModuleStructure struct {
    Variables           []TerraformVariable
    Outputs             []TerraformOutput
    VersionConstraints  VersionConstraints
    Documentation       ModuleDocumentation
}

type TerraformVariable struct {
    Name         string
    Type         string
    Description  string
    Default      interface{}
    Sensitive    bool
    Validation   *VariableValidation
}

type VariableValidation struct {
    Condition    string
    ErrorMessage string
}

func GenerateModuleStructure(graph *DependencyGraph) *ModuleStructure {
    module := &ModuleStructure{}
    
    // Core environment ID
    module.Variables = append(module.Variables, TerraformVariable{
        Name:        "environment_id",
        Type:        "string",
        Description: "PingOne Environment ID",
        Validation: &VariableValidation{
            Condition:    `can(regex("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", var.environment_id))`,
            ErrorMessage: "Must be valid UUID format",
        },
    })
    
    // DaVinci variable overrides
    for _, variable := range graph.GetVariables() {
        module.Variables = append(module.Variables, TerraformVariable{
            Name:        fmt.Sprintf("davinci_variable_%s_value", sanitize(variable.Name)),
            Type:        variable.Type,
            Description: fmt.Sprintf("Value for DaVinci variable '%s'", variable.Name),
            Default:     variable.DefaultValue,
        })
    }
    
    // Flow outputs
    for _, flow := range graph.GetFlows() {
        module.Outputs = append(module.Outputs, TerraformOutput{
            Name:        fmt.Sprintf("davinci_flow_%s_id", sanitize(flow.Name)),
            Description: fmt.Sprintf("ID of %s flow", flow.Name),
            Value:       fmt.Sprintf("pingone_davinci_flow.%s.id", sanitize(flow.Name)),
        })
    }
    
    return module
}
```

---

## Conclusion

The legacy `dvtf-pingctl` demonstrates that a **comprehensive module structure dramatically reduces consumer effort**. By generating ~20,000 lines of boilerplate (variables, outputs, documentation), it transforms a raw HCL dump into a **production-ready, reusable Terraform module**.

**Target**: Add module structure generation to reduce consumer manual work from **11-23 hours → 1-2 hours** (80-90% reduction).

**Priority Order**:
1. Core variables (environment_id, variable values) - CRITICAL
2. Output values (flow IDs) - HIGH  
3. Module README - CRITICAL
4. Connector property variables - HIGH
5. Variable validation - MEDIUM
6. Version constraints - LOW
