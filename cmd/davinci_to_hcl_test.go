// Copyright © 2025 Ping Identity Corporation

package cmd

import (
	"os"
	"strings"
	"testing"
)

// MockLogger is a test implementation of grpc.Logger
type MockLogger struct{}

func (m *MockLogger) Message(msg string, data map[string]string) error {
	return nil
}

func (m *MockLogger) Success(msg string, data map[string]string) error {
	return nil
}

func (m *MockLogger) PluginError(msg string, data map[string]string) error {
	return nil
}

func (m *MockLogger) UserError(msg string, data map[string]string) error {
	return nil
}

func (m *MockLogger) UserFatal(msg string, data map[string]string) error {
	return nil
}

func (m *MockLogger) Warn(msg string, data map[string]string) error {
	return nil
}

// TestDaVinciToHclCommand_Configuration tests that the command configuration
// is properly set up with all required metadata.
func TestDaVinciToHclCommand_Configuration(t *testing.T) {
	cmd := &DaVinciToHclCommand{}

	config, err := cmd.Configuration()
	if err != nil {
		t.Fatalf("Configuration() returned error: %v", err)
	}

	if config.Use != DaVinciToHclUse {
		t.Errorf("Expected Use to be %q, got '%s'", DaVinciToHclUse, config.Use)
	}

	if config.Short == "" {
		t.Error("Expected Short description to be non-empty")
	}

	if config.Long == "" {
		t.Error("Expected Long description to be non-empty")
	}

	if config.Example == "" {
		t.Error("Expected Example to be non-empty")
	}
}

// TestParseArgs tests the flag parsing logic
func TestParseArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantFlowJSON string
		wantOut      string
		wantErr      bool
	}{
		{
			name:         "Required flow-json flag provided",
			args:         []string{"--flow-json", "test.json"},
			wantFlowJSON: "test.json",
			wantOut:      "",
			wantErr:      false,
		},
		{
			name:         "Both flags provided",
			args:         []string{"--flow-json", "test.json", "--out", "output.tf"},
			wantFlowJSON: "test.json",
			wantOut:      "output.tf",
			wantErr:      false,
		},
		{
			name:         "Missing required flag",
			args:         []string{"--out", "output.tf"},
			wantFlowJSON: "",
			wantOut:      "",
			wantErr:      true,
		},
		{
			name:         "No flags provided",
			args:         []string{},
			wantFlowJSON: "",
			wantOut:      "",
			wantErr:      true,
		},
		{
			name:         "Flow JSON with equals syntax",
			args:         []string{"--flow-json=test.json"},
			wantFlowJSON: "test.json",
			wantOut:      "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flowJSON, out, _, err := parseArgs(tt.args) // Ignore skipDependencies in these tests

			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if flowJSON != tt.wantFlowJSON {
					t.Errorf("parseArgs() flowJSON = %v, want %v", flowJSON, tt.wantFlowJSON)
				}

				if out != tt.wantOut {
					t.Errorf("parseArgs() out = %v, want %v", out, tt.wantOut)
				}
			}
		})
	}
}

// TestParseArgsWithSkipDependencies tests parsing of the skip-dependencies flag
func TestParseArgsWithSkipDependencies(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantSkipDeps bool
		wantErr      bool
	}{
		{
			name:         "Skip dependencies flag set to true",
			args:         []string{"--flow-json=test.json", "--skip-dependencies"},
			wantSkipDeps: true,
			wantErr:      false,
		},
		{
			name:         "Skip dependencies flag not set",
			args:         []string{"--flow-json=test.json"},
			wantSkipDeps: false,
			wantErr:      false,
		},
		{
			name:         "Skip dependencies with explicit false",
			args:         []string{"--flow-json=test.json", "--skip-dependencies=false"},
			wantSkipDeps: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, skipDeps, err := parseArgs(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && skipDeps != tt.wantSkipDeps {
				t.Errorf("parseArgs() skipDependencies = %v, want %v", skipDeps, tt.wantSkipDeps)
			}
		})
	}
}

// TestHasFlag tests the helper function for checking flag presence
func TestHasFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		want     bool
	}{
		{
			name:     "Flag exists with double dash",
			args:     []string{"--flow-json", "test.json"},
			flagName: "flow-json",
			want:     true,
		},
		{
			name:     "Flag does not exist",
			args:     []string{"--out", "test.tf"},
			flagName: "flow-json",
			want:     false,
		},
		{
			name:     "Empty args",
			args:     []string{},
			flagName: "flow-json",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasFlag(tt.args, tt.flagName); got != tt.want {
				t.Errorf("hasFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseFlowFiles tests parsing multiple flow file inputs
func TestParseFlowFiles(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFiles []string
		wantErr   bool
	}{
		{
			name:      "Single file",
			args:      []string{"--flow-json", "test.json"},
			wantFiles: []string{"test.json"},
			wantErr:   false,
		},
		{
			name:      "Multiple files - repeated flag",
			args:      []string{"--flow-json", "test1.json", "--flow-json", "test2.json"},
			wantFiles: []string{"test1.json", "test2.json"},
			wantErr:   false,
		},
		{
			name:      "Multiple files - comma separated",
			args:      []string{"--flow-json", "test1.json,test2.json,test3.json"},
			wantFiles: []string{"test1.json", "test2.json", "test3.json"},
			wantErr:   false,
		},
		{
			name:      "Mixed: repeated and comma-separated",
			args:      []string{"--flow-json", "test1.json,test2.json", "--flow-json", "test3.json"},
			wantFiles: []string{"test1.json", "test2.json", "test3.json"},
			wantErr:   false,
		},
		{
			name:      "Empty file list",
			args:      []string{"--flow-json", ""},
			wantFiles: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := parseFlowFiles(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlowFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(files) != len(tt.wantFiles) {
					t.Errorf("parseFlowFiles() got %d files, want %d", len(files), len(tt.wantFiles))
					return
				}

				for i, file := range files {
					if file != tt.wantFiles[i] {
						t.Errorf("parseFlowFiles() file[%d] = %v, want %v", i, file, tt.wantFiles[i])
					}
				}
			}
		})
	}
}

// TestParseDirectoryInput tests parsing directory input
func TestParseDirectoryInput(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		wantDir     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid directory path",
			args:    []string{"--flow-json-dir", tempDir},
			wantDir: tempDir,
			wantErr: false,
		},
		{
			name:    "No directory specified",
			args:    []string{},
			wantDir: "",
			wantErr: false,
		},
		{
			name:        "Both file and directory specified",
			args:        []string{"--flow-json", "test.json", "--flow-json-dir", tempDir},
			wantDir:     "",
			wantErr:     true,
			errContains: "cannot specify both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := parseDirectoryInput(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseDirectoryInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !stringContains(err.Error(), tt.errContains) {
					t.Errorf("parseDirectoryInput() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr && dir != tt.wantDir {
				t.Errorf("parseDirectoryInput() = %v, want %v", dir, tt.wantDir)
			}
		})
	}
}

// TestDiscoverFlowFilesInDirectory tests finding JSON files in a directory
func TestDiscoverFlowFilesInDirectory(t *testing.T) {
	// Create temp directory structure for testing
	tempDir := t.TempDir()

	// Create test files
	createTestFile(t, tempDir, "flow1.json", `{"name":"flow1"}`)
	createTestFile(t, tempDir, "flow2.json", `{"name":"flow2"}`)
	createTestFile(t, tempDir, "not-json.txt", "text file")
	createTestFile(t, tempDir, "readme.md", "# Readme")

	tests := []struct {
		name          string
		dir           string
		wantFileCount int
		wantErr       bool
	}{
		{
			name:          "Directory with multiple JSON files",
			dir:           tempDir,
			wantFileCount: 2, // Only flow1.json and flow2.json
			wantErr:       false,
		},
		{
			name:          "Empty directory",
			dir:           t.TempDir(),
			wantFileCount: 0,
			wantErr:       false,
		},
		{
			name:          "Non-existent directory",
			dir:           "/non/existent/path",
			wantFileCount: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := discoverFlowFilesInDirectory(tt.dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("discoverFlowFilesInDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(files) != tt.wantFileCount {
				t.Errorf("discoverFlowFilesInDirectory() got %d files, want %d", len(files), tt.wantFileCount)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Helper function to create test files
func createTestFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := dir + "/" + filename
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
}

// TestRunConvert_MultipleFiles tests the full runConvert function with multiple files
func TestRunConvert_MultipleFiles(t *testing.T) {
	// Create test data directory
	testDir := t.TempDir()

	// Reusable simple flow JSON for testing
	simpleFlowJSON := `{
  "name": "Test Flow",
  "flowId": "test-flow-id",
  "graphData": {
    "elements": {
      "nodes": [],
      "edges": []
    }
  }
}`

	// Create test flow files
	file1 := testDir + "/flow1.json"
	file2 := testDir + "/flow2.json"

	if err := os.WriteFile(file1, []byte(simpleFlowJSON), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Modify flow name for second file
	flow2JSON := strings.Replace(simpleFlowJSON, "Test Flow", "Test Flow 2", 1)
	flow2JSON = strings.Replace(flow2JSON, "test-flow-id", "test-flow-id-2", 1)
	if err := os.WriteFile(file2, []byte(flow2JSON), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create mock logger
	logger := &MockLogger{}

	// Test with output to file
	t.Run("Multiple files to output file", func(t *testing.T) {
		outFile := testDir + "/output.tf"
		cmd := &DaVinciToHclCommand{}

		err := cmd.runConvert(logger, []string{file1, file2}, outFile, false)
		if err != nil {
			t.Errorf("runConvert() error = %v", err)
			return
		}

		// Verify output file was created
		content, err := os.ReadFile(outFile)
		if err != nil {
			t.Errorf("Failed to read output file: %v", err)
			return
		}

		output := string(content)

		// Verify both flows are in the output
		if !strings.Contains(output, "Test Flow") {
			t.Errorf("Output missing first flow")
		}
		if !strings.Contains(output, "Test Flow 2") {
			t.Errorf("Output missing second flow")
		}

		// Verify resources are properly separated
		if strings.Count(output, "resource \"pingone_davinci_flow\"") != 2 {
			t.Errorf("Expected 2 flow resources, got different count")
		}
	})

	// Test with single file (backward compatibility)
	t.Run("Single file backward compatibility", func(t *testing.T) {
		outFile := testDir + "/single_output.tf"
		cmd := &DaVinciToHclCommand{}

		err := cmd.runConvert(logger, []string{file1}, outFile, false)
		if err != nil {
			t.Errorf("runConvert() error = %v", err)
			return
		}

		// Verify output file was created
		content, err := os.ReadFile(outFile)
		if err != nil {
			t.Errorf("Failed to read output file: %v", err)
			return
		}

		output := string(content)

		// Verify single flow is in the output
		if !strings.Contains(output, "Test Flow") {
			t.Errorf("Output missing flow")
		}
		if strings.Count(output, "resource \"pingone_davinci_flow\"") != 1 {
			t.Errorf("Expected 1 flow resource")
		}
	})
}

// TestRunConvert_DirectoryInput tests conversion from a directory of flow files
func TestRunConvert_DirectoryInput(t *testing.T) {
	// Create test data directory
	testDir := t.TempDir()

	// Reusable simple flow JSON template
	flowTemplate := `{
  "name": "Flow %d",
  "flowId": "flow-id-%d",
  "graphData": {
    "elements": {
      "nodes": [],
      "edges": []
    }
  }
}`

	// Create multiple test flow files
	var flowFiles []string
	for i := 1; i <= 3; i++ {
		filename := testDir + "/flow" + string(rune('0'+i)) + ".json"
		content := strings.Replace(flowTemplate, "%d", string(rune('0'+i)), -1)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		flowFiles = append(flowFiles, filename)
	}

	// Create mock logger
	logger := &MockLogger{}
	cmd := &DaVinciToHclCommand{}

	outFile := testDir + "/dir_output.tf"

	err := cmd.runConvert(logger, flowFiles, outFile, false)
	if err != nil {
		t.Errorf("runConvert() error = %v", err)
		return
	}

	// Verify output file was created
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	output := string(content)

	// Verify all flows are in the output
	resourceCount := strings.Count(output, "resource \"pingone_davinci_flow\"")
	if resourceCount != 3 {
		t.Errorf("Expected 3 flow resources, got %d", resourceCount)
	}
}

// TestRunConvert_ErrorHandling tests error scenarios in runConvert
func TestRunConvert_ErrorHandling(t *testing.T) {
	logger := &MockLogger{}
	cmd := &DaVinciToHclCommand{}

	tests := []struct {
		name      string
		flowFiles []string
		wantErr   bool
	}{
		{
			name:      "Empty file list",
			flowFiles: []string{},
			wantErr:   true,
		},
		{
			name:      "Non-existent file",
			flowFiles: []string{"/non/existent/file.json"},
			wantErr:   true,
		},
		{
			name:      "Invalid JSON",
			flowFiles: []string{},
			wantErr:   false, // Will be set up with invalid JSON file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special setup for invalid JSON test
			if tt.name == "Invalid JSON" {
				tempDir := t.TempDir()
				invalidFile := tempDir + "/invalid.json"
				if err := os.WriteFile(invalidFile, []byte("not valid json"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				tt.flowFiles = []string{invalidFile}
				tt.wantErr = true
			}

			err := cmd.runConvert(logger, tt.flowFiles, "", false)

			if (err != nil) != tt.wantErr {
				t.Errorf("runConvert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
