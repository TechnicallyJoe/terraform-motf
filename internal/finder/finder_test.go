package finder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindModule_SingleMatch(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create components/azurerm/storage-account with .tf file
	modulePath := filepath.Join(tmpDir, "azurerm", "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	matches, err := FindModule(tmpDir, "storage-account")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	if matches[0] != modulePath {
		t.Errorf("expected match to be '%s', got '%s'", modulePath, matches[0])
	}
}

func TestFindModule_MultipleMatches(t *testing.T) {
	// Create temp directory structure with name clash
	tmpDir := t.TempDir()

	// Create two modules with the same name under different providers
	module1 := filepath.Join(tmpDir, "azurerm", "storage-account")
	module2 := filepath.Join(tmpDir, "aws", "storage-account")

	for _, path := range []string{module1, module2} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create module directory: %v", err)
		}
		tfFile := filepath.Join(path, "main.tf")
		if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
			t.Fatalf("failed to create .tf file: %v", err)
		}
	}

	matches, err := FindModule(tmpDir, "storage-account")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestFindModule_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a module with a different name
	modulePath := filepath.Join(tmpDir, "other-module")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	matches, err := FindModule(tmpDir, "storage-account")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestFindModule_IgnoresDirectoriesWithoutTfFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with matching name but no .tf files
	modulePath := filepath.Join(tmpDir, "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	// Create a non-.tf file
	otherFile := filepath.Join(modulePath, "README.md")
	if err := os.WriteFile(otherFile, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	matches, err := FindModule(tmpDir, "storage-account")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("expected 0 matches (no .tf files), got %d", len(matches))
	}
}

func TestFindModule_MatchesTfJsonFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create module with .tf.json file
	modulePath := filepath.Join(tmpDir, "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfJsonFile := filepath.Join(modulePath, "main.tf.json")
	if err := os.WriteFile(tfJsonFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create .tf.json file: %v", err)
	}

	matches, err := FindModule(tmpDir, "storage-account")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match for .tf.json, got %d", len(matches))
	}
}

func TestFindModule_NestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create deeply nested module
	modulePath := filepath.Join(tmpDir, "level1", "level2", "level3", "my-module")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	matches, err := FindModule(tmpDir, "my-module")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	if matches[0] != modulePath {
		t.Errorf("expected match to be '%s', got '%s'", modulePath, matches[0])
	}
}

func TestFindModule_SearchPathDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist")

	_, err := FindModule(nonExistentPath, "any-module")
	if err == nil {
		t.Error("expected error for non-existent search path, got nil")
	}
}

func TestHasTerraformFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{
			name:     "has .tf file",
			files:    []string{"main.tf"},
			expected: true,
		},
		{
			name:     "has .tf.json file",
			files:    []string{"main.tf.json"},
			expected: true,
		},
		{
			name:     "has multiple .tf files",
			files:    []string{"main.tf", "variables.tf", "outputs.tf"},
			expected: true,
		},
		{
			name:     "has mixed files",
			files:    []string{"main.tf", "README.md"},
			expected: true,
		},
		{
			name:     "no .tf files",
			files:    []string{"README.md", "config.yaml"},
			expected: false,
		},
		{
			name:     "empty directory",
			files:    []string{},
			expected: false,
		},
		{
			name:     "only .json file (not .tf.json)",
			files:    []string{"config.json"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.files {
				filePath := filepath.Join(tmpDir, file)
				if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			result := HasTerraformFiles(tmpDir)
			if result != tt.expected {
				t.Errorf("HasTerraformFiles() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasTerraformFiles_NonExistentDir(t *testing.T) {
	result := HasTerraformFiles("/non/existent/path")
	if result != false {
		t.Error("expected false for non-existent directory")
	}
}

func TestListAllModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple modules
	modules := []string{
		filepath.Join(tmpDir, "azurerm", "storage-account"),
		filepath.Join(tmpDir, "azurerm", "key-vault"),
		filepath.Join(tmpDir, "aws", "s3-bucket"),
	}

	for _, path := range modules {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create module directory: %v", err)
		}
		tfFile := filepath.Join(path, "main.tf")
		if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
			t.Fatalf("failed to create .tf file: %v", err)
		}
	}

	result, err := ListAllModules(tmpDir)
	if err != nil {
		t.Fatalf("ListAllModules returned error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 modules, got %d", len(result))
	}

	// Check that all expected modules are present
	expectedNames := []string{"storage-account", "key-vault", "s3-bucket"}
	for _, name := range expectedNames {
		if _, exists := result[name]; !exists {
			t.Errorf("expected module '%s' not found", name)
		}
	}
}

func TestMatchesWildcard(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		// Exact matches
		{"storage-account", "storage-account", true},
		{"storage-account", "storage", false},

		// Wildcards at the end
		{"storage-account", "storage-*", true},
		{"storage-account", "key-*", false},

		// Wildcards at the start
		{"storage-account", "*account", true},
		{"storage-account", "*vault", false},

		// Wildcards in the middle
		{"storage-account", "storage*account", true},
		{"storage-account", "key*account", false},

		// Multiple wildcards
		{"my-storage-account", "*storage*", true},
		{"my-storage-account", "*key*", false},

		// Only wildcard
		{"storage-account", "*", true},
		{"anything", "*", true},

		// Multiple wildcards at various positions
		{"prod-storage-account-east", "prod*storage*east", true},
		{"prod-storage-account-west", "prod*storage*east", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_with_"+tt.pattern, func(t *testing.T) {
			result := MatchesWildcard(tt.name, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesWildcard(%q, %q) = %v, expected %v",
					tt.name, tt.pattern, result, tt.expected)
			}
		})
	}
}

// Tests for skipping .terraform and other excluded directories

func TestFindModule_SkipsTerraformDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid module
	validModule := filepath.Join(tmpDir, "storage-account")
	if err := os.MkdirAll(validModule, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(validModule, "main.tf"), []byte("# valid module"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create a .terraform directory with a cached module that has the same name
	cachedModule := filepath.Join(validModule, ".terraform", "modules", "storage-account")
	if err := os.MkdirAll(cachedModule, 0755); err != nil {
		t.Fatalf("failed to create cached module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cachedModule, "main.tf"), []byte("# cached module"), 0644); err != nil {
		t.Fatalf("failed to create cached main.tf: %v", err)
	}

	matches, err := FindModule(tmpDir, "storage-account")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	// Should only find the valid module, not the cached one
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %v", len(matches), matches)
	}

	if matches[0] != validModule {
		t.Errorf("expected match to be '%s', got '%s'", validModule, matches[0])
	}
}

func TestListAllModules_SkipsTerraformDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid module
	validModule := filepath.Join(tmpDir, "my-module")
	if err := os.MkdirAll(validModule, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(validModule, "main.tf"), []byte("# valid module"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create a .terraform directory with cached modules
	cachedModule := filepath.Join(validModule, ".terraform", "modules", "registry-module")
	if err := os.MkdirAll(cachedModule, 0755); err != nil {
		t.Fatalf("failed to create cached module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cachedModule, "main.tf"), []byte("# cached module"), 0644); err != nil {
		t.Fatalf("failed to create cached main.tf: %v", err)
	}

	modules, err := ListAllModules(tmpDir)
	if err != nil {
		t.Fatalf("ListAllModules returned error: %v", err)
	}

	// Should only find my-module, not registry-module
	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d: %v", len(modules), modules)
	}

	if _, exists := modules["my-module"]; !exists {
		t.Error("expected to find 'my-module'")
	}

	if _, exists := modules["registry-module"]; exists {
		t.Error("should not find 'registry-module' from .terraform directory")
	}
}

func TestIsSkipDir(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"examples", true},
		{"tests", true},
		{"modules", true},
		{".terraform", true},
		{".git", true},
		{"node_modules", true},
		{".spacelift", true},
		{"components", false},
		{"azurerm", false},
		{"src", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSkipDir(tt.name); got != tt.want {
				t.Errorf("IsSkipDir(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPathContainsSkipDir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"components/azurerm/storage-account/examples/basic", true},
		{"components/azurerm/storage-account/tests", true},
		{"components/azurerm/storage-account/modules/internal", true},
		{"components/azurerm/storage-account", false},
		{"projects/prod-infra", false},
		{"examples/basic", true},
		{"tests", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := PathContainsSkipDir(tt.path); got != tt.want {
				t.Errorf("PathContainsSkipDir(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestFindModule_SkipsGitDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid module
	validModule := filepath.Join(tmpDir, "my-module")
	if err := os.MkdirAll(validModule, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(validModule, "main.tf"), []byte("# valid module"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create a .git directory (should be skipped)
	gitDir := filepath.Join(tmpDir, ".git", "my-module")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "main.tf"), []byte("# git content"), 0644); err != nil {
		t.Fatalf("failed to create git main.tf: %v", err)
	}

	matches, err := FindModule(tmpDir, "my-module")
	if err != nil {
		t.Fatalf("FindModule returned error: %v", err)
	}

	// Should only find the valid module
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %v", len(matches), matches)
	}

	if matches[0] != validModule {
		t.Errorf("expected match to be '%s', got '%s'", validModule, matches[0])
	}
}
