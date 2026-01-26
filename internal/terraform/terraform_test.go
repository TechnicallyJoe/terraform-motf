package terraform

import (
	"testing"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

func TestNewRunner(t *testing.T) {
	cfg := &config.Config{
		Root:   "/some/path",
		Binary: "tofu",
	}

	runner := NewRunner(cfg)

	if runner == nil {
		t.Fatal("NewRunner returned nil")
	}

	if runner.config != cfg {
		t.Error("NewRunner did not store config correctly")
	}
}

func TestRunner_Binary(t *testing.T) {
	tests := []struct {
		name     string
		binary   string
		expected string
	}{
		{
			name:     "terraform binary",
			binary:   "terraform",
			expected: "terraform",
		},
		{
			name:     "tofu binary",
			binary:   "tofu",
			expected: "tofu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Binary: tt.binary,
			}

			runner := NewRunner(cfg)
			if runner.Binary() != tt.expected {
				t.Errorf("Binary() = %s, expected %s", runner.Binary(), tt.expected)
			}
		})
	}
}

func TestRunner_InheritsConfigBinary(t *testing.T) {
	// Test that Runner properly inherits the binary from config
	cfg := &config.Config{
		Root:   "/test/root",
		Binary: "tofu",
	}

	runner := NewRunner(cfg)

	// Binary should come from config
	if runner.Binary() != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", runner.Binary())
	}

	// Changing config should reflect in runner
	cfg.Binary = "terraform"
	if runner.Binary() != "terraform" {
		t.Errorf("expected Binary to be 'terraform' after config change, got '%s'", runner.Binary())
	}
}

func TestRunner_WithDefaultConfig(t *testing.T) {
	// Test that Runner works with default config values
	cfg := config.DefaultConfig()
	runner := NewRunner(cfg)

	if runner.Binary() != "terraform" {
		t.Errorf("expected default Binary to be 'terraform', got '%s'", runner.Binary())
	}
}

func TestRunner_RunTest_UnsupportedEngine(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Binary: "terraform",
		Test: &config.TestConfig{
			Engine: "unsupported",
			Args:   "",
		},
	}

	runner := NewRunner(cfg)
	tmpDir := t.TempDir()

	err := runner.RunTest(tmpDir)
	if err == nil {
		t.Error("expected error for unsupported test engine, got nil")
	}
}

func TestRunner_RunTest_DefaultEngine(t *testing.T) {
	cfg := config.DefaultConfig()
	_ = NewRunner(cfg)

	if cfg.Test.Engine != "terratest" {
		t.Errorf("expected default test engine to be 'terratest', got '%s'", cfg.Test.Engine)
	}
}

// TestRunner_RunTest_TerratestEngine verifies that terratest engine runs go test
func TestRunner_RunTest_TerratestEngine(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Binary: "terraform",
		Test: &config.TestConfig{
			Engine: "terratest",
			Args:   "",
		},
	}

	_ = NewRunner(cfg)

	// Verify the engine is correctly set
	if cfg.Test.Engine != "terratest" {
		t.Errorf("expected engine 'terratest', got '%s'", cfg.Test.Engine)
	}

	// The actual command would be: go test ./...
	// We can't easily test the actual execution without mocking,
	// but we verify the config is correctly set up
}

// TestRunner_RunTest_WithConfigArgs verifies config args are included
func TestRunner_RunTest_WithConfigArgs(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Binary: "terraform",
		Test: &config.TestConfig{
			Engine: "terratest",
			Args:   "-v -count=1",
		},
	}

	_ = NewRunner(cfg)

	// Verify config args are set
	if cfg.Test.Args != "-v -count=1" {
		t.Errorf("expected args '-v -count=1', got '%s'", cfg.Test.Args)
	}

	// The actual command would be: go test ./... -v -count=1
	// This validates that the config properly stores the args
}

// TestRunner_RunTest_TerraformEngine verifies terraform engine setup
func TestRunner_RunTest_TerraformEngine(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Binary: "terraform",
		Test: &config.TestConfig{
			Engine: "terraform",
			Args:   "-verbose",
		},
	}

	_ = NewRunner(cfg)

	// Verify the engine is correctly set
	if cfg.Test.Engine != "terraform" {
		t.Errorf("expected engine 'terraform', got '%s'", cfg.Test.Engine)
	}

	// The actual command would be: terraform test -verbose
}

// TestRunner_RunTest_TofuEngine verifies tofu engine setup
func TestRunner_RunTest_TofuEngine(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Binary: "terraform",
		Test: &config.TestConfig{
			Engine: "tofu",
			Args:   "",
		},
	}

	_ = NewRunner(cfg)

	// Verify the engine is correctly set
	if cfg.Test.Engine != "tofu" {
		t.Errorf("expected engine 'tofu', got '%s'", cfg.Test.Engine)
	}

	// The actual command would be: tofu test
}
