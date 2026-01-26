package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/terraform"
)

func TestDescribeCmd_Output(t *testing.T) {
	// Create a temp module with terraform files
	tmpDir := t.TempDir()

	// Create .git to establish root
	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	// Create components/test-module structure
	moduleDir := filepath.Join(tmpDir, "components", "test-module")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("failed to create module dir: %v", err)
	}

	// Create a simple terraform file
	tfContent := `
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0.0"
    }
  }
}

variable "name" {
  type        = string
  description = "The name of the resource"
}

variable "location" {
  type        = string
  default     = "eastus"
  description = "The Azure region"
}

output "id" {
  value       = "test-id"
  description = "The resource ID"
}
`
	if err := os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write main.tf: %v", err)
	}

	// Reset flags
	describeJsonFlag = false
	pathFlag = moduleDir

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"describe"})

	err := rootCmd.Execute()
	pathFlag = "" // Reset

	if err != nil {
		t.Fatalf("describe command failed: %v", err)
	}

	output := buf.String()

	// Check expected content
	if !strings.Contains(output, "Module:") {
		t.Error("expected output to contain 'Module:'")
	}
	if !strings.Contains(output, "Terraform:") {
		t.Error("expected output to contain 'Terraform:'")
	}
	if !strings.Contains(output, "Providers:") {
		t.Error("expected output to contain 'Providers:'")
	}
	if !strings.Contains(output, "azurerm") {
		t.Error("expected output to contain 'azurerm' provider")
	}
	if !strings.Contains(output, "Variables:") {
		t.Error("expected output to contain 'Variables:'")
	}
	if !strings.Contains(output, "name") {
		t.Error("expected output to contain 'name' variable")
	}
	if !strings.Contains(output, "Outputs:") {
		t.Error("expected output to contain 'Outputs:'")
	}
}

func TestDescribeCmd_JSONOutput(t *testing.T) {
	// Create a temp module with terraform files
	tmpDir := t.TempDir()

	// Create .git to establish root
	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	// Create components/test-module structure
	moduleDir := filepath.Join(tmpDir, "components", "test-module")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("failed to create module dir: %v", err)
	}

	// Create a simple terraform file
	tfContent := `
variable "name" {
  type        = string
  description = "The name"
}

output "id" {
  value = "test"
}
`
	if err := os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write main.tf: %v", err)
	}

	// Set flags for JSON output
	describeJsonFlag = true
	pathFlag = moduleDir

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"describe", "--json"})

	err := rootCmd.Execute()
	describeJsonFlag = false
	pathFlag = ""

	if err != nil {
		t.Fatalf("describe command failed: %v", err)
	}

	// Parse JSON output
	var schema terraform.ModuleSchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}

	if schema.Name != "test-module" {
		t.Errorf("expected name 'test-module', got '%s'", schema.Name)
	}
	if len(schema.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(schema.Variables))
	}
	if len(schema.Outputs) != 1 {
		t.Errorf("expected 1 output, got %d", len(schema.Outputs))
	}
}

func TestDescribeCmd_ModuleNotFound(t *testing.T) {
	describeJsonFlag = false
	pathFlag = "/nonexistent/module/path"

	rootCmd.SetArgs([]string{"describe"})

	err := rootCmd.Execute()
	pathFlag = ""

	if err == nil {
		t.Error("expected error for non-existent module, got nil")
	}
}

func TestFormatDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		required bool
		want     string
	}{
		{"required", nil, true, "(required)"},
		{"nil", nil, false, "null"},
		{"empty string", "", false, `""`},
		{"string", "hello", false, `"hello"`},
		{"bool true", true, false, "true"},
		{"bool false", false, false, "false"},
		{"number", float64(42), false, "42"},
		{"float", float64(3.14), false, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDefault(tt.value, tt.required)
			if got != tt.want {
				t.Errorf("formatDefault(%v, %v) = %q, want %q", tt.value, tt.required, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is..."},
		{"ab", 5, "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     []string
	}{
		{"empty", "", 60, nil},
		{"short", "hello world", 60, []string{"hello world"}},
		{"wrap", "this is a longer text that should wrap", 20, []string{"this is a longer", "text that should", "wrap"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.maxWidth)
			if len(got) != len(tt.want) {
				t.Errorf("wrapText(%q, %d) = %v, want %v", tt.text, tt.maxWidth, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapText(%q, %d)[%d] = %q, want %q", tt.text, tt.maxWidth, i, got[i], tt.want[i])
				}
			}
		})
	}
}
