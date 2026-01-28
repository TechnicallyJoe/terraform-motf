package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

func TestChangedCmd_HasFlags(t *testing.T) {
	// Test --ref flag
	ref := changedCmd.Flags().Lookup("ref")
	if ref == nil {
		t.Fatal("changedCmd should have --ref flag")
	}

	// Test --json flag
	jsonFlag := changedCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("changedCmd should have --json flag")
	}

	// Test --names flag
	namesFlag := changedCmd.Flags().Lookup("names")
	if namesFlag == nil {
		t.Fatal("changedCmd should have --names flag")
	}
}

func TestChangedCmd_ShortDescription(t *testing.T) {
	if changedCmd.Short == "" {
		t.Error("changedCmd should have a short description")
	}
	if changedCmd.Long == "" {
		t.Error("changedCmd should have a long description")
	}
}

func TestFindParentModule(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()

	// Create a module with a tests subdirectory
	moduleDir := filepath.Join(tmpDir, "components", "storage-account")
	testsDir := filepath.Join(moduleDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a .tf file in the module (not in tests)
	if err := os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte("# terraform"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a test file in tests/
	if err := os.WriteFile(filepath.Join(testsDir, "main_test.go"), []byte("package test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		startPath string
		stopPath  string
		want      string
	}{
		{
			name:      "finds parent module from tests dir",
			startPath: testsDir,
			stopPath:  tmpDir,
			want:      moduleDir,
		},
		{
			name:      "returns module dir when already at module",
			startPath: moduleDir,
			stopPath:  tmpDir,
			want:      moduleDir,
		},
		{
			name:      "returns empty when no parent module",
			startPath: filepath.Join(tmpDir, "components"),
			stopPath:  tmpDir,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findParentModule(tt.startPath, tt.stopPath)
			if got != tt.want {
				t.Errorf("findParentModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveChangedModules(t *testing.T) {
	// Create a temp directory structure simulating a repo
	tmpDir := t.TempDir()

	// Create module directories
	storageDir := filepath.Join(tmpDir, "components", "azurerm", "storage-account")
	kvDir := filepath.Join(tmpDir, "components", "azurerm", "key-vault")
	projectDir := filepath.Join(tmpDir, "projects", "prod-infra")

	for _, dir := range []string{storageDir, kvDir, projectDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte("# terraform"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Set up config
	withConfig(t, &config.Config{Root: "", Binary: "terraform"})

	tests := []struct {
		name         string
		changedPaths []string
		wantNames    []string
	}{
		{
			name:         "single module",
			changedPaths: []string{"components/azurerm/storage-account"},
			wantNames:    []string{"storage-account"},
		},
		{
			name:         "multiple modules",
			changedPaths: []string{"components/azurerm/storage-account", "components/azurerm/key-vault"},
			wantNames:    []string{"key-vault", "storage-account"},
		},
		{
			name:         "deduplicates",
			changedPaths: []string{"components/azurerm/storage-account", "components/azurerm/storage-account"},
			wantNames:    []string{"storage-account"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modules := resolveChangedModules(tmpDir, tmpDir, tt.changedPaths)

			var gotNames []string
			for _, m := range modules {
				gotNames = append(gotNames, m.Name)
			}

			if len(gotNames) != len(tt.wantNames) {
				t.Errorf("got %d modules, want %d", len(gotNames), len(tt.wantNames))
				return
			}

			for i, name := range gotNames {
				if name != tt.wantNames[i] {
					t.Errorf("module[%d] = %s, want %s", i, name, tt.wantNames[i])
				}
			}
		})
	}
}

func TestOutputChangedModules_EmptyJSON(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changedJsonFlag = true
	changedNamesOnlyFlag = false
	defer func() {
		changedJsonFlag = false
	}()

	err := outputChangedModules(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "[]" {
		t.Errorf("expected '[]', got '%s'", output)
	}
}

func TestOutputChangedModules_NamesOnly(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	changedJsonFlag = false
	changedNamesOnlyFlag = true
	defer func() {
		changedNamesOnlyFlag = false
	}()

	modules := []ModuleInfo{
		{Name: "storage-account", Type: "component", Path: "components/azurerm/storage-account"},
		{Name: "key-vault", Type: "component", Path: "components/azurerm/key-vault"},
	}

	err := outputChangedModules(modules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "storage-account" {
		t.Errorf("expected 'storage-account', got '%s'", lines[0])
	}
	if lines[1] != "key-vault" {
		t.Errorf("expected 'key-vault', got '%s'", lines[1])
	}
}
