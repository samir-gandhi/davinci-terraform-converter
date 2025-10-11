# Skip Dependencies Feature

## Overview

The `--skip-dependencies` flag allows you to generate HCL with hardcoded connection IDs instead of Terraform resource references.

## Use Cases

### Manual Testing
When you don't have connector instance resources managed in Terraform yet, use `--skip-dependencies` to get hardcoded IDs that can be manually updated in your Terraform configuration.

### Gradual Migration
Convert flows first, then add connector instances later. With hardcoded IDs, you can apply flows immediately without having all dependencies managed as Terraform resources.

## Usage

### With Dependencies (Default)
Generates Terraform references:
```bash
./davinci-convert convert --flow-json my-flow.json
```

Output:
```hcl
connection_id = pingone_davinci_connector_instance.httpconnector_conn-abc-123.id
```

### Without Dependencies
Generates hardcoded IDs:
```bash
./davinci-convert convert --flow-json my-flow.json --skip-dependencies
```

Output:
```hcl
connection_id = "conn-abc-123"
```

## Implementation Details

- Flag added to CLI: `--skip-dependencies` (boolean, defaults to false)
- New converter functions:
  - `ConvertWithOptions(flowJSON []byte, skipDependencies bool)`
  - `ConvertMultiFlowWithOptions(flowJSON []byte, skipDependencies bool)`
- Original functions (`Convert`, `ConvertMultiFlow`) now call the new functions with `skipDependencies=false`

## Future: Part 4 Integration

When Part 4 (Dependency Resolution) is implemented:
- The flag will work with selective exports
- Missing dependencies will generate TODO placeholders
- Naming consistency between flow references and exported connector instances will be validated

See `.github/prompts/06-part4-dependencies.md` for details on the naming contract that must be maintained.
