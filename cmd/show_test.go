package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/tfpl/internal/config"
)

func TestDirHasContent_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if dirHasContent(emptyDir) {
		t.Error("expected false for empty directory")
	}
}

func TestDirHasContent_WithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	dirWithFiles := filepath.Join(tmpDir, "with-files")
	if err := os.MkdirAll(dirWithFiles, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dirWithFiles, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !dirHasContent(dirWithFiles) {
		t.Error("expected true for directory with files")
	}
}

func TestDirHasContent_NonExistent(t *testing.T) {
	if dirHasContent("/non/existent/path") {
		t.Error("expected false for non-existent directory")
	}
}

func TestDirHasContent_WithSubdirs(t *testing.T) {
	tmpDir := t.TempDir()

	dirWithSubdirs := filepath.Join(tmpDir, "with-subdirs")
	subdir := filepath.Join(dirWithSubdirs, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	if !dirHasContent(dirWithSubdirs) {
		t.Error("expected true for directory with subdirectories")
	}
}

func TestListItems_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	items := listItems(filepath.Join(tmpDir, "examples"), tmpDir)

	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestListItems_WithMainTf(t *testing.T) {
	tmpDir := t.TempDir()

	examplesDir := filepath.Join(tmpDir, "examples")

	// Create example with main.tf (should be included)
	basicDir := filepath.Join(examplesDir, "basic")
	if err := os.MkdirAll(basicDir, 0755); err != nil {
		t.Fatalf("failed to create basic example: %v", err)
	}
	if err := os.WriteFile(filepath.Join(basicDir, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create example with main.tf (should be included)
	advancedDir := filepath.Join(examplesDir, "advanced")
	if err := os.MkdirAll(advancedDir, 0755); err != nil {
		t.Fatalf("failed to create advanced example: %v", err)
	}
	if err := os.WriteFile(filepath.Join(advancedDir, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	items := listItems(examplesDir, tmpDir)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Check that both are found (order may vary)
	names := make(map[string]bool)
	for _, item := range items {
		names[item.Name] = true
	}

	if !names["basic"] {
		t.Error("expected to find 'basic'")
	}
	if !names["advanced"] {
		t.Error("expected to find 'advanced'")
	}
}

func TestListItems_IgnoresDirsWithoutMainTf(t *testing.T) {
	tmpDir := t.TempDir()

	examplesDir := filepath.Join(tmpDir, "examples")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatalf("failed to create examples dir: %v", err)
	}

	// Create a file (should be ignored)
	if err := os.WriteFile(filepath.Join(examplesDir, "README.md"), []byte("# Examples"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create a directory WITHOUT main.tf (should be ignored)
	emptyDir := filepath.Join(examplesDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("failed to create empty dir: %v", err)
	}

	// Create a directory WITH main.tf (should be included)
	basicDir := filepath.Join(examplesDir, "basic")
	if err := os.MkdirAll(basicDir, 0755); err != nil {
		t.Fatalf("failed to create basic dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(basicDir, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	items := listItems(examplesDir, tmpDir)

	if len(items) != 1 {
		t.Fatalf("expected 1 item (only dir with main.tf), got %d", len(items))
	}

	if items[0].Name != "basic" {
		t.Errorf("expected item name 'basic', got '%s'", items[0].Name)
	}
}

func TestListTestFiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	tests := listTestFiles(filepath.Join(tmpDir, "tests"), tmpDir)

	if len(tests) != 0 {
		t.Errorf("expected 0 tests, got %d", len(tests))
	}
}

func TestListTestFiles_WithTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("failed to create tests dir: %v", err)
	}

	// Create test files
	if err := os.WriteFile(filepath.Join(testsDir, "basic_test.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create basic_test.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "advanced_test.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create advanced_test.go: %v", err)
	}

	tests := listTestFiles(testsDir, tmpDir)

	if len(tests) != 2 {
		t.Fatalf("expected 2 tests, got %d", len(tests))
	}

	names := make(map[string]bool)
	for _, test := range tests {
		names[test.Name] = true
	}

	if !names["basic_test.go"] {
		t.Error("expected to find 'basic_test.go'")
	}
	if !names["advanced_test.go"] {
		t.Error("expected to find 'advanced_test.go'")
	}
}

func TestListTestFiles_IgnoresNonTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("failed to create tests dir: %v", err)
	}

	// Create non-test files (should be ignored)
	if err := os.WriteFile(filepath.Join(testsDir, "helpers.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create helpers.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "go.mod"), []byte("module tests"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create test file (should be included)
	if err := os.WriteFile(filepath.Join(testsDir, "basic_test.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create basic_test.go: %v", err)
	}

	tests := listTestFiles(testsDir, tmpDir)

	if len(tests) != 1 {
		t.Fatalf("expected 1 test (ignoring non-test files), got %d", len(tests))
	}

	if tests[0].Name != "basic_test.go" {
		t.Errorf("expected test name 'basic_test.go', got '%s'", tests[0].Name)
	}
}

func TestListTestFiles_IgnoresDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("failed to create tests dir: %v", err)
	}

	// Create a directory named like a test (should be ignored)
	if err := os.MkdirAll(filepath.Join(testsDir, "something_test.go"), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create actual test file
	if err := os.WriteFile(filepath.Join(testsDir, "real_test.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create real_test.go: %v", err)
	}

	tests := listTestFiles(testsDir, tmpDir)

	if len(tests) != 1 {
		t.Fatalf("expected 1 test (ignoring directories), got %d", len(tests))
	}

	if tests[0].Name != "real_test.go" {
		t.Errorf("expected test name 'real_test.go', got '%s'", tests[0].Name)
	}
}

func TestFormatBool(t *testing.T) {
	if formatBool(true) != "Yes" {
		t.Errorf("expected 'Yes' for true, got '%s'", formatBool(true))
	}
	if formatBool(false) != "No" {
		t.Errorf("expected 'No' for false, got '%s'", formatBool(false))
	}
}

func TestFormatType(t *testing.T) {
	if formatType("") != "unknown" {
		t.Errorf("expected 'unknown' for empty string, got '%s'", formatType(""))
	}
	if formatType("component") != "component" {
		t.Errorf("expected 'component', got '%s'", formatType("component"))
	}
}

func TestGetModuleDetails_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create a basic module structure
	modulePath := filepath.Join(tmpDir, DirComponents, "azurerm", "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	// Create main.tf
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	if details.Name != "storage-account" {
		t.Errorf("expected name 'storage-account', got '%s'", details.Name)
	}
	if details.Type != TypeComponent {
		t.Errorf("expected type '%s', got '%s'", TypeComponent, details.Type)
	}
	if details.HasSubmodules {
		t.Error("expected HasSubmodules to be false")
	}
	if details.HasTests {
		t.Error("expected HasTests to be false")
	}
	if details.HasExamples {
		t.Error("expected HasExamples to be false")
	}
	if len(details.Examples) != 0 {
		t.Errorf("expected 0 examples, got %d", len(details.Examples))
	}
	if details.SpaceliftVersion != "" {
		t.Errorf("expected empty SpaceliftVersion, got '%s'", details.SpaceliftVersion)
	}
}

func TestGetModuleDetails_WithAllFeatures(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create a module with all features
	modulePath := filepath.Join(tmpDir, DirComponents, "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	// Create main.tf
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create modules directory with content
	modulesDir := filepath.Join(modulePath, "modules", "submodule")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		t.Fatalf("failed to create modules directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulesDir, "main.tf"), []byte("# submodule"), 0644); err != nil {
		t.Fatalf("failed to create submodule main.tf: %v", err)
	}

	// Create tests directory with content
	testsDir := filepath.Join(modulePath, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("failed to create tests directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "basic_test.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create examples directory with examples (need main.tf to be detected)
	examplesDir := filepath.Join(modulePath, "examples")
	basicExampleDir := filepath.Join(examplesDir, "basic")
	if err := os.MkdirAll(basicExampleDir, 0755); err != nil {
		t.Fatalf("failed to create basic example: %v", err)
	}
	if err := os.WriteFile(filepath.Join(basicExampleDir, "main.tf"), []byte("# example"), 0644); err != nil {
		t.Fatalf("failed to create basic example main.tf: %v", err)
	}
	advancedExampleDir := filepath.Join(examplesDir, "advanced")
	if err := os.MkdirAll(advancedExampleDir, 0755); err != nil {
		t.Fatalf("failed to create advanced example: %v", err)
	}
	if err := os.WriteFile(filepath.Join(advancedExampleDir, "main.tf"), []byte("# example"), 0644); err != nil {
		t.Fatalf("failed to create advanced example main.tf: %v", err)
	}

	// Create spacelift config
	spaceliftDir := filepath.Join(modulePath, ".spacelift")
	if err := os.MkdirAll(spaceliftDir, 0755); err != nil {
		t.Fatalf("failed to create .spacelift directory: %v", err)
	}
	spaceliftConfig := "module_version: 2.1.0\nother_key: value"
	if err := os.WriteFile(filepath.Join(spaceliftDir, "config.yml"), []byte(spaceliftConfig), 0644); err != nil {
		t.Fatalf("failed to create spacelift config: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	if details.Name != "storage-account" {
		t.Errorf("expected name 'storage-account', got '%s'", details.Name)
	}
	if details.Type != TypeComponent {
		t.Errorf("expected type '%s', got '%s'", TypeComponent, details.Type)
	}
	if !details.HasSubmodules {
		t.Error("expected HasSubmodules to be true")
	}
	if !details.HasTests {
		t.Error("expected HasTests to be true")
	}
	if !details.HasExamples {
		t.Error("expected HasExamples to be true")
	}
	if len(details.Examples) != 2 {
		t.Errorf("expected 2 examples, got %d", len(details.Examples))
	}
	if details.SpaceliftVersion != "2.1.0" {
		t.Errorf("expected SpaceliftVersion '2.1.0', got '%s'", details.SpaceliftVersion)
	}
}

func TestGetModuleDetails_BaseType(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	modulePath := filepath.Join(tmpDir, DirBases, "k8s-argocd")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	if details.Type != TypeBase {
		t.Errorf("expected type '%s', got '%s'", TypeBase, details.Type)
	}
}

func TestGetModuleDetails_ProjectType(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	modulePath := filepath.Join(tmpDir, DirProjects, "prod-infra")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	if details.Type != TypeProject {
		t.Errorf("expected type '%s', got '%s'", TypeProject, details.Type)
	}
}

func TestGetModuleDetails_EmptyDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create a module with empty directories (should report false)
	modulePath := filepath.Join(tmpDir, DirComponents, "empty-dirs")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	// Create empty directories
	if err := os.MkdirAll(filepath.Join(modulePath, "modules"), 0755); err != nil {
		t.Fatalf("failed to create modules directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(modulePath, "tests"), 0755); err != nil {
		t.Fatalf("failed to create tests directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(modulePath, "examples"), 0755); err != nil {
		t.Fatalf("failed to create examples directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	// Empty directories should report false
	if details.HasSubmodules {
		t.Error("expected HasSubmodules to be false for empty directory")
	}
	if details.HasTests {
		t.Error("expected HasTests to be false for empty directory")
	}
	if details.HasExamples {
		t.Error("expected HasExamples to be false for empty directory")
	}
}

func TestShowCommand_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create resource-group like in demo
	modulePath := filepath.Join(tmpDir, DirComponents, "azurerm", "resource-group")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create examples like in demo
	examplesDir := filepath.Join(modulePath, "examples", "basic")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatalf("failed to create examples directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(examplesDir, "main.tf"), []byte("# example"), 0644); err != nil {
		t.Fatalf("failed to create example main.tf: %v", err)
	}

	// Create tests like in demo
	testsDir := filepath.Join(modulePath, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("failed to create tests directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "basic_test.go"), []byte("package tests"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// Test that we can find and get details for the module
	foundPath, err := findModuleInAllDirs("resource-group")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	details, err := getModuleDetails(foundPath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	if details.Name != "resource-group" {
		t.Errorf("expected name 'resource-group', got '%s'", details.Name)
	}
	if !details.HasTests {
		t.Error("expected HasTests to be true")
	}
	if !details.HasExamples {
		t.Error("expected HasExamples to be true")
	}
	if len(details.Examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(details.Examples))
	}
	if details.Examples[0].Name != "basic" {
		t.Errorf("expected example name 'basic', got '%s'", details.Examples[0].Name)
	}
}

// Edge case tests for show command

func TestGetModuleDetails_NoSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create a minimal module with just main.tf
	modulePath := filepath.Join(tmpDir, DirComponents, "azurerm", "simple-module")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	if details.HasSubmodules {
		t.Error("expected HasSubmodules to be false")
	}
	if details.HasTests {
		t.Error("expected HasTests to be false")
	}
	if details.HasExamples {
		t.Error("expected HasExamples to be false")
	}
	if len(details.Submodules) != 0 {
		t.Errorf("expected 0 submodules, got %d", len(details.Submodules))
	}
	if len(details.Tests) != 0 {
		t.Errorf("expected 0 tests, got %d", len(details.Tests))
	}
	if len(details.Examples) != 0 {
		t.Errorf("expected 0 examples, got %d", len(details.Examples))
	}
}

func TestGetModuleDetails_EmptySubdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create module with empty subdirectories (they exist but have no content)
	modulePath := filepath.Join(tmpDir, DirComponents, "azurerm", "with-empty-dirs")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	// Create empty subdirectories
	for _, dir := range []string{DirExamples, DirModules, DirTests} {
		if err := os.MkdirAll(filepath.Join(modulePath, dir), 0755); err != nil {
			t.Fatalf("failed to create %s directory: %v", dir, err)
		}
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	// Directories exist but are empty, so HasX should be false
	if details.HasSubmodules {
		t.Error("expected HasSubmodules to be false for empty directory")
	}
	if details.HasTests {
		t.Error("expected HasTests to be false for empty directory")
	}
	if details.HasExamples {
		t.Error("expected HasExamples to be false for empty directory")
	}
}

func TestListItems_WithVariablesTfOnly(t *testing.T) {
	tmpDir := t.TempDir()

	examplesDir := filepath.Join(tmpDir, DirExamples)

	// Create example with only variables.tf (no main.tf) - should still be included
	basicDir := filepath.Join(examplesDir, "basic")
	if err := os.MkdirAll(basicDir, 0755); err != nil {
		t.Fatalf("failed to create basic example: %v", err)
	}
	if err := os.WriteFile(filepath.Join(basicDir, "variables.tf"), []byte("# variables"), 0644); err != nil {
		t.Fatalf("failed to create variables.tf: %v", err)
	}

	items := listItems(examplesDir, tmpDir)

	if len(items) != 1 {
		t.Fatalf("expected 1 item (dir with .tf file), got %d", len(items))
	}

	if items[0].Name != "basic" {
		t.Errorf("expected item name 'basic', got '%s'", items[0].Name)
	}
}

func TestGetModuleDetails_UnknownModuleType(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create a module outside the standard directories
	modulePath := filepath.Join(tmpDir, "custom", "my-module")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create main.tf: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	details, err := getModuleDetails(modulePath)
	if err != nil {
		t.Fatalf("getModuleDetails returned error: %v", err)
	}

	// Type should be empty for modules not in standard directories
	if details.Type != "" {
		t.Errorf("expected empty type for non-standard module, got '%s'", details.Type)
	}
}
