package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadModuleSchema(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a terraform file
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
	if err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write main.tf: %v", err)
	}

	schema, err := LoadModuleSchema(tmpDir, "")
	if err != nil {
		t.Fatalf("LoadModuleSchema failed: %v", err)
	}

	if schema.TerraformVersion != ">= 1.5.0" {
		t.Errorf("expected terraform version '>= 1.5.0', got '%s'", schema.TerraformVersion)
	}

	if len(schema.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(schema.Providers))
	}
	if schema.Providers[0].Name != "azurerm" {
		t.Errorf("expected provider 'azurerm', got '%s'", schema.Providers[0].Name)
	}

	if len(schema.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(schema.Variables))
	}

	if len(schema.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(schema.Outputs))
	}
}

func TestLoadModuleSchema_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "components", "test-module")

	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("failed to create module dir: %v", err)
	}

	tfContent := `variable "name" { type = string }`
	if err := os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write main.tf: %v", err)
	}

	schema, err := LoadModuleSchema(moduleDir, tmpDir)
	if err != nil {
		t.Fatalf("LoadModuleSchema failed: %v", err)
	}

	expected := filepath.Join("components", "test-module")
	if schema.Path != expected {
		t.Errorf("expected path '%s', got '%s'", expected, schema.Path)
	}
}

func TestLoadModuleSchema_InvalidPath(t *testing.T) {
	_, err := LoadModuleSchema("/nonexistent/path", "")
	if err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

func TestBuildVariableList_RequiredFirst(t *testing.T) {
	tmpDir := t.TempDir()

	// Create terraform with mixed required/optional variables
	tfContent := `
variable "optional_a" {
  type    = string
  default = "a"
}

variable "required_z" {
  type = string
}

variable "optional_b" {
  type    = string
  default = "b"
}

variable "required_a" {
  type = string
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write main.tf: %v", err)
	}

	schema, err := LoadModuleSchema(tmpDir, "")
	if err != nil {
		t.Fatalf("LoadModuleSchema failed: %v", err)
	}

	// Verify order: required first (alphabetically), then optional (alphabetically)
	expectedOrder := []struct {
		name     string
		required bool
	}{
		{"required_a", true},
		{"required_z", true},
		{"optional_a", false},
		{"optional_b", false},
	}

	if len(schema.Variables) != len(expectedOrder) {
		t.Fatalf("expected %d variables, got %d", len(expectedOrder), len(schema.Variables))
	}

	for i, expected := range expectedOrder {
		if schema.Variables[i].Name != expected.name {
			t.Errorf("variable[%d]: expected name '%s', got '%s'", i, expected.name, schema.Variables[i].Name)
		}
		if schema.Variables[i].Required != expected.required {
			t.Errorf("variable[%d] '%s': expected required=%v, got %v",
				i, expected.name, expected.required, schema.Variables[i].Required)
		}
	}
}

func TestVariableInfo_EmptyValueForType(t *testing.T) {
	tests := []struct {
		name     string
		tfType   string
		expected string
	}{
		{"string", "string", `""`},
		{"empty type defaults to string", "", `""`},
		{"number", "number", "0"},
		{"bool", "bool", "false"},
		{"list", "list(string)", "[]"},
		{"list any", "list(any)", "[]"},
		{"set", "set(string)", "[]"},
		{"map", "map(string)", "{}"},
		{"object", "object({name=string})", "{}"},
		{"unknown", "tuple", "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := VariableInfo{Type: tt.tfType}
			got := v.EmptyValueForType()
			if got != tt.expected {
				t.Errorf("EmptyValueForType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestVariableInfo_FullDefaultString(t *testing.T) {
	tests := []struct {
		name     string
		defVal   any
		expected string
	}{
		{"nil", nil, "null"},
		{"empty string", "", `""`},
		{"string", "hello", `"hello"`},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"number", float64(42), "42"},
		{"float", float64(3.14), "3.14"},
		{"list", []any{"a", "b"}, `["a","b"]`},
		{"map", map[string]any{"key": "value"}, `{"key":"value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := VariableInfo{Default: tt.defVal}
			got := v.FullDefaultString()
			if got != tt.expected {
				t.Errorf("FullDefaultString() = %q, want %q", got, tt.expected)
			}
		})
	}
}
