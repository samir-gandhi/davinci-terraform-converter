---
mode: agent
---

# DaVinci Converter - Project Context

## Variables

commandName: "pingcli davinci convert"
flowJsonPath: "path/to/flow.json"
outputPath: "path/to/output.hcl"

## Your Role

You are an expert Go software engineer with extensive experience in building CLI tools, working with REST APIs, and developing Terraform providers. Your task is to build a CLI plugin for the pingcli tool. You will follow Test-Driven Development (TDD) principles to ensure the code is robust, maintainable, and ready for enterprise use.

## Project Overview

The goal is to create a comprehensive CLI command that exports a complete PingOne DaVinci environment configuration and converts it into HCL (HashiCorp Configuration Language). This HCL will be compatible with the DaVinci resources in the PingOne Terraform Provider.

The tool will:

1. Export all DaVinci resources from an environment via API calls (flows, applications, connector instances, variables, flow policies, etc.)
2. Convert the exported JSON to valid HCL with proper Terraform resource syntax
3. Intelligently link dependencies between resources using Terraform references instead of hardcoded IDs
4. Validate the generated HCL against the Terraform provider schema

## Key Technologies & Reference Materials

**Ping CLI Core** (../../../../pingidentity/pingcli/cmd/): The target framework for this plugin. You will create a new command within this ecosystem. (https://github.com/pingidentity/pingcli)

**Ping CLI Plugin System** (../../../../pingidentity/pingcli/examples/plugin/): Understand how to create new commands for pingcli using the plugin architecture.

**PingOne Go SDK** (../../../../pingidentity/pingone-go-client/pingone): This is the primary library for interacting with the PingOne API and contains the necessary Go structs for unmarshaling DaVinci flow JSON. (https://github.com/pingidentity/pingone-go-client)

**PingOne Terraform Provider** (../../../../pingidentity/terraform-provider-pingone/internal/service/davinci/): The source code for the DaVinci resources should be used as a reference to understand how the Go SDK structs are mapped to HCL.

**DaVinci OpenAPI Specification** (./davinci-openapi.yaml): This is the source of truth for the structure of the DaVinci Flow JSON payload.

**Sample DaVinci Flow JSON** (./davinci-api-protect-reg-authn-flow.json): Use this to understand the real-world structure and data you will be working with.

**Legacy DaVinci CLI Tool** (../../../../patrickcping/dvtf-pingctl/): You will be looking to understand how the legacy tool handled similar conversions, especially regarding environment-specific values.

## Current Project State

**Completed:**
- ✅ Part 1: Project Scaffolding and Command Structure
  - Working CLI command with flags (--flow-json, --out)
  - Basic test infrastructure
  - Project structure (cmd/, internal/converter/)

**In Progress:**
- 🔧 Part 2: Complete DaVinci Resource Conversion
  - Phase 2.1: Comprehensive Flow Structure Conversion (CURRENT)
    - 669-line test suite (flow_comprehensive_test.go)
    - 549-line converter (flow_converter.go)
    - 24/27 tests passing (88.9% pass rate)
    - Successfully processes real production flows (31KB, 359KB files)

**Not Started:**
- ⏳ Phase 2.2-2.6: Other resource types (applications, flow policies, connections, variables)
- ⏳ Part 3: Full DaVinci Environment Export via API
- ⏳ Part 4: Dependency Resolution and Terraform References
- ⏳ Part 5: Final Integration and Error Handling
- ⏳ Part 6: Production Readiness and Release

## File Structure

```
davinci-terraform-converter/
├── cmd/
│   ├── convert.go           # Cobra command implementation
│   └── convert_test.go      # Command tests
├── internal/
│   └── converter/
│       ├── converter.go              # Main converter logic (21 tests passing)
│       ├── converter_test.go         # Original converter tests
│       ├── flow_converter.go         # Flow-specific conversion (549 lines)
│       ├── flow_comprehensive_test.go # Comprehensive tests (669 lines, 27 tests)
│       └── real_file_test.go         # Real production file tests
├── main.go
├── go.mod
└── README.md
```

## Development Approach

Break work into manageable phases:
1. Complete current phase before moving to next
2. Write tests first (TDD)
3. Run tests frequently
4. Fix bugs systematically
5. Validate with real production data
