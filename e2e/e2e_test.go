package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/motf")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build motf: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// initGitRepo initializes a git repository in the given directory with user config.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to init git repo: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	_ = cmd.Run()
}

// createModules creates terraform modules under components/ in the given directory.
func createModules(t *testing.T, dir string, names []string) {
	t.Helper()
	for _, name := range names {
		moduleDir := filepath.Join(dir, "components", name)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			t.Fatalf("failed to create module dir %s: %v", name, err)
		}
		content := fmt.Sprintf("# %s\nterraform {}\n", name)
		if err := os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write tf file for %s: %v", name, err)
		}
	}
}

// commitAll stages and commits all changes in the given directory.
func commitAll(t *testing.T, dir, message string) {
	t.Helper()
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to commit: %v\nOutput: %s", err, output)
	}
}

// addUncommittedFile creates a new file in each module to simulate uncommitted changes.
func addUncommittedFile(t *testing.T, dir string, modules []string, filename, contentFmt string) {
	t.Helper()
	for _, mod := range modules {
		moduleDir := filepath.Join(dir, "components", mod)
		content := fmt.Sprintf(contentFmt, mod)
		if err := os.WriteFile(filepath.Join(moduleDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s for %s: %v", filename, mod, err)
		}
	}
}

// setupCleanGitRepo creates a temp directory with a git repo and polylith structure for testing.
// Returns the path to the temp directory.
func setupCleanGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	createModules(t, tmpDir, []string{"test-module"})
	commitAll(t, tmpDir, "initial")
	return tmpDir
}

// setupGitRepoWithModules creates a temp directory with a git repo and specified modules.
// All modules are committed. Returns the path to the temp directory.
func setupGitRepoWithModules(t *testing.T, modules []string) string {
	t.Helper()
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	createModules(t, tmpDir, modules)
	commitAll(t, tmpDir, "initial")
	return tmpDir
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
	if !strings.Contains(outputStr, "binary:") {
		t.Error("config output should contain 'binary:'")
	}
	if !strings.Contains(outputStr, "terraform") {
		t.Error("config output should contain 'terraform'")
	}
	if !strings.Contains(outputStr, "Config file:") {
		t.Error("config output should contain 'Config file:'")
	}
	// Demo directory has a .motf.yml file, so it should show the path
	if !strings.Contains(outputStr, ".motf.yml") {
		t.Error("config output should contain '.motf.yml' path")
	}
}

func TestE2E_DescribeCommand(t *testing.T) {
	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run describe on storage-account module (has variables, outputs, providers)
	cmd := exec.Command(motfBinary, "describe", "storage-account")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf describe failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check basic structure
	if !strings.Contains(outputStr, "Module:") {
		t.Error("describe output should contain 'Module:'")
	}
	if !strings.Contains(outputStr, "storage-account") {
		t.Error("describe output should contain module name 'storage-account'")
	}
	if !strings.Contains(outputStr, "Path:") {
		t.Error("describe output should contain 'Path:'")
	}
	if !strings.Contains(outputStr, "Variables:") {
		t.Error("describe output should contain 'Variables:'")
	}
	if !strings.Contains(outputStr, "Outputs:") {
		t.Error("describe output should contain 'Outputs:'")
	}
}

func TestE2E_DescribeCommand_JSON(t *testing.T) {
	motfBinary := buildMotf(t)
	demoPath := getDemoPath(t)

	// Run describe with --json flag
	cmd := exec.Command(motfBinary, "describe", "storage-account", "--json")
	cmd.Dir = demoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf describe --json failed: %v\nOutput: %s", err, output)
	}

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("describe --json output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Check expected fields
	if result["name"] != "storage-account" {
		t.Errorf("expected name 'storage-account', got '%v'", result["name"])
	}
	if _, ok := result["path"]; !ok {
		t.Error("JSON output should contain 'path' field")
	}
	if _, ok := result["variables"]; !ok {
		t.Error("JSON output should contain 'variables' field")
	}
	if _, ok := result["outputs"]; !ok {
		t.Error("JSON output should contain 'outputs' field")
	}
}

func TestE2E_ChangedCommand(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := setupCleanGitRepo(t)

	// Run changed with ref HEAD (no commits to compare, no uncommitted changes)
	cmd := exec.Command(motfBinary, "changed", "--ref", "HEAD")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf changed failed: %v\nOutput: %s", err, output)
	}

	// Should output "No changed modules found" in a clean repo
	if !strings.Contains(string(output), "No changed modules found") {
		t.Errorf("expected 'No changed modules found' in clean repo, got: %s", output)
	}
}

func TestE2E_ChangedCommand_JSON(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := setupCleanGitRepo(t)

	// Run changed with --json flag
	cmd := exec.Command(motfBinary, "changed", "--ref", "HEAD", "--json")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf changed --json failed: %v\nOutput: %s", err, output)
	}

	// Should be valid JSON (empty array)
	var result []interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("changed --json output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Should be empty in a clean repo
	if len(result) != 0 {
		t.Errorf("expected empty array in clean repo, got: %v", result)
	}
}

func TestE2E_ChangedCommand_Names(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := setupCleanGitRepo(t)

	// Run changed with --names flag
	cmd := exec.Command(motfBinary, "changed", "--ref", "HEAD", "--names")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf changed --names failed: %v\nOutput: %s", err, output)
	}

	// Should be empty (no output) in a clean repo
	if strings.TrimSpace(string(output)) != "" {
		t.Errorf("expected empty output in clean repo with --names, got: %s", output)
	}
}

func TestE2E_FmtChanged_NoOp(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := setupCleanGitRepo(t)

	// In a clean repo with no changes, fmt --changed should be a no-op.
	cmd := exec.Command(motfBinary, "fmt", "--changed", "--ref", "HEAD")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf fmt --changed failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(string(output), "No changed modules found") {
		t.Errorf("expected no-op message, got: %s", output)
	}
}

func TestE2E_ChangedCommand_DetectsUncommitted(t *testing.T) {
	motfBinary := buildMotf(t)
	tmpDir := setupCleanGitRepo(t)

	// Create an uncommitted change in the module
	addUncommittedFile(t, tmpDir, []string{"test-module"}, "outputs.tf", "output \"test\" { value = \"changed\" }\n")

	// Run changed - should detect the uncommitted file
	cmd := exec.Command(motfBinary, "changed", "--ref", "HEAD", "--names")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf changed failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "test-module") {
		t.Errorf("expected to detect uncommitted change in test-module, got: %s", output)
	}
}

func TestE2E_ArgsFlag(t *testing.T) {
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
	tmpDir := t.TempDir()

	// Create a .motf.yml that uses tofu
	configContent := "binary: tofu\nroot: .\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".motf.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create a component module
	createModules(t, tmpDir, []string{"test-component"})

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
	tmpDir := t.TempDir()

	// Create a .motf.yml that uses tofu
	configContent := "binary: tofu\nroot: .\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".motf.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create a component module
	createModules(t, tmpDir, []string{"test-component"})

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

	// Create a component module
	createModules(t, tmpDir, []string{"test-component"})

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

	// Create a component module
	createModules(t, tmpDir, []string{"test-component"})

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

// TestE2E_ParallelFlag tests the --parallel flag with --changed
func TestE2E_ParallelFlag(t *testing.T) {
	motfBinary := buildMotf(t)
	modules := []string{"module-a", "module-b", "module-c"}
	tmpDir := setupGitRepoWithModules(t, modules)

	// Make uncommitted changes to all modules
	addUncommittedFile(t, tmpDir, modules, "variables.tf", "variable \"test_%s\" {\n  type = string\n}\n")

	// Run fmt --changed --parallel
	cmd := exec.Command(motfBinary, "fmt", "--changed", "--ref", "HEAD", "--parallel")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf fmt --changed --parallel failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify all modules were processed (they should appear in output with prefixes)
	for _, mod := range modules {
		if !strings.Contains(outputStr, mod) {
			t.Errorf("expected output to contain module '%s', got: %s", mod, outputStr)
		}
	}
}

// TestE2E_ParallelFlag_MaxParallel tests the --max-parallel flag
func TestE2E_ParallelFlag_MaxParallel(t *testing.T) {
	motfBinary := buildMotf(t)
	modules := []string{"mod-x", "mod-y", "mod-z"}
	tmpDir := setupGitRepoWithModules(t, modules)

	// Make uncommitted changes to all modules
	addUncommittedFile(t, tmpDir, modules, "outputs.tf", "output \"out_%s\" {\n  value = \"test\"\n}\n")

	// Run fmt --changed --parallel --max-parallel=1 (effectively sequential but using parallel infra)
	cmd := exec.Command(motfBinary, "fmt", "--changed", "--ref", "HEAD", "--parallel", "--max-parallel", "1")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("motf fmt --changed --parallel --max-parallel=1 failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify all modules were processed
	for _, mod := range modules {
		if !strings.Contains(outputStr, mod) {
			t.Errorf("expected output to contain module '%s', got: %s", mod, outputStr)
		}
	}
}
