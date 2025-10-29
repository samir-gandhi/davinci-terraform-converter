package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorCreateDirectories(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create generator
	config := ModuleConfig{
		OutputDir:      tmpDir,
		ModuleDirName:  "test-module",
		IncludeImports: false,
		IncludeValues:  false,
		EnvironmentID:  "test-env-id",
	}
	generator := NewGenerator(config)

	// Generate directories
	err := generator.createDirectories()
	require.NoError(t, err)

	// Verify child module directory exists
	childModulePath := filepath.Join(tmpDir, "test-module")
	info, err := os.Stat(childModulePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGeneratorVersionsTF(t *testing.T) {
	tmpDir := t.TempDir()

	config := ModuleConfig{
		OutputDir:     tmpDir,
		ModuleDirName: "test-module",
	}
	generator := NewGenerator(config)

	// Create directories first
	err := generator.createDirectories()
	require.NoError(t, err)

	// Generate versions.tf
	err = generator.generateVersionsTF()
	require.NoError(t, err)

	// Verify file was created
	versionsPath := filepath.Join(tmpDir, "test-module", "versions.tf")
	content, err := os.ReadFile(versionsPath)
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(content), "terraform {")
	assert.Contains(t, string(content), "required_version")
	assert.Contains(t, string(content), "pingone")
	assert.Contains(t, string(content), "pingidentity/pingone")
}

func TestGeneratorVariablesTF(t *testing.T) {
	tmpDir := t.TempDir()

	config := ModuleConfig{
		OutputDir:     tmpDir,
		ModuleDirName: "test-module",
	}
	generator := NewGenerator(config)

	// Create directories
	err := generator.createDirectories()
	require.NoError(t, err)

	// Generate variables.tf with some test variables
	variables := []Variable{
		{
			Name:         "test_flow_name",
			Type:         "string",
			Description:  "Name of the test flow",
			Default:      "Test Flow",
			ResourceType: "flow",
			ResourceName: "test_flow",
		},
		{
			Name:         "test_var_value",
			Type:         "string",
			Description:  "Value of test variable",
			Default:      "test value",
			Sensitive:    true,
			ResourceType: "variable",
			ResourceName: "test_var",
		},
	}

	err = generator.generateVariablesTF(variables)
	require.NoError(t, err)

	// Verify file was created
	variablesPath := filepath.Join(tmpDir, "test-module", "variables.tf")
	content, err := os.ReadFile(variablesPath)
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(content), "variable \"pingone_environment_id\"")
	assert.Contains(t, string(content), "variable \"test_flow_name\"")
	assert.Contains(t, string(content), "variable \"test_var_value\"")
	assert.Contains(t, string(content), "sensitive   = true")
	assert.Contains(t, string(content), "can(regex(")
}

func TestGeneratorOutputsTF(t *testing.T) {
	tmpDir := t.TempDir()

	config := ModuleConfig{
		OutputDir:     tmpDir,
		ModuleDirName: "test-module",
	}
	generator := NewGenerator(config)

	// Create directories
	err := generator.createDirectories()
	require.NoError(t, err)

	// Generate outputs.tf with test outputs
	outputs := []Output{
		{
			Name:        "flow_id",
			Description: "The ID of the flow",
			Value:       "pingone_davinci_flow.test.id",
		},
		{
			Name:        "secret_value",
			Description: "A secret value",
			Value:       "pingone_davinci_variable.secret.value",
			Sensitive:   true,
		},
	}

	err = generator.generateOutputsTF(outputs)
	require.NoError(t, err)

	// Verify file was created
	outputsPath := filepath.Join(tmpDir, "test-module", "outputs.tf")
	content, err := os.ReadFile(outputsPath)
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(content), "output \"flow_id\"")
	assert.Contains(t, string(content), "output \"secret_value\"")
	assert.Contains(t, string(content), "pingone_davinci_flow.test.id")
	assert.Contains(t, string(content), "sensitive   = true")
}

func TestGeneratorModuleTF(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		includeValues bool
		expectedEnv   string
	}{
		{
			name:          "Without values",
			includeValues: false,
			expectedEnv:   "pingone_environment_id = \"\"",
		},
		{
			name:          "With values",
			includeValues: true,
			expectedEnv:   "pingone_environment_id = \"test-env-123\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ModuleConfig{
				OutputDir:     tmpDir,
				ModuleDirName: "test-module",
				IncludeValues: tt.includeValues,
				EnvironmentID: "test-env-123",
			}
			generator := NewGenerator(config)

			// Create test structure
			structure := &ModuleStructure{
				Config: config,
				Variables: []Variable{
					{
						Name:         "test_var",
						Type:         "string",
						Description:  "Test variable",
						Default:      "default value",
						ResourceType: "flow",
					},
				},
			}

			// Generate module.tf
			err := generator.generateModuleTF(structure)
			require.NoError(t, err)

			// Verify file was created
			modulePath := filepath.Join(tmpDir, "module.tf")
			content, err := os.ReadFile(modulePath)
			require.NoError(t, err)

			// Verify content
			assert.Contains(t, string(content), "module \"davinci\"")
			assert.Contains(t, string(content), "source = \"./test-module\"")
			assert.Contains(t, string(content), tt.expectedEnv)
		})
	}
}

func TestGeneratorResourceFiles(t *testing.T) {
	tmpDir := t.TempDir()

	config := ModuleConfig{
		OutputDir:     tmpDir,
		ModuleDirName: "test-module",
	}
	generator := NewGenerator(config)

	// Create directories
	err := generator.createDirectories()
	require.NoError(t, err)

	// Generate resource files
	resources := ModuleResources{
		FlowsHCL:        "resource \"pingone_davinci_flow\" \"test\" {}",
		ConnectionsHCL:  "resource \"pingone_davinci_connector_instance\" \"http\" {}",
		VariablesHCL:    "resource \"pingone_davinci_variable\" \"company_name\" {}",
		ApplicationsHCL: "resource \"pingone_davinci_application\" \"app\" {}",
	}

	err = generator.generateResourceFiles(resources)
	require.NoError(t, err)

	// Verify files were created
	childModulePath := filepath.Join(tmpDir, "test-module")

	flowsPath := filepath.Join(childModulePath, "flows.tf")
	content, err := os.ReadFile(flowsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "pingone_davinci_flow")

	connectionsPath := filepath.Join(childModulePath, "connections.tf")
	content, err = os.ReadFile(connectionsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "pingone_davinci_connector_instance")

	variablesPath := filepath.Join(childModulePath, "variables_dv.tf")
	content, err = os.ReadFile(variablesPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "pingone_davinci_variable")
}

func TestGeneratorImportsTF(t *testing.T) {
	tmpDir := t.TempDir()

	config := ModuleConfig{
		OutputDir:      tmpDir,
		ModuleDirName:  "test-module",
		IncludeImports: true,
	}
	generator := NewGenerator(config)

	// Generate imports.tf
	importBlocks := []ImportBlock{
		{
			To: "module.davinci.pingone_davinci_flow.test",
			ID: "env-id:flow-id",
		},
		{
			To: "module.davinci.pingone_davinci_variable.company_name",
			ID: "env-id:var-id",
		},
	}

	err := generator.generateImportsTF(importBlocks)
	require.NoError(t, err)

	// Verify file was created
	importsPath := filepath.Join(tmpDir, "imports.tf")
	content, err := os.ReadFile(importsPath)
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(content), "import {")
	assert.Contains(t, string(content), "module.davinci.pingone_davinci_flow.test")
	assert.Contains(t, string(content), "env-id:flow-id")
	assert.Contains(t, string(content), "module.davinci.pingone_davinci_variable.company_name")
}

func TestFullModuleGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	config := ModuleConfig{
		OutputDir:      tmpDir,
		ModuleDirName:  "davinci-module",
		IncludeImports: true,
		IncludeValues:  true,
		EnvironmentID:  "test-env-123",
	}
	generator := NewGenerator(config)

	// Create full module structure
	structure := &ModuleStructure{
		Config: config,
		Variables: []Variable{
			{
				Name:         "davinci_flow_main_name",
				Type:         "string",
				Description:  "Name of main flow",
				Default:      "Main Flow",
				ResourceType: "flow",
			},
		},
		Outputs: []Output{
			{
				Name:        "main_flow_id",
				Description: "ID of main flow",
				Value:       "pingone_davinci_flow.main.id",
			},
		},
		Resources: ModuleResources{
			FlowsHCL:       "resource \"pingone_davinci_flow\" \"main\" {\n  environment_id = var.pingone_environment_id\n  name = var.davinci_flow_main_name\n}",
			ConnectionsHCL: "resource \"pingone_davinci_connector_instance\" \"http\" {}",
		},
		ImportBlocks: []ImportBlock{
			{
				To: "module.davinci.pingone_davinci_flow.main",
				ID: "test-env-123:flow-123",
			},
		},
	}

	// Generate
	err := generator.Generate(structure)
	require.NoError(t, err)

	// Verify all files exist
	childModulePath := filepath.Join(tmpDir, "davinci-module")

	// Check child module files
	assert.FileExists(t, filepath.Join(childModulePath, "versions.tf"))
	assert.FileExists(t, filepath.Join(childModulePath, "variables.tf"))
	assert.FileExists(t, filepath.Join(childModulePath, "outputs.tf"))
	assert.FileExists(t, filepath.Join(childModulePath, "flows.tf"))
	assert.FileExists(t, filepath.Join(childModulePath, "connections.tf"))

	// Check root module files
	assert.FileExists(t, filepath.Join(tmpDir, "module.tf"))
	assert.FileExists(t, filepath.Join(tmpDir, "imports.tf"))
}
