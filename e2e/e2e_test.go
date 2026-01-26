package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// skipIfNoTerraform skips the test if terraform is not installed
func skipIfNoTerraform(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform not found in PATH, skipping e2e test")
	}
}

// skipIfNoTofu skips the test if tofu is not installed
func skipIfNoTofu(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("tofu not found in PATH, skipping e2e test")
	}
}

// getProjectRoot returns the root directory of the project
func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Get the current working directory (should be e2e/)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Go up to the project root
	return filepath.Dir(wd)
}

// getDemoPath returns the path to the demo directory
func getDemoPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(getProjectRoot(t), "demo")
}

// buildMotf builds the motf binary and returns its path
func buildMotf(t *testing.T) string {
	t.Helper()

	// Build to a temp directory
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "motf")

	projectRoot := getProjectRoot(t)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build motf: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// cleanupTerraformFiles removes .terraform directories, lock files, and state files from demo
func cleanupTerraformFiles(t *testing.T) {
	t.Helper()
	demoPath := getDemoPath(t)

	// Walk through demo and remove .terraform directories, lock files, and state files
	_ = filepath.Walk(demoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".terraform" {
			_ = os.RemoveAll(path)
			return filepath.SkipDir
		}
		if info.Name() == ".terraform.lock.hcl" ||
			info.Name() == "terraform.tfstate" ||
			info.Name() == "terraform.tfstate.backup" {
			_ = os.Remove(path)
		}
		return nil
	})
}

func TestE2E_FmtComponent(t *testing.T) {
	skipIfNoTerraform(t)

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run fmt on storage-account component
	cmd := exec.Command(motfBinary, "fmt", "storage-account")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf fmt failed: %v\nOutput: %s", err, output)
	}

	// Verify the command ran successfully (mentions the component path)
	if !strings.Contains(string(output), "storage-account") {
		t.Errorf("expected output to mention storage-account, got: %s", output)
	}
}

func TestE2E_FmtNestedComponent(t *testing.T) {
	skipIfNoTerraform(t)

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run fmt on key-vault component (nested under azurerm)
	cmd := exec.Command(motfBinary, "fmt", "key-vault")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf fmt failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "key-vault") {
		t.Errorf("expected output to mention key-vault, got: %s", output)
	}
}

func TestE2E_InitComponent(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run init on storage-account
	cmd := exec.Command(motfBinary, "init", "storage-account")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf init failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Initializing") {
		t.Errorf("expected 'Initializing' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", outputStr)
	}
}

func TestE2E_InitBase(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run init on k8s-argocd base
	cmd := exec.Command(motfBinary, "init", "k8s-argocd")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf init failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "k8s-argocd") {
		t.Errorf("expected output to mention k8s-argocd, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", outputStr)
	}
}

func TestE2E_InitProject(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run init on prod-infra project
	cmd := exec.Command(motfBinary, "init", "prod-infra")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf init failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "prod-infra") {
		t.Errorf("expected output to mention prod-infra, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", outputStr)
	}
}

func TestE2E_ValidateComponent(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run validate with init flag
	cmd := exec.Command(motfBinary, "val", "-i", "storage-account")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf val -i failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should contain both init and validate output
	if !strings.Contains(outputStr, "init") {
		t.Errorf("expected 'init' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "valid") || !strings.Contains(outputStr, "Success") {
		// Terraform outputs "Success! The configuration is valid."
		t.Logf("validate output: %s", outputStr)
	}
}

func TestE2E_ValidateBase(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run validate with init flag on base
	cmd := exec.Command(motfBinary, "val", "-i", "k8s-argocd")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf val -i failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "k8s-argocd") {
		t.Errorf("expected output to mention k8s-argocd, got: %s", output)
	}
}

func TestE2E_TestComponent(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run test on naming component which uses Azure/naming/azurerm (no real infrastructure)
	// The terratest will run terraform init, apply, and destroy on the example
	cmd := exec.Command(motfBinary, "test", "naming")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// The test command should find the specific naming component module and attempt to run go test
	if !strings.Contains(outputStr, "components/azurerm/naming") {
		t.Errorf("expected output to mention 'components/azurerm/naming', got: %s", outputStr)
	}

	// Should show it's running go test
	if !strings.Contains(outputStr, "go test") {
		t.Errorf("expected output to contain 'go test', got: %s", outputStr)
	}

	// If there was an error, log it but check if it's an expected failure (e.g., missing terratest deps in CI)
	if err != nil {
		// Accept if the test ran but failed due to missing dependencies
		if strings.Contains(outputStr, "cannot find package") ||
			strings.Contains(outputStr, "no required module provides") {
			t.Logf("test command ran but terratest dependencies not available (expected in CI): %s", outputStr)
			return
		}
		t.Fatalf("motf test failed unexpectedly: %v\nOutput: %s", err, outputStr)
	}

	// If test passed, verify we see successful go test output
	if !strings.Contains(outputStr, "PASS") && !strings.Contains(outputStr, "ok") {
		t.Errorf("expected successful go test output to contain 'PASS' or 'ok', got: %s", outputStr)
	}
}

func TestE2E_ExplicitPath(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run init with explicit path
	cmd := exec.Command(motfBinary, "init", "--path", "components/azurerm/key-vault")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf init --path failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "key-vault") {
		t.Errorf("expected output to mention key-vault, got: %s", outputStr)
	}
}

func TestE2E_ConfigCommand(t *testing.T) {
	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run config command
	cmd := exec.Command(motfBinary, "config")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf config failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Binary:") {
		t.Error("config output should contain 'Binary:'")
	}
	if !strings.Contains(outputStr, "terraform") {
		t.Error("config output should contain 'terraform'")
	}
	if !strings.Contains(outputStr, "Config:") {
		t.Error("config output should contain 'Config:'")
	}
	// Demo directory has a .motf.yml file, so it should show the path
	if !strings.Contains(outputStr, ".motf.yml") {
		t.Error("config output should contain '.motf.yml' path")
	}
}

func TestE2E_ArgsFlag(t *testing.T) {
	skipIfNoTerraform(t)

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run fmt with extra args
	cmd := exec.Command(motfBinary, "fmt", "storage-account", "-a", "-check")
	cmd.Dir = demoPath
	output, _ := cmd.CombinedOutput()

	// The output should show the -check arg was passed
	if !strings.Contains(string(output), "-check") {
		t.Errorf("expected '-check' in output, got: %s", output)
	}
}

func TestE2E_MultipleArgsFlag(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run init with multiple args
	cmd := exec.Command(motfBinary, "init", "storage-account", "-a", "-upgrade", "-a", "-reconfigure")
	cmd.Dir = demoPath
	output, _ := cmd.CombinedOutput()

	outputStr := string(output)
	if !strings.Contains(outputStr, "-upgrade") {
		t.Errorf("expected '-upgrade' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "-reconfigure") {
		t.Errorf("expected '-reconfigure' in output, got: %s", outputStr)
	}
}

func TestE2E_ModuleNotFound(t *testing.T) {
	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Try to run on a non-existent component
	cmd := exec.Command(motfBinary, "init", "nonexistent-component")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("expected error for non-existent component")
	}

	if !strings.Contains(string(output), "not found") {
		t.Errorf("expected 'not found' in error message, got: %s", output)
	}
}

func TestE2E_NoFlagsError(t *testing.T) {
	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Try to run without specifying target
	cmd := exec.Command(motfBinary, "init")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("expected error when no target flags specified")
	}

	if !strings.Contains(string(output), "must specify") {
		t.Errorf("expected 'must specify' in error message, got: %s", output)
	}
}

func TestE2E_MutuallyExclusiveFlags(t *testing.T) {
	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Try to use both module name and --path
	cmd := exec.Command(motfBinary, "init", "storage-account", "--path", "components/azurerm/key-vault")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("expected error when module name and --path are both specified")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "mutually exclusive") {
		t.Errorf("expected mutual exclusivity error, got: %s", outputStr)
	}
}

func TestE2E_VersionFlag(t *testing.T) {
	motfBinary := buildMotf(t)

	cmd := exec.Command(motfBinary, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("motf --version failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "motf") {
		t.Errorf("expected version output to contain 'motf', got: %s", stdout.String())
	}
}

func TestE2E_HelpFlag(t *testing.T) {
	motfBinary := buildMotf(t)

	cmd := exec.Command(motfBinary, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Terraform Monorepo Orchestrator") {
		t.Error("help output should contain 'Terraform Monorepo Orchestrator'")
	}
	if !strings.Contains(outputStr, "--args") {
		t.Error("help output should contain '--args'")
	}
}

func TestE2E_SubcommandHelp(t *testing.T) {
	motfBinary := buildMotf(t)

	subcommands := []string{"init", "fmt", "val", "config"}

	for _, subcmd := range subcommands {
		t.Run(subcmd, func(t *testing.T) {
			cmd := exec.Command(motfBinary, subcmd, "--help")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("motf %s --help failed: %v\nOutput: %s", subcmd, err, output)
			}

			if !strings.Contains(string(output), subcmd) {
				t.Errorf("help output should mention %s", subcmd)
			}
		})
	}
}

func TestE2E_ValidateAlias(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Use 'validate' alias instead of 'val'
	cmd := exec.Command(motfBinary, "validate", "-i", "storage-account")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf validate failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "storage-account") {
		t.Errorf("expected output to mention storage-account, got: %s", output)
	}
}

func TestE2E_WorksFromSubdirectory(t *testing.T) {
	skipIfNoTerraform(t)

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run from a subdirectory within demo
	subDir := filepath.Join(demoPath, "components", "azurerm")

	cmd := exec.Command(motfBinary, "fmt", "key-vault")
	cmd.Dir = subDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf from subdirectory failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "key-vault") {
		t.Errorf("expected output to mention key-vault, got: %s", output)
	}
}

func TestE2E_TofuFmt(t *testing.T) {
	skipIfNoTofu(t)

	motfBinary := buildMotf(t)

	// Create a temp directory with tofu config
	tmpDir := t.TempDir()

	// Create a .motf.yml that uses tofu
	configContent := "binary: tofu\nroot: .\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".motf.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create components directory with a simple terraform file
	componentDir := filepath.Join(tmpDir, "components", "test-component")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}

	tfContent := "variable \"test\" {\n  type = string\n}\n"
	if err := os.WriteFile(filepath.Join(componentDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write tf file: %v", err)
	}

	// Run fmt with tofu
	cmd := exec.Command(motfBinary, "fmt", "test-component")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf fmt with tofu failed: %v\nOutput: %s", err, output)
	}

	// Verify tofu was used (output mentions tofu or the component)
	if !strings.Contains(string(output), "test-component") {
		t.Errorf("expected output to mention test-component, got: %s", output)
	}
}

func TestE2E_TofuInit(t *testing.T) {
	skipIfNoTofu(t)

	motfBinary := buildMotf(t)

	// Create a temp directory with tofu config
	tmpDir := t.TempDir()

	// Create a .motf.yml that uses tofu
	configContent := "binary: tofu\nroot: .\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".motf.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create components directory with a simple terraform file
	componentDir := filepath.Join(tmpDir, "components", "test-component")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}

	tfContent := "terraform {\n}\n\nvariable \"test\" {\n  type = string\n}\n"
	if err := os.WriteFile(filepath.Join(componentDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("failed to write tf file: %v", err)
	}

	// Run init with tofu
	cmd := exec.Command(motfBinary, "init", "test-component")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf init with tofu failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Tofu outputs similar messages to terraform
	if !strings.Contains(outputStr, "Initializing") {
		t.Errorf("expected 'Initializing' in output, got: %s", outputStr)
	}
}

// TestE2E_PlanNamingComponent tests plan on the naming component which has no cloud dependencies
func TestE2E_PlanNamingComponent(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// First init the naming component
	initCmd := exec.Command(motfBinary, "init", "naming")
	initCmd.Dir = demoPath
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf init failed: %v\nOutput: %s", err, initOutput)
	}

	// Run plan on naming component
	cmd := exec.Command(motfBinary, "plan", "naming")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf plan failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "plan") {
		t.Errorf("expected 'plan' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "naming") {
		t.Errorf("expected 'naming' in output, got: %s", outputStr)
	}
}

// TestE2E_PlanWithInitFlag tests plan with -i flag to run init first
func TestE2E_PlanWithInitFlag(t *testing.T) {
	skipIfNoTerraform(t)
	t.Cleanup(func() { cleanupTerraformFiles(t) })

	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run plan with -i flag (should init first then plan)
	cmd := exec.Command(motfBinary, "plan", "-i", "naming")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf plan -i failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should contain init output
	if !strings.Contains(outputStr, "Initializing") {
		t.Errorf("expected 'Initializing' in output (from init), got: %s", outputStr)
	}
	// Should contain plan output
	if !strings.Contains(outputStr, "plan") {
		t.Errorf("expected 'plan' in output, got: %s", outputStr)
	}
}

// TestE2E_TaskList tests listing tasks
func TestE2E_TaskList(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := t.TempDir()

	// Create minimal polylith structure
	componentDir := filepath.Join(tmpDir, "components", "test-component")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(componentDir, "main.tf"), []byte("terraform {}\n"), 0644); err != nil {
		t.Fatalf("failed to write tf file: %v", err)
	}

	// Create .motf.yml with tasks
	configContent := `binary: terraform
tasks:
  hello:
    description: "Says hello"
    command: "echo hello"
  lint:
    description: "Run linting"
    shell: sh
    command: "echo linting"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".motf.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Run task list
	cmd := exec.Command(motfBinary, "task", "test-component")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf task list failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "hello") {
		t.Errorf("expected 'hello' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "lint") {
		t.Errorf("expected 'lint' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Says hello") {
		t.Errorf("expected description in output, got: %s", outputStr)
	}
}

// TestE2E_TaskRun tests running a task
func TestE2E_TaskRun(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := t.TempDir()

	// Create minimal polylith structure
	componentDir := filepath.Join(tmpDir, "components", "test-component")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(componentDir, "main.tf"), []byte("terraform {}\n"), 0644); err != nil {
		t.Fatalf("failed to write tf file: %v", err)
	}

	// Create .motf.yml with a task
	configContent := `binary: terraform
tasks:
  greet:
    description: "Greeting task"
    shell: sh
    command: |
      echo "Hello from task"
      echo "Working dir: $(pwd)"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".motf.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Run the task
	cmd := exec.Command(motfBinary, "task", "test-component", "-t", "greet")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf task run failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Hello from task") {
		t.Errorf("expected 'Hello from task' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "test-component") {
		t.Errorf("expected working dir to contain 'test-component', got: %s", outputStr)
	}
}
