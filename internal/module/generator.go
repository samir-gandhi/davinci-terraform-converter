package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generator handles the generation of Terraform module structure
type Generator struct {
	config ModuleConfig
}

// NewGenerator creates a new module generator with the given configuration
func NewGenerator(config ModuleConfig) *Generator {
	return &Generator{
		config: config,
	}
}

// Generate creates the complete module structure
func (g *Generator) Generate(structure *ModuleStructure) error {
	// Create directory structure
	if err := g.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate child module files
	if err := g.generateVersionsTF(); err != nil {
		return fmt.Errorf("failed to generate versions.tf: %w", err)
	}

	if err := g.generateVariablesTF(structure.Variables); err != nil {
		return fmt.Errorf("failed to generate variables.tf: %w", err)
	}

	if err := g.generateOutputsTF(structure.Outputs); err != nil {
		return fmt.Errorf("failed to generate outputs.tf: %w", err)
	}

	if err := g.generateResourceFiles(structure.Resources); err != nil {
		return fmt.Errorf("failed to generate resource files: %w", err)
	}

	// Generate root module files
	if err := g.generateModuleTF(structure); err != nil {
		return fmt.Errorf("failed to generate module.tf: %w", err)
	}

	if g.config.IncludeImports {
		if err := g.generateImportsTF(structure.ImportBlocks); err != nil {
			return fmt.Errorf("failed to generate imports.tf: %w", err)
		}
	}

	return nil
}

// createDirectories creates the necessary directory structure
func (g *Generator) createDirectories() error {
	childModulePath := filepath.Join(g.config.OutputDir, g.config.ModuleDirName)
	return os.MkdirAll(childModulePath, 0755)
}

// childModulePath returns the full path to the child module directory
func (g *Generator) childModulePath() string {
	return filepath.Join(g.config.OutputDir, g.config.ModuleDirName)
}

// writeFile writes content to a file in the specified directory
func (g *Generator) writeFile(dir, filename, content string) error {
	filePath := filepath.Join(dir, filename)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// generateVersionsTF creates the versions.tf file in the child module
func (g *Generator) generateVersionsTF() error {
	content := `terraform {
  required_version = ">= 1.3"

  required_providers {
    pingone = {
      source  = "pingidentity/pingone"
      version = ">= 1.0.0"
    }
  }
}
`
	return g.writeFile(g.childModulePath(), "versions.tf", content)
}

// generateVariablesTF creates the variables.tf file in the child module
func (g *Generator) generateVariablesTF(variables []Variable) error {
	var sb strings.Builder

	// Always include the core environment_id variable
	sb.WriteString(`variable "pingone_environment_id" {
  type        = string
  description = "The PingOne environment ID to configure DaVinci resources in"

  validation {
    condition     = can(regex("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", var.pingone_environment_id))
    error_message = "The PingOne Environment ID must be a valid PingOne resource ID (UUID format)."
  }
}

`)

	// Group variables by resource type for better organization
	groupedVars := g.groupVariablesByResourceType(variables)

	// Generate variables in a logical order
	order := []string{"flow", "variable", "connection", "application", "flow_policy"}
	for _, resourceType := range order {
		vars, exists := groupedVars[resourceType]
		if !exists {
			continue
		}

		sb.WriteString(fmt.Sprintf("# %s Variables\n\n", strings.Title(resourceType)))

		for _, v := range vars {
			sb.WriteString(g.generateVariableBlock(v))
			sb.WriteString("\n")
		}
	}

	return g.writeFile(g.childModulePath(), "variables.tf", sb.String())
}

// groupVariablesByResourceType groups variables by their resource type
func (g *Generator) groupVariablesByResourceType(variables []Variable) map[string][]Variable {
	grouped := make(map[string][]Variable)
	for _, v := range variables {
		grouped[v.ResourceType] = append(grouped[v.ResourceType], v)
	}
	return grouped
}

// generateVariableBlock generates a single variable block
func (g *Generator) generateVariableBlock(v Variable) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("variable \"%s\" {\n", v.Name))
	sb.WriteString(fmt.Sprintf("  type        = %s\n", v.Type))
	sb.WriteString(fmt.Sprintf("  description = %q\n", v.Description))

	if v.Default != nil {
		sb.WriteString(fmt.Sprintf("  default     = %s\n", g.formatDefaultValue(v.Default, v.Type)))
	}

	if v.Sensitive {
		sb.WriteString("  sensitive   = true\n")
	}

	if v.Validation != nil {
		sb.WriteString("\n  validation {\n")
		sb.WriteString(fmt.Sprintf("    condition     = %s\n", v.Validation.Condition))
		sb.WriteString(fmt.Sprintf("    error_message = %q\n", v.Validation.ErrorMessage))
		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

// formatDefaultValue formats a default value based on its type
func (g *Generator) formatDefaultValue(value interface{}, varType string) string {
	if value == nil {
		return "null"
	}

	switch varType {
	case "string":
		return fmt.Sprintf("%q", value)
	case "number":
		return fmt.Sprintf("%v", value)
	case "bool":
		return fmt.Sprintf("%v", value)
	default:
		return fmt.Sprintf("%q", value)
	}
}

// generateOutputsTF creates the outputs.tf file in the child module
func (g *Generator) generateOutputsTF(outputs []Output) error {
	var sb strings.Builder

	for _, o := range outputs {
		sb.WriteString(fmt.Sprintf("output \"%s\" {\n", o.Name))
		sb.WriteString(fmt.Sprintf("  description = %q\n", o.Description))
		sb.WriteString(fmt.Sprintf("  value       = %s\n", o.Value))

		if o.Sensitive {
			sb.WriteString("  sensitive   = true\n")
		}

		sb.WriteString("}\n\n")
	}

	return g.writeFile(g.childModulePath(), "outputs.tf", sb.String())
}

// generateResourceFiles creates the resource files in the child module
func (g *Generator) generateResourceFiles(resources ModuleResources) error {
	// Generate flows.tf
	if resources.FlowsHCL != "" {
		if err := g.writeFile(g.childModulePath(), "flows.tf", resources.FlowsHCL); err != nil {
			return err
		}
	}

	// Generate connections.tf
	if resources.ConnectionsHCL != "" {
		if err := g.writeFile(g.childModulePath(), "connections.tf", resources.ConnectionsHCL); err != nil {
			return err
		}
	}

	// Generate variables_dv.tf (DaVinci variables)
	if resources.VariablesHCL != "" {
		if err := g.writeFile(g.childModulePath(), "variables_dv.tf", resources.VariablesHCL); err != nil {
			return err
		}
	}

	// Generate applications.tf
	if resources.ApplicationsHCL != "" {
		if err := g.writeFile(g.childModulePath(), "applications.tf", resources.ApplicationsHCL); err != nil {
			return err
		}
	}

	// Generate flow_policies.tf
	if resources.FlowPoliciesHCL != "" {
		if err := g.writeFile(g.childModulePath(), "flow_policies.tf", resources.FlowPoliciesHCL); err != nil {
			return err
		}
	}

	return nil
}

// generateModuleTF creates the module.tf file in the root module
func (g *Generator) generateModuleTF(structure *ModuleStructure) error {
	var sb strings.Builder

	sb.WriteString("module \"davinci\" {\n")
	sb.WriteString(fmt.Sprintf("  source = \"./%s\"\n\n", g.config.ModuleDirName))

	// Core environment ID
	if g.config.IncludeValues {
		sb.WriteString(fmt.Sprintf("  pingone_environment_id = %q\n\n", g.config.EnvironmentID))
	} else {
		sb.WriteString("  pingone_environment_id = \"\"  # TODO: Provide PingOne environment ID\n\n")
	}

	// Group variables by resource type
	groupedVars := g.groupVariablesByResourceType(structure.Variables)

	// Generate variable inputs
	order := []string{"flow", "variable", "connection", "application", "flow_policy"}
	for _, resourceType := range order {
		vars, exists := groupedVars[resourceType]
		if !exists {
			continue
		}

		sb.WriteString(fmt.Sprintf("  # %s Variables\n", strings.Title(resourceType)))

		for _, v := range vars {
			sb.WriteString(g.generateModuleInput(v))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}\n")

	return g.writeFile(g.config.OutputDir, "module.tf", sb.String())
}

// generateModuleInput generates a single module input line
func (g *Generator) generateModuleInput(v Variable) string {
	if !g.config.IncludeValues {
		// Empty value with comment
		if v.IsSecret {
			return fmt.Sprintf("  %s = \"\"  # TODO: Provide secret value\n", v.Name)
		}
		switch v.Type {
		case "string":
			return fmt.Sprintf("  %s = \"\"\n", v.Name)
		case "number":
			return fmt.Sprintf("  %s = 0\n", v.Name)
		case "bool":
			return fmt.Sprintf("  %s = false\n", v.Name)
		default:
			return fmt.Sprintf("  %s = null\n", v.Name)
		}
	}

	// Include actual value
	if v.IsSecret {
		return fmt.Sprintf("  %s = \"\"  # TODO: Provide secret value\n", v.Name)
	}

	if v.Default != nil {
		return fmt.Sprintf("  %s = %s\n", v.Name, g.formatDefaultValue(v.Default, v.Type))
	}

	return fmt.Sprintf("  %s = null\n", v.Name)
}

// generateImportsTF creates the imports.tf file in the root module
func (g *Generator) generateImportsTF(importBlocks []ImportBlock) error {
	var sb strings.Builder

	for _, ib := range importBlocks {
		sb.WriteString("import {\n")
		sb.WriteString(fmt.Sprintf("  to = %s\n", ib.To))
		sb.WriteString(fmt.Sprintf("  id = %q\n", ib.ID))
		sb.WriteString("}\n\n")
	}

	return g.writeFile(g.config.OutputDir, "imports.tf", sb.String())
}
