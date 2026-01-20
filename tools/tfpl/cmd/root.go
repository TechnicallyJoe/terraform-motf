package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechnicallyJoe/sturdy-parakeet/tools/tfpl/internal/config"
	"github.com/TechnicallyJoe/sturdy-parakeet/tools/tfpl/internal/finder"
	"github.com/TechnicallyJoe/sturdy-parakeet/tools/tfpl/internal/terraform"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var (
	componentFlag string
	baseFlag      string
	projectFlag   string
	pathFlag      string
	initFlag      bool
	cfg           config.Config
)

var rootCmd = &cobra.Command{
	Use:     "tfpl",
	Short:   "Terraform Polylith - CLI tool for polylith-style Terraform repositories",
	Long:    `tfpl is a CLI tool that supports polylith-style Terraform repositories and makes it easy to work with them.`,
	Version: version,
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

		// Set the terraform binary
		terraform.SetBinary(cfg.Binary)

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run terraform/tofu init on a component, base, or project",
	Long:  `Run terraform/tofu init on a component, base, or project.`,
	Example: `  tfpl init -c storage-account
  tfpl init -b k8s-argocd
  tfpl init -p spacelift-modules
  tfpl init --path iac/components/azurerm/storage-account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath()
		if err != nil {
			return err
		}

		return terraform.RunInit(targetPath)
	},
}

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Run terraform/tofu fmt on a component, base, or project",
	Long:  `Run terraform/tofu fmt on a component, base, or project.`,
	Example: `  tfpl fmt -c storage-account
  tfpl fmt -b k8s-argocd
  tfpl fmt -i -b k8s-argocd
  tfpl fmt --path iac/components/azurerm/storage-account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath()
		if err != nil {
			return err
		}

		// Run init first if requested
		if initFlag {
			if err := terraform.RunInit(targetPath); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}
		}

		return terraform.RunFmt(targetPath)
	},
}

var valCmd = &cobra.Command{
	Use:     "val",
	Aliases: []string{"validate"},
	Short:   "Run terraform/tofu validate on a component, base, or project",
	Long:    `Run terraform/tofu validate on a component, base, or project.`,
	Example: `  tfpl val -c storage-account
  tfpl val -b k8s-argocd
  tfpl val -i -b k8s-argocd
  tfpl validate -p spacelift-modules
  tfpl val --path iac/components/azurerm/storage-account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath()
		if err != nil {
			return err
		}

		// Run init first if requested
		if initFlag {
			if err := terraform.RunInit(targetPath); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}
		}

		return terraform.RunValidate(targetPath)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long:  `Display the current tfpl configuration loaded from .tfpl.yml or defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Current configuration:")
		fmt.Printf("  Root:   %s\n", cfg.Root)
		if cfg.Root == "" {
			fmt.Println("          (repository root)")
		}
		fmt.Printf("  Binary: %s\n", cfg.Binary)
		return nil
	},
}

func init() {
	// Add persistent flags
	rootCmd.PersistentFlags().StringVarP(&componentFlag, "component", "c", "", "Component name to operate on")
	rootCmd.PersistentFlags().StringVarP(&baseFlag, "base", "b", "", "Base name to operate on")
	rootCmd.PersistentFlags().StringVarP(&projectFlag, "project", "p", "", "Project name to operate on")
	rootCmd.PersistentFlags().StringVar(&pathFlag, "path", "", "Explicit path (mutually exclusive with -c, -b, -p)")

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

// resolveTargetPath resolves the target path based on flags
func resolveTargetPath() (string, error) {
	// Validate mutual exclusivity
	flagsSet := 0
	if componentFlag != "" {
		flagsSet++
	}
	if baseFlag != "" {
		flagsSet++
	}
	if projectFlag != "" {
		flagsSet++
	}
	if pathFlag != "" {
		flagsSet++
	}

	if flagsSet == 0 {
		return "", fmt.Errorf("must specify one of: --component, --base, --project, or --path")
	}

	if flagsSet > 1 {
		if pathFlag != "" {
			return "", fmt.Errorf("--path is mutually exclusive with --component, --base, and --project")
		}
		return "", fmt.Errorf("only one of --component, --base, or --project can be specified at a time")
	}

	// If explicit path is provided, use it
	if pathFlag != "" {
		absPath, err := filepath.Abs(pathFlag)
		if err != nil {
			return "", fmt.Errorf("failed to resolve path: %w", err)
		}

		// Check if path exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", absPath)
		}

		return absPath, nil
	}

	// Determine the search base directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	searchBase := wd
	if cfg.Root != "" {
		searchBase = filepath.Join(wd, cfg.Root)
	}

	// Determine module type and name
	var moduleType, moduleName string
	if componentFlag != "" {
		moduleType = "components"
		moduleName = componentFlag
	} else if baseFlag != "" {
		moduleType = "bases"
		moduleName = baseFlag
	} else if projectFlag != "" {
		moduleType = "projects"
		moduleName = projectFlag
	}

	// Search for the module
	searchPath := filepath.Join(searchBase, moduleType)
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return "", fmt.Errorf("%s directory does not exist: %s", moduleType, searchPath)
	}

	matches, err := finder.FindModule(searchPath, moduleName)
	if err != nil {
		return "", fmt.Errorf("failed to search for module: %w", err)
	}

	if len(matches) == 0 {
		singularType := strings.TrimSuffix(moduleType, "s")
		return "", fmt.Errorf("no %s named '%s' found in %s", singularType, moduleName, searchPath)
	}

	if len(matches) > 1 {
		singularType := strings.TrimSuffix(moduleType, "s")
		fmt.Fprintf(os.Stderr, "Error: multiple %s named '%s' found - name clash detected:\n", singularType, moduleName)
		for i, match := range matches {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, match)
		}
		fmt.Fprintln(os.Stderr)
		return "", fmt.Errorf("please use --path to specify the exact path")
	}

	return matches[0], nil
}
