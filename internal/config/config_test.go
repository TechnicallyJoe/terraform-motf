package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_WithConfigFileAndRootSet(t *testing.T) {
	// Create a temp directory structure with a git repo and config file
	tmpDir := t.TempDir()

	// Create .git directory to mark repo root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .tfpl.yml with Root set
	configContent := `root: /custom/root/path
binary: tofu
`
	configPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Root != "/custom/root/path" {
		t.Errorf("expected Root to be '/custom/root/path', got '%s'", cfg.Root)
	}
	if cfg.Binary != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", cfg.Binary)
	}
}

func TestLoad_WithConfigFileWithoutRoot(t *testing.T) {
	// Create a temp directory structure with a git repo and config file
	tmpDir := t.TempDir()

	// Create .git directory to mark repo root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .tfpl.yml without Root set
	configContent := `binary: tofu
`
	configPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Root should default to git root
	if cfg.Root != tmpDir {
		t.Errorf("expected Root to be '%s' (git root), got '%s'", tmpDir, cfg.Root)
	}
	if cfg.Binary != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", cfg.Binary)
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Create a temp directory structure with only a git repo
	tmpDir := t.TempDir()

	// Create .git directory to mark repo root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Root should default to git root
	if cfg.Root != tmpDir {
		t.Errorf("expected Root to be '%s' (git root), got '%s'", tmpDir, cfg.Root)
	}
	// Binary should be default
	if cfg.Binary != "terraform" {
		t.Errorf("expected Binary to be 'terraform', got '%s'", cfg.Binary)
	}
}

func TestLoad_StopsAtGitRoot(t *testing.T) {
	// Create a temp directory structure:
	// tmpDir/
	//   .git/
	//   .tfpl.yml (with custom root)
	//   subdir/
	//     nested/
	tmpDir := t.TempDir()

	// Create .git directory to mark repo root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .tfpl.yml at repo root
	configContent := `binary: tofu
`
	configPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create nested subdirectory
	nestedDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	// Load from nested directory - should find config at repo root
	cfg, err := Load(nestedDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Binary != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", cfg.Binary)
	}
	if cfg.Root != tmpDir {
		t.Errorf("expected Root to be '%s' (git root), got '%s'", tmpDir, cfg.Root)
	}
}

func TestLoad_DoesNotSearchBeyondGitRoot(t *testing.T) {
	// Create a temp directory structure:
	// tmpDir/
	//   .tfpl.yml (should NOT be found)
	//   repo/
	//     .git/
	//     subdir/
	tmpDir := t.TempDir()

	// Create .tfpl.yml OUTSIDE the git repo (should not be found)
	outsideConfig := `binary: tofu
root: /outside/root
`
	outsideConfigPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(outsideConfigPath, []byte(outsideConfig), 0644); err != nil {
		t.Fatalf("failed to create outside config file: %v", err)
	}

	// Create repo directory with .git
	repoDir := filepath.Join(tmpDir, "repo")
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create subdirectory inside repo
	subDir := filepath.Join(repoDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Load from subdir - should NOT find the config outside the git repo
	cfg, err := Load(subDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Should get defaults, not the outside config
	if cfg.Binary != "terraform" {
		t.Errorf("expected Binary to be 'terraform' (default), got '%s'", cfg.Binary)
	}
	if cfg.Root != repoDir {
		t.Errorf("expected Root to be '%s' (git root), got '%s'", repoDir, cfg.Root)
	}
}

func TestLoad_ConfigInSubdirectory(t *testing.T) {
	// Create a temp directory structure:
	// tmpDir/
	//   .git/
	//   subdir/
	//     .tfpl.yml
	//     nested/
	tmpDir := t.TempDir()

	// Create .git directory to mark repo root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create subdirectory with config
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	configContent := `binary: tofu
`
	configPath := filepath.Join(subDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(subDir, "nested")
	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	// Load from nested directory - should find config in subdir
	cfg, err := Load(nestedDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Binary != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", cfg.Binary)
	}
	// Root should still default to git root since not specified in config
	if cfg.Root != tmpDir {
		t.Errorf("expected Root to be '%s' (git root), got '%s'", tmpDir, cfg.Root)
	}
}

func TestIsGitRoot(t *testing.T) {
	t.Run("with .git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("failed to create .git directory: %v", err)
		}

		if !isGitRoot(tmpDir) {
			t.Error("expected isGitRoot to return true for directory with .git folder")
		}
	})

	t.Run("with .git file (worktree)", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitFile := filepath.Join(tmpDir, ".git")
		if err := os.WriteFile(gitFile, []byte("gitdir: /some/path"), 0644); err != nil {
			t.Fatalf("failed to create .git file: %v", err)
		}

		if !isGitRoot(tmpDir) {
			t.Error("expected isGitRoot to return true for directory with .git file")
		}
	})

	t.Run("without .git", func(t *testing.T) {
		tmpDir := t.TempDir()

		if isGitRoot(tmpDir) {
			t.Error("expected isGitRoot to return false for directory without .git")
		}
	})
}

func TestFindGitRoot(t *testing.T) {
	t.Run("finds git root from nested directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("failed to create .git directory: %v", err)
		}

		nestedDir := filepath.Join(tmpDir, "a", "b", "c")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("failed to create nested directory: %v", err)
		}

		root := findGitRoot(nestedDir)
		if root != tmpDir {
			t.Errorf("expected git root to be '%s', got '%s'", tmpDir, root)
		}
	})

	t.Run("returns empty string when no git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "a", "b")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("failed to create nested directory: %v", err)
		}

		root := findGitRoot(nestedDir)
		if root != "" {
			t.Errorf("expected empty string when no git repo, got '%s'", root)
		}
	})
}

func TestLoad_InvalidBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .tfpl.yml with invalid binary
	configContent := `binary: invalid
`
	configPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	_, err := Load(tmpDir)
	if err == nil {
		t.Error("expected error for invalid binary, got nil")
	}
}

func TestLoad_TestConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Load without config file - should get defaults
	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Test == nil {
		t.Fatal("expected Test config to be initialized")
	}
	if cfg.Test.Engine != "terratest" {
		t.Errorf("expected Test.Engine to be 'terratest', got '%s'", cfg.Test.Engine)
	}
	if cfg.Test.Args != "" {
		t.Errorf("expected Test.Args to be empty, got '%s'", cfg.Test.Args)
	}
}

func TestLoad_TestConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .tfpl.yml with test config
	configContent := `binary: terraform
test:
  engine: terraform
  args: -verbose
`
	configPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Test == nil {
		t.Fatal("expected Test config to be initialized")
	}
	if cfg.Test.Engine != "terraform" {
		t.Errorf("expected Test.Engine to be 'terraform', got '%s'", cfg.Test.Engine)
	}
	if cfg.Test.Args != "-verbose" {
		t.Errorf("expected Test.Args to be '-verbose', got '%s'", cfg.Test.Args)
	}
}

func TestLoad_TestConfigPartialFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .tfpl.yml with partial test config (only args)
	configContent := `binary: terraform
test:
  args: -v -timeout=30m
`
	configPath := filepath.Join(tmpDir, ".tfpl.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Test == nil {
		t.Fatal("expected Test config to be initialized")
	}
	// Engine should default to terratest even if not specified
	if cfg.Test.Engine != "terratest" {
		t.Errorf("expected Test.Engine to default to 'terratest', got '%s'", cfg.Test.Engine)
	}
	if cfg.Test.Args != "-v -timeout=30m" {
		t.Errorf("expected Test.Args to be '-v -timeout=30m', got '%s'", cfg.Test.Args)
	}
}
