package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

// Tests for getBasePath

func TestGetBasePath_EmptyRoot(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	result, err := getBasePath()
	if err != nil {
		t.Fatalf("getBasePath returned error: %v", err)
	}

	if result != tmpDir {
		t.Errorf("expected '%s', got '%s'", tmpDir, result)
	}
}

func TestGetBasePath_RelativeRoot(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "iac", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	result, err := getBasePath()
	if err != nil {
		t.Fatalf("getBasePath returned error: %v", err)
	}

	expected := filepath.Join(tmpDir, "iac")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestGetBasePath_AbsoluteRoot(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: tmpDir, Binary: "terraform"})

	result, err := getBasePath()
	if err != nil {
		t.Fatalf("getBasePath returned error: %v", err)
	}

	if result != tmpDir {
		t.Errorf("expected '%s', got '%s'", tmpDir, result)
	}
}

// Tests for getModuleType

func TestGetModuleType_Component(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/repo/components/storage", TypeComponent},
		{"C:\\repo\\components\\storage", TypeComponent},
		{"/repo/bases/argocd", TypeBase},
		{"C:\\repo\\bases\\argocd", TypeBase},
		{"/repo/projects/prod", TypeProject},
		{"C:\\repo\\projects\\prod", TypeProject},
		{"/repo/other/module", ""},
	}

	for _, tt := range tests {
		result := getModuleType(tt.path)
		if result != tt.expected {
			t.Errorf("getModuleType(%s) = '%s', expected '%s'", tt.path, result, tt.expected)
		}
	}
}

// Tests for resolveExplicitPath

func TestResolveExplicitPath_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	testPath := filepath.Join(tmpDir, "test-module")
	if err := os.MkdirAll(testPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	result, err := resolveExplicitPath(testPath)
	if err != nil {
		t.Fatalf("resolveExplicitPath returned error: %v", err)
	}

	if result != testPath {
		t.Errorf("expected '%s', got '%s'", testPath, result)
	}
}

func TestResolveExplicitPath_NonExistent(t *testing.T) {
	_, err := resolveExplicitPath("/non/existent/path")
	if err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

// Tests for resolveTargetPath

func TestResolveTargetPath_NoArgs(t *testing.T) {
	resetFlags(t)

	_, err := resolveTargetPath([]string{})
	if err == nil {
		t.Error("expected error when no args are provided")
	}
}

func TestResolveTargetPath_PathMutuallyExclusive(t *testing.T) {
	resetFlags(t)
	pathFlag = "/some/path"

	_, err := resolveTargetPath([]string{"storage"})
	if err == nil {
		t.Error("expected error when path is combined with module name")
	}
}

func TestResolveTargetPath_WithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)

	testPath := filepath.Join(tmpDir, "my-module")
	if err := os.MkdirAll(testPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	pathFlag = testPath

	result, err := resolveTargetPath([]string{})
	if err != nil {
		t.Fatalf("resolveTargetPath returned error: %v", err)
	}

	if result != testPath {
		t.Errorf("expected '%s', got '%s'", testPath, result)
	}
}

// Tests for findModuleInAllDirs

func TestFindModuleInAllDirs_ComponentFound(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))

	result, err := findModuleInAllDirs("storage-account")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestFindModuleInAllDirs_BaseFound(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirBases, "k8s-argocd"))

	result, err := findModuleInAllDirs("k8s-argocd")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestFindModuleInAllDirs_ProjectFound(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirProjects, "prod-infra"))

	result, err := findModuleInAllDirs("prod-infra")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestFindModuleInAllDirs_ModuleNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	for _, dir := range ModuleDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create %s directory: %v", dir, err)
		}
	}

	_, err := findModuleInAllDirs("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent module")
	}
}

func TestFindModuleInAllDirs_NoDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	_, err := findModuleInAllDirs("any-module")
	if err == nil {
		t.Error("expected error when directories do not exist")
	}
}

func TestFindModuleInAllDirs_WithConfigRoot(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "iac", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join("iac", DirComponents, "storage-account"))

	result, err := findModuleInAllDirs("storage-account")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestFindModuleInAllDirs_NameClash(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))
	createTerraformModule(t, tmpDir, filepath.Join(DirBases, "storage-account"))

	_, err := findModuleInAllDirs("storage-account")
	if err == nil {
		t.Error("expected error for name clash")
	}
}

// Tests for resolveTargetWithExample

func TestResolveTargetWithExample_NoExample(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))

	result, err := resolveTargetWithExample([]string{"storage-account"}, "")
	if err != nil {
		t.Fatalf("resolveTargetWithExample returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestResolveTargetWithExample_WithExample(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))
	examplePath := filepath.Join(modulePath, "examples", "basic")
	if err := os.MkdirAll(examplePath, 0755); err != nil {
		t.Fatalf("failed to create example directory: %v", err)
	}

	// Create main.tf in the example
	exampleTfFile := filepath.Join(examplePath, "main.tf")
	if err := os.WriteFile(exampleTfFile, []byte("# example terraform"), 0644); err != nil {
		t.Fatalf("failed to create example main.tf: %v", err)
	}

	result, err := resolveTargetWithExample([]string{"storage-account"}, "basic")
	if err != nil {
		t.Fatalf("resolveTargetWithExample returned error: %v", err)
	}

	if result != examplePath {
		t.Errorf("expected '%s', got '%s'", examplePath, result)
	}
}

func TestResolveTargetWithExample_ExampleNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))

	_, err := resolveTargetWithExample([]string{"storage-account"}, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent example, got nil")
	}
}

func TestResolveTargetWithExample_ExampleMissingTfFiles(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))
	examplePath := filepath.Join(modulePath, "examples", "basic")
	if err := os.MkdirAll(examplePath, 0755); err != nil {
		t.Fatalf("failed to create example directory: %v", err)
	}

	// No .tf files in example directory

	_, err := resolveTargetWithExample([]string{"storage-account"}, "basic")
	if err == nil {
		t.Error("expected error for example without .tf files, got nil")
	}
}

func TestResolveTargetWithExample_ExampleWithNonMainTf(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "storage-account"))
	examplePath := filepath.Join(modulePath, "examples", "basic")
	if err := os.MkdirAll(examplePath, 0755); err != nil {
		t.Fatalf("failed to create example directory: %v", err)
	}

	// Create only variables.tf in the example (not main.tf)
	exampleTfFile := filepath.Join(examplePath, "variables.tf")
	if err := os.WriteFile(exampleTfFile, []byte("# example variables"), 0644); err != nil {
		t.Fatalf("failed to create example variables.tf: %v", err)
	}

	result, err := resolveTargetWithExample([]string{"storage-account"}, "basic")
	if err != nil {
		t.Fatalf("resolveTargetWithExample returned error: %v", err)
	}

	if result != examplePath {
		t.Errorf("expected '%s', got '%s'", examplePath, result)
	}
}
