package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

// Tests for root command flags

func TestArgsFlag_Empty(t *testing.T) {
	resetFlags(t)

	if len(argsFlag) != 0 {
		t.Errorf("expected empty argsFlag, got %v", argsFlag)
	}
}

func TestArgsFlag_SingleArg(t *testing.T) {
	resetFlags(t)
	argsFlag = []string{"-upgrade"}

	if len(argsFlag) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(argsFlag))
	}

	if argsFlag[0] != "-upgrade" {
		t.Errorf("expected '-upgrade', got '%s'", argsFlag[0])
	}
}

func TestArgsFlag_MultipleArgs(t *testing.T) {
	resetFlags(t)
	argsFlag = []string{"-upgrade", "-reconfigure", "-backend=false"}

	if len(argsFlag) != 3 {
		t.Fatalf("expected 3 args, got %d", len(argsFlag))
	}

	expected := []string{"-upgrade", "-reconfigure", "-backend=false"}
	for i, arg := range argsFlag {
		if arg != expected[i] {
			t.Errorf("arg[%d] = '%s', expected '%s'", i, arg, expected[i])
		}
	}
}

func TestArgsFlag_PreservesOrder(t *testing.T) {
	resetFlags(t)
	argsFlag = []string{"-var=foo=bar", "-var=baz=qux", "-target=module.test"}

	expected := []string{"-var=foo=bar", "-var=baz=qux", "-target=module.test"}
	for i, arg := range argsFlag {
		if arg != expected[i] {
			t.Errorf("order not preserved: got %v, expected %v", argsFlag, expected)
			break
		}
	}
}

func TestPathFlag_Reset(t *testing.T) {
	resetFlags(t)
	pathFlag = "/some/path"

	if pathFlag != "/some/path" {
		t.Errorf("expected '/some/path', got '%s'", pathFlag)
	}
}

func TestInitFlag_Default(t *testing.T) {
	resetFlags(t)

	if initFlag != false {
		t.Error("expected initFlag to be false by default")
	}
}

func TestSearchFlag_Default(t *testing.T) {
	resetFlags(t)

	if searchFlag != "" {
		t.Errorf("expected empty searchFlag, got '%s'", searchFlag)
	}
}

// Integration-style tests that verify module finding for test command scenarios

func TestTestCommand_WithModuleName(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{
		Root:   tmpDir,
		Binary: "terraform",
		Test:   &config.TestConfig{Engine: "terratest", Args: ""},
	})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "test-module"))

	// Add go.mod and test file for terratest
	goMod := filepath.Join(modulePath, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	testFile := filepath.Join(modulePath, "module_test.go")
	testContent := `package test

import "testing"

func TestExample(t *testing.T) {
	t.Log("test passed")
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := findModuleInAllDirs("test-module")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestTestCommand_WithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	resetFlags(t)

	modulePath := filepath.Join(tmpDir, "test-module")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	pathFlag = modulePath
	result, err := resolveTargetPath([]string{})

	if err != nil {
		t.Fatalf("resolveTargetPath returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}

func TestTestCommand_WithArgs(t *testing.T) {
	resetFlags(t)
	testArgs := []string{"-v", "-timeout=30m", "-count=1"}
	argsFlag = testArgs

	if len(argsFlag) != len(testArgs) {
		t.Fatalf("expected %d args, got %d", len(testArgs), len(argsFlag))
	}

	for i, arg := range argsFlag {
		if arg != testArgs[i] {
			t.Errorf("arg[%d] = '%s', expected '%s'", i, arg, testArgs[i])
		}
	}
}

func TestTestCommand_FindsResourceGroup(t *testing.T) {
	tmpDir := t.TempDir()
	withConfig(t, &config.Config{
		Root:   tmpDir,
		Binary: "terraform",
		Test:   &config.TestConfig{Engine: "terratest", Args: ""},
	})
	withWorkingDir(t, tmpDir)

	modulePath := createTerraformModule(t, tmpDir, filepath.Join(DirComponents, "azurerm", "resource-group"))

	// Add tests directory with test files
	testsPath := filepath.Join(modulePath, "tests")
	if err := os.MkdirAll(testsPath, 0755); err != nil {
		t.Fatalf("failed to create tests directory: %v", err)
	}

	goMod := filepath.Join(testsPath, "go.mod")
	if err := os.WriteFile(goMod, []byte("module tests\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	testFile := filepath.Join(testsPath, "basic_test.go")
	testContent := `package tests

import "testing"

func TestBasicExample(t *testing.T) {
	t.Log("resource-group test")
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := findModuleInAllDirs("resource-group")
	if err != nil {
		t.Fatalf("findModuleInAllDirs returned error: %v", err)
	}

	if result != modulePath {
		t.Errorf("expected '%s', got '%s'", modulePath, result)
	}
}
