package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/TechnicallyJoe/terraform-motf/internal/finder"
	"github.com/TechnicallyJoe/terraform-motf/internal/spacelift"
	"github.com/spf13/cobra"
)

// listJsonFlag controls JSON output for list command
var listJsonFlag bool

// listNamesOnlyFlag outputs only module names (not paths)
var listNamesOnlyFlag bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all modules (components, bases, and projects)",
	Long: `List all modules found in components, bases, and projects directories.

Use the --search/-s flag to filter modules using wildcards.
Use the --changed flag to show only modules with changes compared to a git ref.
Use the --json flag to output in JSON format for scripting.

Examples:
  motf list                        # List all modules
  motf list -s storage             # List modules containing "storage"
  motf list -s *account*           # List modules with "account" anywhere in the name
  motf list --json                 # Output as JSON
  motf list --changed              # List only changed modules
  motf list --changed --ref HEAD~5 # List modules changed in last 5 commits
  motf list --changed --names      # Output only changed module names (for scripting)
  motf list --changed -s storage   # List changed modules matching "storage"`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&searchFlag, "search", "s", "", "Filter modules using wildcards (e.g., *storage*)")
	listCmd.Flags().BoolVar(&listJsonFlag, "json", false, "Output in JSON format")
	listCmd.Flags().BoolVar(&listNamesOnlyFlag, "names", false, "Output only module names (one per line)")
	listCmd.Flags().BoolVar(&changedFlag, "changed", false, "List only modules changed compared to --ref")
	listCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	var modules []ModuleInfo

	if changedFlag {
		// Get changed modules
		modules, err = detectChangedModules(refFlag)
		if err != nil {
			return err
		}

		// Apply search filter if specified
		if searchFlag != "" && len(modules) > 0 {
			var filtered []ModuleInfo
			for _, mod := range modules {
				if finder.MatchesWildcard(mod.Name, searchFlag) {
					filtered = append(filtered, mod)
				}
			}
			modules = filtered
		}

		// Populate version info for display
		for i := range modules {
			absPath := filepath.Join(basePath, modules[i].Path)
			modules[i].Version = spacelift.ReadModuleVersion(absPath)
		}
	} else {
		modules, err = collectModules(basePath, searchFlag)
		if err != nil {
			return err
		}
	}

	if len(modules) == 0 {
		if listJsonFlag {
			fmt.Println("[]")
			return nil
		}
		if listNamesOnlyFlag {
			return nil
		}
		if changedFlag {
			fmt.Println("No changed modules found")
		} else if searchFlag != "" {
			fmt.Printf("No modules found matching '%s'\n", searchFlag)
		} else {
			fmt.Println("No modules found")
		}
		return nil
	}

	sortModules(modules)

	if listJsonFlag {
		return printModulesJSON(modules)
	}

	if listNamesOnlyFlag {
		for _, mod := range modules {
			fmt.Println(mod.Name)
		}
		return nil
	}

	printModules(modules)
	return nil
}

// collectModules discovers all modules across components, bases, and projects directories
func collectModules(basePath, searchFilter string) ([]ModuleInfo, error) {
	var allModules []ModuleInfo

	for _, moduleDir := range ModuleDirs {
		searchPath := filepath.Join(basePath, moduleDir)

		// Skip if directory doesn't exist
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}

		// List all modules in this directory
		modules, err := finder.ListAllModules(searchPath)
		if err != nil {
			return nil, fmt.Errorf("failed to list modules in %s: %w", moduleDir, err)
		}

		// Process each module
		for name, path := range modules {
			// Apply search filter if specified
			if searchFilter != "" && !finder.MatchesWildcard(name, searchFilter) {
				continue
			}

			// Make path relative to basePath
			relativePath, err := filepath.Rel(basePath, path)
			if err != nil {
				relativePath = path // Fallback to full path if relative fails
			}

			allModules = append(allModules, ModuleInfo{
				Name:    name,
				Type:    getModuleType(path),
				Path:    relativePath,
				Version: spacelift.ReadModuleVersion(path),
			})
		}
	}

	return allModules, nil
}

// sortModules sorts modules alphabetically by path
func sortModules(modules []ModuleInfo) {
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})
}

// printModules outputs the list of modules to stdout in table format
func printModules(modules []ModuleInfo) {
	// Calculate column widths
	nameWidth := len("NAME")
	typeWidth := len("TYPE")
	pathWidth := len("PATH")
	versionWidth := len("VERSION")

	for _, mod := range modules {
		if len(mod.Name) > nameWidth {
			nameWidth = len(mod.Name)
		}
		if len(mod.Type) > typeWidth {
			typeWidth = len(mod.Type)
		}
		if len(mod.Path) > pathWidth {
			pathWidth = len(mod.Path)
		}
		if len(mod.Version) > versionWidth {
			versionWidth = len(mod.Version)
		}
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %s\n", nameWidth, "NAME", typeWidth, "TYPE", pathWidth, "PATH", "VERSION")

	// Print modules
	for _, mod := range modules {
		version := mod.Version
		if version == "" {
			version = "-"
		}
		fmt.Printf("%-*s  %-*s  %-*s  %s\n", nameWidth, mod.Name, typeWidth, mod.Type, pathWidth, mod.Path, version)
	}
}

// printModulesJSON outputs the list of modules in JSON format
func printModulesJSON(modules []ModuleInfo) error {
	output, err := json.MarshalIndent(modules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
