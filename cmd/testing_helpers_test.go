package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

// resetFlags resets all package-level flags to their default values.
// Call this in t.Cleanup() to ensure flags are reset after each test.
func resetFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		pathFlag = ""
		argsFlag = []string{}
		initFlag = false
		searchFlag = ""
		exampleFlag = ""
	})
}

// withConfig sets the global cfg for the duration of the test.
// It will be reset to nil after the test completes.
func withConfig(t *testing.T, c *config.Config) {
	t.Helper()
	cfg = c
	t.Cleanup(func() {
		cfg = nil
	})
}

// withWorkingDir temporarily changes the working directory for the test.
// It will be restored after the test completes.
func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", dir, err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})
}

// createTerraformModule creates a minimal terraform module in the given directory.
// Returns the full path to the created module.
func createTerraformModule(t *testing.T, baseDir, relativePath string) string {
	t.Helper()
	modulePath := filepath.Join(baseDir, relativePath)
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	tfFile := filepath.Join(modulePath, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# terraform"), 0644); err != nil {
		t.Fatalf("failed to create .tf file: %v", err)
	}

	return modulePath
}
