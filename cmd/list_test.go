package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
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

func TestSortModules_ByPath(t *testing.T) {
	modules := []ModuleInfo{
		{Name: "zebra", Type: TypeProject, Path: "projects/zebra"},
		{Name: "alpha", Type: TypeComponent, Path: "components/alpha"},
		{Name: "beta", Type: TypeBase, Path: "bases/beta"},
		{Name: "gamma", Type: TypeComponent, Path: "components/gamma"},
		{Name: "delta", Type: TypeBase, Path: "bases/delta"},
	}

	sortModules(modules)

	// Expected order: sorted alphabetically by path
	expected := []struct {
		name string
		path string
	}{
		{"beta", "bases/beta"},
		{"delta", "bases/delta"},
		{"alpha", "components/alpha"},
		{"gamma", "components/gamma"},
		{"zebra", "projects/zebra"},
	}

	for i, exp := range expected {
		if modules[i].Name != exp.name || modules[i].Path != exp.path {
			t.Errorf("position %d: expected {%s, %s}, got {%s, %s}",
				i, exp.name, exp.path, modules[i].Name, modules[i].Path)
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
