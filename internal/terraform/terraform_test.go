package terraform

import (
	"testing"

	"github.com/TechnicallyJoe/sturdy-parakeet/internal/config"
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

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		extraArgs []string
		expected  []string
	}{
		{
			name:      "command only",
			command:   "init",
			extraArgs: []string{},
			expected:  []string{"init"},
		},
		{
			name:      "command with one arg",
			command:   "init",
			extraArgs: []string{"-upgrade"},
			expected:  []string{"init", "-upgrade"},
		},
		{
			name:      "command with multiple args",
			command:   "init",
			extraArgs: []string{"-upgrade", "-reconfigure"},
			expected:  []string{"init", "-upgrade", "-reconfigure"},
		},
		{
			name:      "fmt with check flag",
			command:   "fmt",
			extraArgs: []string{"-check"},
			expected:  []string{"fmt", "-check"},
		},
		{
			name:      "validate with json output",
			command:   "validate",
			extraArgs: []string{"-json"},
			expected:  []string{"validate", "-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildArgs(tt.command, tt.extraArgs...)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d args, got %d", len(tt.expected), len(result))
			}

			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("arg[%d] = %s, expected %s", i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestBuildArgs_PreservesArgOrder(t *testing.T) {
	args := BuildArgs("init", "-backend=false", "-upgrade", "-reconfigure")

	expected := []string{"init", "-backend=false", "-upgrade", "-reconfigure"}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg order not preserved: got %v, expected %v", args, expected)
			break
		}
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
