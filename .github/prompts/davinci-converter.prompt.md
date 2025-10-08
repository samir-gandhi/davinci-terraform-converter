---
mode: agent
---
Used to define the main command name throughout the prompt.

commandName: "pingcli davinci convert"
Default path for the input flow, can be changed when using the prompt.

flowJsonPath: "path/to/flow.json"
Default path for the output HCL, can be changed when using the prompt.
outputPath: "path/to/output.hcl"

# Project Brief: PingOne DaVinci Flow to HCL Converter CLI Plugin

## Your Role

You are an expert Go software engineer with extensive experience in building CLI tools, working with REST APIs, and developing Terraform providers. Your task is to build a CLI plugin for the pingcli tool. You will follow Test-Driven Development (TDD) principles to ensure the code is robust, maintainable, and ready for enterprise use. Break up the parts into manageable sections, and ensure each part is completed before moving to the next.

## Project Overview

The goal is to create a CLI command that ingests a PingOne DaVinci flow (in its native JSON format) and converts it into HCL (HashiCorp Configuration Language). This HCL will be compatible with the new DaVinci resources in the PingOne Terraform Provider. The tool must intelligently handle environment-specific values (like connection IDs, variables, etc.) by converting them into references to other Terraform-managed resources.
Key Technologies & Reference Materials

You will need to be familiar with the following:

    [Ping CLI Core](../../../../pingidentity/pingcli/cmd/) (pingcli): The target framework for this plugin. You will create a new command within this ecosystem. (https://github.com/pingidentity/pingcli)

    [Ping CLI Plugin System](../../../../pingidentity/pingcli/examples/plugin/): Understand how to create new commands for pingcli using the plugin architecture.

    [PingOne Go SDK](../../../../pingidentity/pingone-go-client/pingone): This is the primary library for interacting with the PingOne API and contains the necessary Go structs for unmarshaling DaVinci flow JSON. (https://github.com/pingidentity/pingone-go-client)

    [PingOne Terraform Provider](../../../../pingidentity/terraform-provider-pingone/internal/service/davinci/): The source code for the DaVinci resources should be used as a reference to understand how the Go SDK structs are mapped to HCL.

    [DaVinci OpenAPI Specification](./davinci-openapi.yaml): This is the source of truth for the structure of the DaVinci Flow JSON payload.
    
    [Sample DaVinci Flow JSON](./davinci-api-protect-reg-authn-flow.json): Use this to understand the real-world structure and data you will be working with.

    [Legacy DaVinci CLI Tool](../../../../patrickcping/dvtf-pingctl/): You will be looking to understand how the legacy tool handled similar conversions, especially regarding environment-specific values.

#{part1}
Part 1: Project Scaffolding and Command Structure

Your first task is to set up the project structure and create the basic CLI command.

    Initialize Project: Create a new Go module for the project.

    Project Layout: Structure the project using standard Go conventions (e.g., a cmd directory for the main command, internal for the core logic, and pkg if any code is intended to be shared).

    Create Cobra Command:

        Using the cobra library (which pingcli uses), scaffold a new command.

        The command should be callable like: ${commandName}.

        Define two flags for the command:

            --flow-json <path>: A required string flag pointing to the input DaVinci flow JSON file.

            --out <path>: An optional string flag for the output HCL file. If not provided, the output should go to stdout.

    Initial Implementation: The command's execution logic should, for now, simply print a message like "Executing DaVinci flow conversion for file: ${flowJsonPath}". This confirms the command structure is working.

    Write a Placeholder Test: Create a simple test file for the command to ensure the basic structure is testable.

#{part2}
Part 2: Core Conversion Logic (TDD)

Now, focus on the core logic of converting a simple flow JSON to HCL, without handling environment-specific dependencies yet.

    Create Test Case:

        In a new internal/converter package, create a converter_test.go file.

        Define a test function, TestSimpleFlowConversion.

        Inside the test, define a simple DaVinci flow as a multi-line JSON string. This flow should not contain any connections, variables, or subflows.

        Define the expected HCL output as a string.

        Write the test assertion first: call a (not-yet-written) Convert() function and assert that its output matches the expected HCL. This test will fail.

    Define Structs:

        Create the Go structs required to unmarshal the DaVinci flow JSON.

        Crucially, reference the structs in the pingone-go-client and the DaVinci OpenAPI spec. This ensures your data model is accurate. Do not reinvent these structs if they already exist.

    Implement Convert() Function:

        Create the Convert(flowJSON []byte) (string, error) function.

        Implement the logic:

            Unmarshal the input JSON into your Go structs.

            Use the data from the structs to construct the HCL for a pingone_davinci_flow resource. You can use Go's text/template package for this.

            Return the generated HCL string.

        Run the test until it passes.

#{part3}
Part 3: Handling Environment-Specific Dependencies (The Scalable Method)

This is the most critical part. We need a scalable way to identify and handle environment-specific values. We will start by converting them to HCL with placeholders.

    Refine Test Case:

        Create a new test, TestFlowWithDependenciesConversion.

        Use a sample flow JSON that does contain a hardcoded connectionID within one of its nodes.

        Define the expected HCL. The connection_id in the output HCL should not be the hardcoded ID. Instead, it should be a commented-out placeholder that includes the original ID for user reference.

        Example Expected HCL Snippet:

        # ... other HCL ...
        capability_name = "my-capability"
        connection_id   = "" # TODO: Replace with Terraform reference for original ID: 01a2b3c4-d5e6-f789-0123-456a7b8c9d0e
        # ... other HCL ...

    Implement a Resolver:

        Create a resolver mechanism within your converter package. This could be a struct or a set of functions.

        The resolver's job is to traverse the unmarshaled Go struct representation of the flow.

        It should contain a pre-defined list or map of JSON paths/field names that are known to be environment-specific (e.g., nodes.capability.connectionID).

        When it finds a match, it should replace the value with the placeholder string format shown above.

    Integrate Resolver: Modify your Convert() function to use this resolver after unmarshaling the JSON but before generating the HCL.

    Expand and Scale:

        Ensure the resolver is designed to be easily extensible. It should be simple to add new paths for other environment-specific items like variables and subFlows in the future.

        Add tests for flows containing variables and subflow references, ensuring they are also converted to appropriate TODO placeholders in the HCL.

#{part4}
Part 4: Final Integration and Error Handling

Finally, connect the logic to the CLI command and ensure it is robust.

    Connect Logic to Command: Update your Cobra command's execution logic from Part 1. It should now:

        Read the contents of the file specified by the --flow-json flag.

        Pass the file contents to the converter.Convert() function.

        Handle any errors returned from the converter (e.g., invalid JSON) and print a user-friendly error message to stderr.

    Handle Output:

        If the --out flag is provided, write the resulting HCL string to that file.

        If the --out flag is not provided, print the HCL string to stdout.

    Write Integration Tests: Create tests for the command itself. You can do this by executing the root command in your test code and capturing the output/error streams to verify:

        A valid flow JSON produces the expected HCL to stdout.

        Using the --out flag creates a file with the correct content.

        Providing a path to a non-existent JSON file returns an error.

        Providing a malformed JSON file returns an error.

    Plan for Deprecation:

        Ensure any logic that might be deprecated (like API client initialization, although not needed yet) is encapsulated in its own package or struct. Add a comment indicating that this component is a candidate for future replacement by a shared function from the main pingcli application.