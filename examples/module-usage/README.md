# Module Usage Examples

This directory contains shell scripts demonstrating various module generation workflows.

## Quick Start

```bash
# Set required environment variables
export PINGONE_WORKER_ENVIRONMENT_ID="your-worker-env-id"
export PINGONE_EXPORT_ENVIRONMENT_ID="your-target-env-id"
export PINGONE_WORKER_CLIENT_ID="your-client-id"
export PINGONE_WORKER_CLIENT_SECRET="your-client-secret"
export PINGONE_REGION_CODE="NA"  # or EU, AP, CA

# Run any example
chmod +x examples/module-usage/*.sh
./examples/module-usage/01-shareable-module.sh
```

## Examples

### 01-shareable-module.sh

Generate a reusable Terraform module without hardcoded values.

**Use Case:**
- Share DaVinci configurations across teams
- Publish to Terraform Registry
- Version control module for multiple environments

**Output Structure:**
```
shareable-module/
├── davinci-module/
│   ├── versions.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── flows.tf
│   ├── connections.tf
│   └── variables_dv.tf
└── module.tf
```

**Key Features:**
- Variables with empty defaults
- No environment-specific values
- Requires user to provide values when using module

### 02-environment-deploy.sh

Generate environment-specific configuration with actual values from API.

**Use Case:**
- Deploy to dev/staging/prod environments
- Migrate existing DaVinci environment to Terraform
- Quick environment replication

**Command:**
```bash
./examples/module-usage/02-environment-deploy.sh dev
./examples/module-usage/02-environment-deploy.sh prod
```

**Key Features:**
- `--include-values`: Populates variables with actual values
- `--include-imports`: Generates import blocks for existing resources
- Secrets marked as TODO for manual entry
- Ready for immediate terraform apply

### 03-custom-module-name.sh

Generate module with custom directory name for organization.

**Use Case:**
- Organize multiple modules by purpose
- Descriptive naming for different flow categories
- Multi-module repository structure

**Command:**
```bash
./examples/module-usage/03-custom-module-name.sh my-auth-flows
./examples/module-usage/03-custom-module-name.sh partner-integrations
```

**Key Features:**
- `--module-dir`: Custom child module directory name
- Useful for mono-repo module organization
- Clear module purpose from directory name

### 04-legacy-single-file.sh

Generate single HCL file without module structure (legacy mode).

**Use Case:**
- Backwards compatibility with existing workflows
- Quick exports for analysis
- Integration with tools expecting single file
- Custom module structure requirements

**Key Features:**
- `--module=false`: Disables module generation
- Single output file with all resources
- Compatible with pre-module workflows
- Simple for small configurations

## Comparison Matrix

| Feature | Shareable Module | Environment Deploy | Custom Name | Legacy Single File |
|---------|-----------------|-------------------|-------------|-------------------|
| Module Structure | ✅ | ✅ | ✅ | ❌ |
| Actual Values | ❌ | ✅ | ❌ | N/A |
| Import Blocks | ✅ | ✅ | ✅ | ✅ (with flag) |
| Reusable | ✅ | ❌ | ✅ | ❌ |
| Ready to Deploy | ❌ | ✅ | ❌ | ✅ (with vars) |
| Custom Name | ❌ | ❌ | ✅ | N/A |

## Common Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--module` | `true` | Generate module structure |
| `--module-dir` | `davinci-module` | Child module directory name |
| `--include-values` | `false` | Populate variables with actual values |
| `--include-imports` | `true` | Generate import blocks |
| `--skip-imports` | `false` | Skip import block generation |
| `--out` | Current dir | Output directory |

## Best Practices

### For Shareable Modules
1. Keep secrets undefined (let consumers provide)
2. Document all variables with descriptions
3. Use validation blocks for variable constraints
4. Version control the module
5. Tag releases for semantic versioning

### For Environment Deployments
1. Review secret TODO comments before apply
2. Use separate directories for each environment
3. Store sensitive values in secret management systems
4. Test in dev/staging before prod
5. Use Terraform workspaces or separate state files

### For Multi-Module Repos
1. Use descriptive module-dir names
2. Organize by domain or feature (auth, registration, mfa)
3. Maintain separate README for each module
4. Document inter-module dependencies
5. Use consistent naming conventions

## Troubleshooting

### Module Not Found
```bash
# Ensure you're in the correct directory
cd shareable-module
terraform init
```

### Variables Not Populated
```bash
# Use --include-values flag
davinci-convert export --include-values --out ./env-config
```

### Import Blocks Not Working
```bash
# Requires Terraform 1.5+
terraform version

# Or disable if using older version
davinci-convert export --skip-imports --out ./config
```

### Custom Module Name Not Applied
```bash
# Check --module-dir flag
davinci-convert export --module-dir my-module --out ./output
# Verify module.tf references "./my-module"
```
