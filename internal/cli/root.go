package cli

import (
	"fmt"
	"os"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
	"github.com/TechnicallyJoe/terraform-motf/internal/terraform"
	"github.com/spf13/cobra"
)

// version info set by ldflags at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	cfg    *config.Config
	runner *terraform.Runner

	// Global flags (persistent across all commands)
	pathFlag   string   // Explicit path to module
	argsFlag   []string // Extra arguments passed to terraform/tofu
	configFlag string   // Explicit path to config file

	// Command-specific flags
	// Note: These are registered per-command but share state here for simplicity.
	// Each command that uses these flags registers them in its own init().
	initFlag        bool   // Run init before the command (fmt, validate)
	changedFlag     bool   // Run command against changed modules
	refFlag         string // Ref for change detection (defaults to auto-detect)
	searchFlag      string // Filter pattern for list command
	exampleFlag     string // Target a specific example instead of the module (init, fmt, validate)
	parallelFlag    bool   // Run commands in parallel (init, fmt, validate, test, plan, task)
	maxParallelFlag int    // Maximum parallel jobs to run (default: number of CPU cores)
)

// versionTemplate returns the version string with commit and date.
// It uses ldflags values if set, otherwise falls back to Go build info.
func versionTemplate() string {
	v, c, d := effectiveVersion()
	return fmt.Sprintf("motf version %s\ncommit: %s\nbuilt:  %s\n", v, c, d)
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:     "motf",
	Short:   "Terraform Monorepo Orchestrator (pronounced 'motif')",
	Version: version,
	Long: `motf (Terraform Monorepo Orchestrator) is a CLI tool for working with Terraform monorepos.

It supports running terraform/tofu commands on components, bases, and projects organized
in a structured monorepo.`,
	Example: `  motf fmt storage-account         # Run fmt on storage-account (searches all types)
  motf val k8s-argocd              # Run validate on k8s-argocd
  motf val -i k8s-argocd           # Run init then validate on k8s-argocd
  motf init k8s-argocd             # Run init on k8s-argocd
  motf fmt --path iac/components/azurerm/storage-account  # Run fmt on explicit path
  motf init storage-account -a -upgrade -a -reconfigure  # Run init with extra args`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		cfg, err = config.Load(wd, configFlag)
		if err != nil {
			return err
		}

		// Merge CLI flags into config (CLI takes priority)
		// Centralize the "CLI overrides config" logic here
		if cmd.Flags().Changed("max-parallel") {
			if cfg.Parallelism == nil {
				cfg.Parallelism = &config.ParallelismConfig{}
			}
			cfg.Parallelism.MaxJobs = maxParallelFlag
		}

		// Create terraform runner with config
		runner = terraform.NewRunner(cfg)

		return nil
	},
}

func init() {
	// Set custom version template
	rootCmd.SetVersionTemplate(versionTemplate())

	// Add persistent flags
	rootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "Path to config file (default: searches for .motf.yml)")
	rootCmd.PersistentFlags().StringVar(&pathFlag, "path", "", "Explicit path (mutually exclusive with module name)")
	rootCmd.PersistentFlags().StringArrayVarP(&argsFlag, "args", "a", []string{}, "Extra arguments to pass to terraform/tofu (can be specified multiple times)")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
