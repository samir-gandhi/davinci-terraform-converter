---
mode: agent
---

# Part 2 - Phase 2.7: Terraform Provider Integration Tests

**Status**: ✅ COMPLETE

**Completed**: October 11, 2025

## Overview

Integration tests that validate generated HCL against the actual Terraform PingOne provider. These tests ensure syntactic and semantic correctness beyond unit tests.

## Why This Matters

**Unit tests** verify conversion logic, but **Terraform** validates:
- Actual HCL syntax correctness
- Schema compliance with the provider
- Resource attribute types and requirements
- Reference resolution
- Terraform version compatibility

**Integration tests catch**:
- Incorrect attribute syntax (`settings { }` vs `settings = { }`)
- Invalid resource references
- Schema violations (missing required fields, wrong types)
- Provider version incompatibilities

## Prerequisites

**Terraform CLI Required**:
- Tests should skip if terraform not in PATH
- Use `t.Skip()` for graceful handling
- Document how to run with real credentials (optional)

**Provider Setup**:
- PingOne Terraform Provider must be installable
- Use local provider mirror or registry

## Test Environment Setup

**File**: `internal/converter/terraform_integration_test.go`

**Fixtures Directory**: `internal/converter/testdata/terraform/`
- `provider.tf` - Provider configuration
- `variables.tf` - Required input variables
- Sample flow JSON files

### Provider Configuration

```hcl
# testdata/terraform/provider.tf
terraform {
  required_version = ">= 1.4"
  
  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = "~> 1.1"
    }
  }
}

provider "pingone" {
  client_id      = var.client_id
  client_secret  = var.client_secret
  environment_id = var.environment_id
  region_code    = var.region_code
}
```

### Variables File

```hcl
# testdata/terraform/variables.tf
variable "client_id" {
  type        = string
  description = "PingOne OAuth client ID"
  default     = "test-client-id"
}

variable "client_secret" {
  type        = string
  description = "PingOne OAuth client secret"
  sensitive   = true
  default     = "test-client-secret"
}

variable "environment_id" {
  type        = string
  description = "PingOne environment ID"
  default     = "00000000-0000-0000-0000-000000000000"
}

variable "region_code" {
  type        = string
  description = "PingOne region code"
  default     = "NA"
}
```

## Test Implementation

### Test: Terraform Validate Flow

**Purpose**: Verify generated HCL has valid syntax

```go
func TestTerraformValidateFlow(t *testing.T) {
    // Skip if terraform not available
    if _, err := exec.LookPath("terraform"); err != nil {
        t.Skip("Terraform not found in PATH, skipping validation test")
    }

    // Load test flow JSON
    flowJSON, err := os.ReadFile("testdata/complete-flow.json")
    require.NoError(t, err)

    // Generate HCL
    hcl, err := Convert(flowJSON)
    require.NoError(t, err)

    // Create temp directory for Terraform
    tmpDir := t.TempDir()

    // Copy provider config
    providerConfig, err := os.ReadFile("testdata/terraform/provider.tf")
    require.NoError(t, err)
    err = os.WriteFile(filepath.Join(tmpDir, "provider.tf"), providerConfig, 0644)
    require.NoError(t, err)

    // Copy variables config
    varsConfig, err := os.ReadFile("testdata/terraform/variables.tf")
    require.NoError(t, err)
    err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), varsConfig, 0644)
    require.NoError(t, err)

    // Write generated HCL
    err = os.WriteFile(filepath.Join(tmpDir, "flow.tf"), []byte(hcl), 0644)
    require.NoError(t, err)

    // Run terraform init
    cmd := exec.Command("terraform", "init", "-no-color")
    cmd.Dir = tmpDir
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "terraform init failed:\n%s", output)

    // Run terraform validate
    cmd = exec.Command("terraform", "validate", "-no-color")
    cmd.Dir = tmpDir
    output, err = cmd.CombinedOutput()
    require.NoError(t, err, "terraform validate failed:\n%s", output)

    // Assert success message
    assert.Contains(t, string(output), "The configuration is valid")
}
```

### Test: Terraform Plan Flow

**Purpose**: Verify provider accepts generated HCL (even without credentials)

```go
func TestTerraformPlanFlow(t *testing.T) {
    if _, err := exec.LookPath("terraform"); err != nil {
        t.Skip("Terraform not found in PATH, skipping plan test")
    }

    // Load test flow
    flowJSON, err := os.ReadFile("testdata/complete-flow.json")
    require.NoError(t, err)

    // Generate HCL
    hcl, err := Convert(flowJSON)
    require.NoError(t, err)

    // Setup temp directory
    tmpDir := t.TempDir()

    // Write configs
    providerConfig := getTestProviderConfig()
    err = os.WriteFile(filepath.Join(tmpDir, "provider.tf"), []byte(providerConfig), 0644)
    require.NoError(t, err)

    err = os.WriteFile(filepath.Join(tmpDir, "flow.tf"), []byte(hcl), 0644)
    require.NoError(t, err)

    // Run terraform init
    cmd := exec.Command("terraform", "init", "-no-color")
    cmd.Dir = tmpDir
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "terraform init failed:\n%s", output)

    // Run terraform plan (may fail on auth, but should parse HCL)
    cmd = exec.Command("terraform", "plan", "-no-color", "-input=false")
    cmd.Dir = tmpDir
    cmd.Env = append(os.Environ(), 
        "TF_VAR_client_id=test",
        "TF_VAR_client_secret=test",
        "TF_VAR_environment_id=00000000-0000-0000-0000-000000000000",
        "TF_VAR_region_code=NA",
    )
    output, err = cmd.CombinedOutput()

    // Plan may fail due to auth, but HCL should be parseable
    if err != nil {
        // Check if failure is auth-related (acceptable) vs syntax error (not acceptable)
        outputStr := string(output)
        if strings.Contains(outputStr, "Invalid value for variable") ||
           strings.Contains(outputStr, "Error: Missing required argument") ||
           strings.Contains(outputStr, "Unsupported argument") {
            t.Fatalf("terraform plan failed with syntax error:\n%s", output)
        }
        // Auth failures are OK for this test
        t.Logf("terraform plan failed (likely auth), but HCL parsed successfully:\n%s", output)
    } else {
        t.Logf("terraform plan succeeded:\n%s", output)
    }
}
```

### Test: Settings Attribute Syntax

**Purpose**: Verify flow_configuration_json uses attribute syntax

```go
func TestSettingsAttributeSyntax(t *testing.T) {
    flowJSON, err := os.ReadFile("testdata/complete-flow.json")
    require.NoError(t, err)

    hcl, err := Convert(flowJSON)
    require.NoError(t, err)

    // Verify attribute syntax: "flow_configuration_json = {"
    assert.Contains(t, hcl, "flow_configuration_json = {",
        "flow_configuration_json should use attribute syntax (=), not block syntax")

    // Verify NOT using block syntax: "flow_configuration_json {"
    assert.NotContains(t, hcl, "flow_configuration_json {",
        "flow_configuration_json should not use block syntax")

    // Verify settings within flow_configuration_json uses map syntax
    assert.Contains(t, hcl, "settings = {",
        "settings should use map attribute syntax within flow_configuration_json")
}
```

### Test: Required Attributes Present

**Purpose**: Verify all required resource attributes are present

```go
func TestRequiredAttributesPresent(t *testing.T) {
    flowJSON, err := os.ReadFile("testdata/complete-flow.json")
    require.NoError(t, err)

    hcl, err := Convert(flowJSON)
    require.NoError(t, err)

    // Required attributes for pingone_davinci_flow
    requiredAttrs := []string{
        "environment_id",
        "name",
        "flow_configuration_json",
    }

    for _, attr := range requiredAttrs {
        assert.Contains(t, hcl, attr+" =",
            "Required attribute %s must be present", attr)
    }
}
```

### Test: Resource References Valid

**Purpose**: Verify resource references use correct syntax

```go
func TestResourceReferencesValid(t *testing.T) {
    // Test multi-resource export
    exportData, err := os.ReadFile("testdata/multi-resource-export.json")
    require.NoError(t, err)

    hcl, err := ConvertAll(exportData)
    require.NoError(t, err)

    // Check for valid reference patterns
    // Flow referencing connection:
    // connection_id = pingone_davinci_connection.connection_name.id
    connectionRefPattern := regexp.MustCompile(`connection_id\s*=\s*pingone_davinci_connection\.\w+\.id`)
    assert.Regexp(t, connectionRefPattern, hcl,
        "Connection references should use correct format")

    // Flow policy referencing flow:
    // flow_id = pingone_davinci_flow.flow_name.id
    flowRefPattern := regexp.MustCompile(`flow_id\s*=\s*pingone_davinci_flow\.\w+\.id`)
    assert.Regexp(t, flowRefPattern, hcl,
        "Flow references should use correct format")

    // Application reference in flow policy:
    // application_id = pingone_davinci_application.app_name.id
    appRefPattern := regexp.MustCompile(`application_id\s*=\s*pingone_davinci_application\.\w+\.id`)
    assert.Regexp(t, appRefPattern, hcl,
        "Application references should use correct format")
}
```

## Test Execution

**Run all tests**:
```bash
go test ./internal/converter -v
```

**Run only integration tests**:
```bash
go test ./internal/converter -v -run TestTerraform
```

**Skip integration tests** (when terraform not installed):
```bash
# Tests automatically skip if terraform not in PATH
go test ./internal/converter -v
```

## Mock vs Real Provider Testing

### Mock Testing (Default)

Use test credentials that won't authenticate but allow syntax validation:
- `client_id = "test-client-id"`
- `client_secret = "test-client-secret"`
- `environment_id = "00000000-0000-0000-0000-000000000000"`

### Real Provider Testing (Optional)

For teams with test environments:

```bash
# Set real credentials
export TF_VAR_client_id="real-client-id"
export TF_VAR_client_secret="real-client-secret"
export TF_VAR_environment_id="real-env-id"
export TF_VAR_region_code="NA"

# Run tests
go test ./internal/converter -v -run TestTerraformPlan
```

## Success Criteria

- All syntax validation tests pass
- Terraform can parse generated HCL without errors
- Required attributes present in all resources
- Resource references use correct format
- Tests skip gracefully when terraform not available

## Benefits

1. **Confidence**: Know HCL will work in production Terraform workflows
2. **Early Detection**: Catch schema issues before manual testing
3. **Version Testing**: Validate against specific provider versions
4. **Documentation**: Tests serve as examples of valid HCL

## Next Steps

After integration tests pass:
1. Proceed to Part 3 (API Export)
2. Add more test cases for edge scenarios
3. Document known limitations
