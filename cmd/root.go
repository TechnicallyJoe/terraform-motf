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
	cfg *config.Config

	// Flags
	componentFlag string
	baseFlag      string
	projectFlag   string
	pathFlag      string
	initFlag      bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:     "tfpl",
	Short:   "Terraform Polylith CLI tool",
	Version: version,
	Long: `tfpl (Terraform Polylith) is a CLI tool for working with polylith-style Terraform repositories.

It supports running terraform/tofu commands on components, bases, and projects organized 
in a polylith structure.`,
	Example: `  tfpl fmt -c storage-account      # Run fmt on component storage-account
  tfpl val -b k8s-argocd           # Run validate on base k8s-argocd
  tfpl val -i -b k8s-argocd        # Run init then validate on base k8s-argocd
  tfpl init -b k8s-argocd          # Run init on base k8s-argocd
  tfpl fmt --path iac/components/azurerm/storage-account  # Run fmt on explicit path`,
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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run terraform/tofu init on a component, base, or project",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath()
		if err != nil {
			return err
		}

		return terraform.RunInit(targetPath)
	},
}

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Run terraform/tofu fmt on a component, base, or project",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath()
		if err != nil {
			return err
		}

		// Run init first if flag is set
		if initFlag {
			if err := terraform.RunInit(targetPath); err != nil {
				return err
			}
		}

		return terraform.RunFmt(targetPath)
	},
}

// valCmd represents the validate command
var valCmd = &cobra.Command{
	Use:     "val",
	Aliases: []string{"validate"},
	Short:   "Run terraform/tofu validate on a component, base, or project",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath, err := resolveTargetPath()
		if err != nil {
			return err
		}

		// Run init first if flag is set
		if initFlag {
			if err := terraform.RunInit(targetPath); err != nil {
				return err
			}
		}

		return terraform.RunValidate(targetPath)
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
		return "", fmt.Errorf("must specify one of: --component/-c, --base/-b, --project/-p, or --path")
	}

	if flagsSet > 1 {
		if pathFlag != "" {
			return "", fmt.Errorf("--path is mutually exclusive with --component/-c, --base/-b, and --project/-p")
		}
		return "", fmt.Errorf("only one of --component/-c, --base/-b, or --project/-p can be specified at a time")
	}

	// If explicit path is provided, use it directly
	if pathFlag != "" {
		return resolveExplicitPath(pathFlag)
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

	// Find the module
	return findModule(moduleType, moduleName)
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

// findModule finds a module by type and name
func findModule(moduleType, moduleName string) (string, error) {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Determine search path
	searchPath := wd
	if cfg.Root != "" {
		searchPath = filepath.Join(wd, cfg.Root)
	}
	searchPath = filepath.Join(searchPath, moduleType)

	// Check if search path exists
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return "", fmt.Errorf("%s directory does not exist: %s", moduleType, searchPath)
	}

	// Find the module
	matches, err := finder.FindModule(searchPath, moduleName)
	if err != nil {
		return "", fmt.Errorf("failed to search for module: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("%s '%s' not found in %s", moduleType[:len(moduleType)-1], moduleName, searchPath)
	}

	if len(matches) > 1 {
		// Name clash detected
		moduleTypeSingular := moduleType[:len(moduleType)-1]
		fmt.Fprintf(os.Stderr, "Error: multiple %ss named '%s' found - name clash detected:\n", moduleTypeSingular, moduleName)
		for i, match := range matches {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, match)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Please use --path to specify the exact path")
		return "", fmt.Errorf("name clash detected")
	}

	return matches[0], nil
}
