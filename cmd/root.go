package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TechnicallyJoe/sturdy-parakeet/internal/config"
	"github.com/TechnicallyJoe/sturdy-parakeet/internal/finder"
	"github.com/TechnicallyJoe/sturdy-parakeet/internal/terraform"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var (
	cfg    *config.Config
	runner *terraform.Runner

	// Flags
	pathFlag string
	initFlag bool
	argsFlag []string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:     "tfpl",
	Short:   "Terraform Polylith CLI tool",
	Version: version,
	Long: `tfpl (Terraform Polylith) is a CLI tool for working with polylith-style Terraform repositories.

It supports running terraform/tofu commands on components, bases, and projects organized
in a polylith structure.`,
	Example: `  tfpl fmt storage-account         # Run fmt on storage-account (searches all types)
  tfpl val k8s-argocd              # Run validate on k8s-argocd
  tfpl val -i k8s-argocd           # Run init then validate on k8s-argocd
  tfpl init k8s-argocd             # Run init on k8s-argocd
  tfpl fmt --path iac/components/azurerm/storage-account  # Run fmt on explicit path
  tfpl init storage-account -a -upgrade -a -reconfigure  # Run init with extra args`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		cfg, err = config.Load(wd)
		if err != nil {
			return err
		}

		// Create terraform runner with config
		runner = terraform.NewRunner(cfg)

		return nil
	},
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [module-name]",
	Short: "Run terraform/tofu init on a component, base, or project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath(args)
		if err != nil {
			return err
		}

		return runner.RunInit(targetPath, argsFlag...)
	},
}

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt [module-name]",
	Short: "Run terraform/tofu fmt on a component, base, or project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath(args)
		if err != nil {
			return err
		}

		// Run init first if flag is set
		if initFlag {
			if err := runner.RunInit(targetPath); err != nil {
				return err
			}
		}

		return runner.RunFmt(targetPath, argsFlag...)
	},
}

// valCmd represents the validate command
var valCmd = &cobra.Command{
	Use:     "val [module-name]",
	Aliases: []string{"validate"},
	Short:   "Run terraform/tofu validate on a component, base, or project",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath(args)
		if err != nil {
			return err
		}

		// Run init first if flag is set
		if initFlag {
			if err := runner.RunInit(targetPath); err != nil {
				return err
			}
		}

		return runner.RunValidate(targetPath, argsFlag...)
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Current configuration:")
		fmt.Printf("  Root:   %s\n", cfg.Root)
		fmt.Printf("  Binary: %s\n", cfg.Binary)
		if cfg.ConfigPath != "" {
			fmt.Printf("  Config: %s\n", cfg.ConfigPath)
		} else {
			fmt.Printf("  Config: (using defaults)\n")
		}
	},
}

func init() {
	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&pathFlag, "path", "", "Explicit path (mutually exclusive with module name)")
	rootCmd.PersistentFlags().StringArrayVarP(&argsFlag, "args", "a", []string{}, "Extra arguments to pass to terraform/tofu (can be specified multiple times)")

	// Add init flag for fmt and val commands
	fmtCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Run init before the command")
	valCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Run init before the command")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(fmtCmd)
	rootCmd.AddCommand(valCmd)
	rootCmd.AddCommand(configCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// resolveTargetPath resolves the target path based on args and flags
func resolveTargetPath(args []string) (string, error) {
	// Check if both module name and --path are specified
	if len(args) > 0 && pathFlag != "" {
		return "", fmt.Errorf("--path is mutually exclusive with module name argument")
	}

	// Check if neither module name nor --path is specified
	if len(args) == 0 && pathFlag == "" {
		return "", fmt.Errorf("must specify either a module name or --path")
	}

	// If explicit path is provided, use it directly
	if pathFlag != "" {
		return resolveExplicitPath(pathFlag)
	}

	// Use the module name from args
	moduleName := args[0]

	// Search for the module in all directories
	return findModuleInAllDirs(moduleName)
}

// resolveExplicitPath resolves an explicit path (can be relative or absolute)
func resolveExplicitPath(path string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	return absPath, nil
}

// findModuleInAllDirs searches for a module across all three directories (components, bases, projects)
func findModuleInAllDirs(moduleName string) (string, error) {
	// Determine base search path based on cfg.Root
	var basePath string
	if cfg.Root != "" {
		// cfg.Root can be an absolute path (git root) or relative path
		if filepath.IsAbs(cfg.Root) {
			basePath = cfg.Root
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("failed to get working directory: %w", err)
			}
			basePath = filepath.Join(wd, cfg.Root)
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		basePath = wd
	}

	// Search in all three directories
	moduleTypes := []string{"components", "bases", "projects"}
	var allMatches []string
	
	for _, moduleType := range moduleTypes {
		searchPath := filepath.Join(basePath, moduleType)
		
		// Skip if directory doesn't exist
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}
		
		// Find the module
		matches, err := finder.FindModule(searchPath, moduleName)
		if err != nil {
			return "", fmt.Errorf("failed to search for module in %s: %w", moduleType, err)
		}
		
		allMatches = append(allMatches, matches...)
	}

	if len(allMatches) == 0 {
		return "", fmt.Errorf("module '%s' not found in components, bases, or projects", moduleName)
	}

	if len(allMatches) > 1 {
		// Name clash detected across multiple directories
		fmt.Fprintf(os.Stderr, "Error: multiple modules named '%s' found - name clash detected:\n", moduleName)
		for i, match := range allMatches {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, match)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Please use --path to specify the exact path")
		return "", fmt.Errorf("name clash detected")
	}

	return allMatches[0], nil
}
