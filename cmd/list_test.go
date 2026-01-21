package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/sturdy-parakeet/internal/config"
)

func TestCollectModules_AllTypes(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create modules in all directories
	componentPath := filepath.Join(tmpDir, DirComponents, "storage")
	basePath := filepath.Join(tmpDir, DirBases, "argocd")
	projectPath := filepath.Join(tmpDir, DirProjects, "prod")

	for _, path := range []string{componentPath, basePath, projectPath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		tfFile := filepath.Join(path, "main.tf")
		if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
			t.Fatalf("failed to create .tf file: %v", err)
		}
	}

	modules, err := collectModules(tmpDir, "")
	if err != nil {
		t.Fatalf("collectModules returned error: %v", err)
	}

	if len(modules) != 3 {
		t.Errorf("expected 3 modules, got %d", len(modules))
	}

	// Verify types are correctly identified
	typeCount := make(map[string]int)
	for _, mod := range modules {
		typeCount[mod.Type]++
	}

	if typeCount[TypeComponent] != 1 {
		t.Errorf("expected 1 component, got %d", typeCount[TypeComponent])
	}
	if typeCount[TypeBase] != 1 {
		t.Errorf("expected 1 base, got %d", typeCount[TypeBase])
	}
	if typeCount[TypeProject] != 1 {
		t.Errorf("expected 1 project, got %d", typeCount[TypeProject])
	}
}

func TestCollectModules_WithFilter(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create multiple modules
	for _, name := range []string{"storage-account", "storage-blob", "network"} {
		path := filepath.Join(tmpDir, DirComponents, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		tfFile := filepath.Join(path, "main.tf")
		if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
			t.Fatalf("failed to create .tf file: %v", err)
		}
	}

	modules, err := collectModules(tmpDir, "storage*")
	if err != nil {
		t.Fatalf("collectModules returned error: %v", err)
	}

	if len(modules) != 2 {
		t.Errorf("expected 2 modules matching 'storage*', got %d", len(modules))
	}
}

func TestCollectModules_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()

	cfg = &config.Config{Root: "", Binary: "terraform"}

	// Create directories but no modules
	for _, dir := range ModuleDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
	}

	modules, err := collectModules(tmpDir, "")
	if err != nil {
		t.Fatalf("collectModules returned error: %v", err)
	}

	if len(modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(modules))
	}
}

func TestSortModules_ByTypeAndName(t *testing.T) {
	modules := []ModuleInfo{
		{Name: "zebra", Type: TypeProject},
		{Name: "alpha", Type: TypeComponent},
		{Name: "beta", Type: TypeBase},
		{Name: "gamma", Type: TypeComponent},
		{Name: "delta", Type: TypeBase},
	}

	sortModules(modules)

	// Expected order: components (alpha, gamma), bases (beta, delta), projects (zebra)
	expected := []struct {
		name    string
		modType string
	}{
		{"alpha", TypeComponent},
		{"gamma", TypeComponent},
		{"beta", TypeBase},
		{"delta", TypeBase},
		{"zebra", TypeProject},
	}

	for i, exp := range expected {
		if modules[i].Name != exp.name || modules[i].Type != exp.modType {
			t.Errorf("position %d: expected {%s, %s}, got {%s, %s}",
				i, exp.name, exp.modType, modules[i].Name, modules[i].Type)
		}
	}
}

func TestSortModules_EmptySlice(t *testing.T) {
	modules := []ModuleInfo{}

	// Should not panic
	sortModules(modules)

	if len(modules) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(modules))
	}
}

func TestSortModules_SingleElement(t *testing.T) {
	modules := []ModuleInfo{
		{Name: "only", Type: TypeComponent},
	}

	sortModules(modules)

	if modules[0].Name != "only" {
		t.Errorf("expected 'only', got '%s'", modules[0].Name)
	}
}
