package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TechnicallyJoe/terraform-motf/internal/tasks"
	"gopkg.in/yaml.v3"
)

// TestConfig represents the test configuration section
type TestConfig struct {
	Engine string `yaml:"engine"`
	Args   string `yaml:"args"`
}

// Config represents the .motf.yml configuration file
type Config struct {
	Root       string                       `yaml:"root"`
	Binary     string                       `yaml:"binary"`
	Test       *TestConfig                  `yaml:"test"`
	Tasks      map[string]*tasks.TaskConfig `yaml:"tasks"`
	ConfigPath string                       `yaml:"-"` // Path to the config file, if found
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Root:   "",
		Binary: "terraform",
		Test: &TestConfig{
			Engine: "terratest",
			Args:   "",
		},
	}
}

// isGitRoot checks if the given directory is the root of a Git repository
func isGitRoot(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	// .git can be a directory (regular repo) or a file (worktree/submodule)
	return info.IsDir() || info.Mode().IsRegular()
}

// findGitRoot finds the root of the Git repository starting from startDir
func findGitRoot(startDir string) string {
	dir := startDir
	for {
		if isGitRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, no git repo found
			return ""
		}
		dir = parent
	}
}

// Load searches for .motf.yml starting from startDir and walking up the directory tree
// until it reaches the Git repository root. If configPath is provided, it loads that file directly.
func Load(startDir string, configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Find the git root first - this will be our default Root value
	gitRoot := findGitRoot(startDir)

	// If explicit config path provided, use it directly
	if configPath != "" {
		return loadConfigFile(cfg, configPath, gitRoot)
	}

	// Walk up the directory tree looking for .motf.yml
	dir := startDir
	for {
		configPath := filepath.Join(dir, ".motf.yml")
		if _, err := os.Stat(configPath); err == nil {
			// Found config file
			data, err := os.ReadFile(configPath) //nolint:gosec // configPath is constructed from known directory traversal
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}

			// Validate binary
			if cfg.Binary != "terraform" && cfg.Binary != "tofu" {
				return nil, fmt.Errorf("invalid binary '%s' in config: must be 'terraform' or 'tofu'", cfg.Binary)
			}

			// Ensure Test config has defaults if not set
			if cfg.Test == nil {
				cfg.Test = &TestConfig{
					Engine: "terratest",
					Args:   "",
				}
			} else {
				if cfg.Test.Engine == "" {
					cfg.Test.Engine = "terratest"
				}
			}

			// Store the config file path
			cfg.ConfigPath = configPath

			// If Root is not set in config, default to git root
			// If Root is set and is relative, resolve it relative to the config file directory
			if cfg.Root == "" {
				cfg.Root = gitRoot
			} else if !filepath.IsAbs(cfg.Root) {
				cfg.Root = filepath.Join(dir, cfg.Root)
			}

			return cfg, nil
		}

		// Stop if we've reached the Git repository root
		if isGitRoot(dir) {
			break
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, no config file found
			break
		}
		dir = parent
	}

	// No config file found, set Root to git root and return defaults
	cfg.Root = gitRoot
	return cfg, nil
}

// loadConfigFile loads and validates a config file at the given path.
// The configPath is validated to ensure it's a regular file before reading.
func loadConfigFile(cfg *Config, configPath string, gitRoot string) (*Config, error) {
	// Clean and validate the path
	cleanPath := filepath.Clean(configPath)

	// Resolve to absolute path if relative
	if !filepath.IsAbs(cleanPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		cleanPath = filepath.Join(wd, cleanPath)
	}

	// Verify the file exists and is a regular file.
	// Symlinks are intentionally rejected here to:
	//   - Avoid config files resolving outside the expected repo/git root via symlink chains.
	//   - Keep path resolution predictable for features that derive other paths from the config location.
	//   - Reduce the risk of TOCTOU-style issues where a symlink target changes between validation and use.
	info, err := os.Lstat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access config file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("config path is not a regular file: %s", cleanPath)
	}

	data, err := os.ReadFile(cleanPath) //nolint:gosec // path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate binary
	if cfg.Binary != "terraform" && cfg.Binary != "tofu" {
		return nil, fmt.Errorf("invalid binary '%s' in config: must be 'terraform' or 'tofu'", cfg.Binary)
	}

	// Ensure Test config has defaults if not set
	if cfg.Test == nil {
		cfg.Test = &TestConfig{Engine: "terratest", Args: ""}
	} else if cfg.Test.Engine == "" {
		cfg.Test.Engine = "terratest"
	}

	cfg.ConfigPath = cleanPath

	// Resolve Root relative to config file directory
	dir := filepath.Dir(cleanPath)
	if cfg.Root == "" {
		cfg.Root = gitRoot
	} else if !filepath.IsAbs(cfg.Root) {
		cfg.Root = filepath.Join(dir, cfg.Root)
	}

	return cfg, nil
}
