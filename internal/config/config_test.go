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

	// Create .motf.yml with Root set
	configContent := `root: /custom/root/path
binary: tofu
`
	configPath := filepath.Join(tmpDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir, "")
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

	// Create .motf.yml without Root set
	configContent := `binary: tofu
`
	configPath := filepath.Join(tmpDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir, "")
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

	cfg, err := Load(tmpDir, "")
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
	//   .motf.yml (with custom root)
	//   subdir/
	//     nested/
	tmpDir := t.TempDir()

	// Create .git directory to mark repo root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create .motf.yml at repo root
	configContent := `binary: tofu
`
	configPath := filepath.Join(tmpDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create nested subdirectory
	nestedDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	// Load from nested directory - should find config at repo root
	cfg, err := Load(nestedDir, "")
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
	//   .motf.yml (should NOT be found)
	//   repo/
	//     .git/
	//     subdir/
	tmpDir := t.TempDir()

	// Create .motf.yml OUTSIDE the git repo (should not be found)
	outsideConfig := `binary: tofu
root: /outside/root
`
	outsideConfigPath := filepath.Join(tmpDir, ".motf.yml")
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
	cfg, err := Load(subDir, "")
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
	//     .motf.yml
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
	configPath := filepath.Join(subDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(subDir, "nested")
	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	// Load from nested directory - should find config in subdir
	cfg, err := Load(nestedDir, "")
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

	// Create .motf.yml with invalid binary
	configContent := `binary: invalid
`
	configPath := filepath.Join(tmpDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	_, err := Load(tmpDir, "")
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
	cfg, err := Load(tmpDir, "")
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

	// Create .motf.yml with test config
	configContent := `binary: terraform
test:
  engine: terraform
  args: -verbose
`
	configPath := filepath.Join(tmpDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir, "")
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

	// Create .motf.yml with partial test config (only args)
	configContent := `binary: terraform
test:
  args: -v -timeout=30m
`
	configPath := filepath.Join(tmpDir, ".motf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpDir, "")
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

func TestLoad_ExplicitConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create config file in a custom location
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.Mkdir(customDir, 0755); err != nil {
		t.Fatalf("failed to create custom directory: %v", err)
	}

	configContent := `binary: tofu
root: ../iac
`
	customConfigPath := filepath.Join(customDir, "my-config.yml")
	if err := os.WriteFile(customConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Load with explicit config path
	cfg, err := Load(tmpDir, customConfigPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Binary != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", cfg.Binary)
	}
	if cfg.ConfigPath != customConfigPath {
		t.Errorf("expected ConfigPath to be '%s', got '%s'", customConfigPath, cfg.ConfigPath)
	}
	// Root should be resolved relative to config file directory
	expectedRoot := filepath.Clean(filepath.Join(customDir, "../iac"))
	if cfg.Root != expectedRoot {
		t.Errorf("expected Root to be '%s', got '%s'", expectedRoot, cfg.Root)
	}
}

func TestLoad_ExplicitConfigPath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load a non-existent config file
	_, err := Load(tmpDir, "/nonexistent/config.yml")
	if err == nil {
		t.Error("expected error for non-existent config file, got nil")
	}
}

func TestLoad_ExplicitConfigPath_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load a directory as config file
	_, err := Load(tmpDir, tmpDir)
	if err == nil {
		t.Error("expected error when config path is a directory, got nil")
	}
}

func TestLoad_ExplicitConfigPath_Symlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create a real config file
	configContent := `binary: tofu
`
	realConfigPath := filepath.Join(tmpDir, "real-config.yml")
	if err := os.WriteFile(realConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create a symlink to the config file
	symlinkPath := filepath.Join(tmpDir, "symlink-config.yml")
	if err := os.Symlink(realConfigPath, symlinkPath); err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	// Loading via symlink should fail (symlinks are not regular files)
	_, err := Load(tmpDir, symlinkPath)
	if err == nil {
		t.Error("expected error when config path is a symlink, got nil")
	}
}

func TestLoad_ExplicitConfigPath_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create config file in tmpDir
	configContent := `binary: tofu
`
	configPath := filepath.Join(tmpDir, "test-config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Change to tmpDir and load with relative path
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cfg, err := Load(tmpDir, "test-config.yml")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Binary != "tofu" {
		t.Errorf("expected Binary to be 'tofu', got '%s'", cfg.Binary)
	}
	// ConfigPath should be absolute
	if !filepath.IsAbs(cfg.ConfigPath) {
		t.Errorf("expected ConfigPath to be absolute, got '%s'", cfg.ConfigPath)
	}
}
