package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/sturdy-parakeet/internal/config"
)

// Tests for getBasePath

func TestGetBasePath_EmptyRoot(t *testing.T) {
	tmpDir := t.TempDir()
	cfg = &config.Config{Root: "", Binary: "terraform"}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

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
	cfg = &config.Config{Root: "iac", Binary: "terraform"}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

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
	cfg = &config.Config{Root: tmpDir, Binary: "terraform"}

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
	pathFlag = ""

	_, err := resolveTargetPath([]string{})
	if err == nil {
		t.Error("expected error when no args are provided")
	}
}

func TestResolveTargetPath_PathMutuallyExclusive(t *testing.T) {
	pathFlag = "/some/path"

	_, err := resolveTargetPath([]string{"storage"})
	if err == nil {
		t.Error("expected error when path is combined with module name")
	}

	pathFlag = ""
}

func TestResolveTargetPath_WithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()

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

	pathFlag = ""
}

// Tests for findModuleInAllDirs

func TestFindModuleInAllDirs_ComponentFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	modulePath := filepath.Join(tmpDir, DirComponents, "azurerm", "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

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

	cfg = &config.Config{Root: "", Binary: "terraform"}

	modulePath := filepath.Join(tmpDir, DirBases, "k8s-argocd")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

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

	cfg = &config.Config{Root: "", Binary: "terraform"}

	modulePath := filepath.Join(tmpDir, DirProjects, "prod-infra")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

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

	cfg = &config.Config{Root: "", Binary: "terraform"}

	for _, dir := range ModuleDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create %s directory: %v", dir, err)
		}
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	_, err := findModuleInAllDirs("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent module")
	}
}

func TestFindModuleInAllDirs_NoDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	_, err := findModuleInAllDirs("any-module")
	if err == nil {
		t.Error("expected error when directories do not exist")
	}
}

func TestFindModuleInAllDirs_WithConfigRoot(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "iac", Binary: "terraform"}

	modulePath := filepath.Join(tmpDir, "iac", DirComponents, "storage-account")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

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

	cfg = &config.Config{Root: "", Binary: "terraform"}

	module1 := filepath.Join(tmpDir, DirComponents, "azurerm", "storage-account")
	module2 := filepath.Join(tmpDir, DirBases, "storage-account")

	for _, path := range []string{module1, module2} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create module directory: %v", err)
		}
		tfFile := filepath.Join(path, "main.tf")
		if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
			t.Fatalf("failed to create .tf file: %v", err)
		}
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	_, err := findModuleInAllDirs("storage-account")
	if err == nil {
		t.Error("expected error for name clash")
	}
}

// Tests for readModuleVersion

func TestReadModuleVersion_Found(t *testing.T) {
	tmpDir := t.TempDir()

	spaceliftDir := filepath.Join(tmpDir, ".spacelift")
	if err := os.MkdirAll(spaceliftDir, 0755); err != nil {
		t.Fatalf("failed to create .spacelift directory: %v", err)
	}

	configContent := `module_version: 1.2.3
other_key: value`
	if err := os.WriteFile(filepath.Join(spaceliftDir, "config.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config.yml: %v", err)
	}

	result := readModuleVersion(tmpDir)
	if result != "1.2.3" {
		t.Errorf("expected '1.2.3', got '%s'", result)
	}
}

func TestReadModuleVersion_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	result := readModuleVersion(tmpDir)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestReadModuleVersion_NoVersionKey(t *testing.T) {
	tmpDir := t.TempDir()

	spaceliftDir := filepath.Join(tmpDir, ".spacelift")
	if err := os.MkdirAll(spaceliftDir, 0755); err != nil {
		t.Fatalf("failed to create .spacelift directory: %v", err)
	}

	configContent := `other_key: value
another_key: 123`
	if err := os.WriteFile(filepath.Join(spaceliftDir, "config.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config.yml: %v", err)
	}

	result := readModuleVersion(tmpDir)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}
