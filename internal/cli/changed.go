package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TechnicallyJoe/terraform-motf/internal/finder"
	"github.com/TechnicallyJoe/terraform-motf/internal/git"
	"github.com/spf13/cobra"
)

// changedJsonFlag controls JSON output for changed command
var changedJsonFlag bool

// changedNamesOnlyFlag outputs only module names (not paths)
var changedNamesOnlyFlag bool

// changedCmd represents the changed command
var changedCmd = &cobra.Command{
	Use:   "changed",
	Short: "List modules with changes compared to a base branch",
	Long: `List modules that have file changes compared to a base branch or commit.

This command detects which modules (components, bases, projects) have been
modified in the current branch compared to a base reference (default: origin/main).

Useful for CI pipelines to only run commands on affected modules.

Examples:
  motf changed                          # Compare against auto-detected default branch
  motf changed --ref origin/main        # Compare against origin/main
  motf changed --ref HEAD~5             # Compare against 5 commits ago
  motf changed --json                   # Output as JSON array
  motf changed --names                  # Output only module names (one per line)

CI Usage:
  # Run validate only on changed modules
  for module in $(motf changed --names); do
    motf validate "$module"
  done

  # Or with xargs
  motf changed --names | xargs -I {} motf fmt {}`,
	RunE: runChanged,
}

func init() {
	changedCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref to compare against (default: auto-detect from origin/HEAD)")
	changedCmd.Flags().BoolVar(&changedJsonFlag, "json", false, "Output in JSON format")
	changedCmd.Flags().BoolVar(&changedNamesOnlyFlag, "names", false, "Output only module names (one per line)")
	rootCmd.AddCommand(changedCmd)
}

func runChanged(cmd *cobra.Command, args []string) error {
	modules, err := detectChangedModules(refFlag)
	if err != nil {
		return err
	}
	return outputChangedModules(modules)
}

// detectChangedModules returns modules that have changed compared to baseRef.
// If baseRef is empty, it auto-detects the default branch (origin/HEAD then origin/main/master).
func detectChangedModules(baseRef string) ([]ModuleInfo, error) {
	// Get the git repository root
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get git root: %w", err)
	}

	// Determine base ref
	base := baseRef
	if base == "" {
		detectedBase, err := git.GetDefaultBranch()
		if err != nil {
			return nil, fmt.Errorf("could not auto-detect base branch (use --ref to specify): %w", err)
		}
		base = detectedBase
	}

	// Get changed files
	changedFiles, err := git.GetChangedFiles(repoRoot, base)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	if len(changedFiles) == 0 {
		return nil, nil
	}

	// Get base path for module discovery
	basePath, err := getBasePath()
	if err != nil {
		return nil, err
	}

	// Calculate relative path from repo root to base path
	relBasePath, err := filepath.Rel(repoRoot, basePath)
	if err != nil {
		relBasePath = ""
	}

	// Adjust module dirs to be relative to repo root
	var adjustedModuleDirs []string
	for _, dir := range ModuleDirs {
		if relBasePath != "" && relBasePath != "." {
			adjustedModuleDirs = append(adjustedModuleDirs, filepath.ToSlash(filepath.Join(relBasePath, dir)))
		} else {
			adjustedModuleDirs = append(adjustedModuleDirs, dir)
		}
	}

	// Map changed files to module paths
	changedModulePaths := git.MapFilesToModules(changedFiles, adjustedModuleDirs)
	if len(changedModulePaths) == 0 {
		return nil, nil
	}

	// Convert paths to module info with validation
	modules := resolveChangedModules(basePath, repoRoot, changedModulePaths)

	return modules, nil
}

// resolveChangedModules validates that changed paths are actual modules with .tf files
// and returns module info for each
func resolveChangedModules(basePath, repoRoot string, changedPaths []string) []ModuleInfo {
	var modules []ModuleInfo
	seen := make(map[string]bool)

	for _, modulePath := range changedPaths {
		// Convert the path (relative to repo root) to absolute
		absPath := filepath.Join(repoRoot, modulePath)

		// Check if this directory contains terraform files
		if !finder.HasTerraformFiles(absPath) {
			// The changed file might be in a subdirectory (like tests/ or examples/)
			// Walk up to find the actual module
			absPath = findParentModule(absPath, basePath)
			if absPath == "" {
				continue
			}
			// Recalculate the relative path for deduplication
			relPath, err := filepath.Rel(repoRoot, absPath)
			if err != nil {
				continue
			}
			modulePath = filepath.ToSlash(relPath)
		}

		// Deduplicate
		if seen[modulePath] {
			continue
		}
		seen[modulePath] = true

		// Get module name (last component of path)
		name := filepath.Base(absPath)

		// Make path relative to basePath for display
		displayPath, err := filepath.Rel(basePath, absPath)
		if err != nil {
			displayPath = modulePath
		}

		modules = append(modules, ModuleInfo{
			Name: name,
			Type: getModuleType(absPath),
			Path: displayPath,
		})
	}

	// Sort by path for consistent output
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	return modules
}

// findParentModule walks up the directory tree to find a parent that contains .tf files
func findParentModule(startPath, stopPath string) string {
	current := startPath
	for {
		if finder.HasTerraformFiles(current) {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current || parent == stopPath || !strings.HasPrefix(parent, stopPath) {
			return ""
		}
		current = parent
	}
}

// outputChangedModules outputs the list of changed modules
func outputChangedModules(modules []ModuleInfo) error {
	if len(modules) == 0 {
		if changedJsonFlag {
			fmt.Println("[]")
			return nil
		}
		if !changedNamesOnlyFlag {
			fmt.Println("No changed modules found")
		}
		return nil
	}

	if changedJsonFlag {
		output, err := json.MarshalIndent(modules, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	if changedNamesOnlyFlag {
		for _, mod := range modules {
			fmt.Println(mod.Name)
		}
		return nil
	}

	// Table format (same as list command)
	nameWidth := len("NAME")
	typeWidth := len("TYPE")
	pathWidth := len("PATH")

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
	}

	fmt.Printf("%-*s  %-*s  %s\n", nameWidth, "NAME", typeWidth, "TYPE", "PATH")
	for _, mod := range modules {
		fmt.Printf("%-*s  %-*s  %s\n", nameWidth, mod.Name, typeWidth, mod.Type, mod.Path)
	}

	return nil
}
