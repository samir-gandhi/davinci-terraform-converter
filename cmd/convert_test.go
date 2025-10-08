// Copyright © 2025 Ping Identity Corporation

package cmd

import (
	"testing"
)

// TestDaVinciConvertCommand_Configuration tests that the command configuration
// is properly set up with all required metadata.
func TestDaVinciConvertCommand_Configuration(t *testing.T) {
	cmd := &DaVinciConvertCommand{}

	config, err := cmd.Configuration()
	if err != nil {
		t.Fatalf("Configuration() returned error: %v", err)
	}

	if config.Use != "convert" {
		t.Errorf("Expected Use to be 'convert', got '%s'", config.Use)
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
			flowJSON, out, err := parseArgs(tt.args)

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
