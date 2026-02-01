package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
	"github.com/TechnicallyJoe/terraform-motf/internal/finder"
	"github.com/TechnicallyJoe/terraform-motf/internal/git"
)

// runOnChangedModules detects changed modules and runs fn on each module.
// When parallelFlag is set, modules are processed concurrently.
// parallelFlag is a package-level CLI flag set by command-line arguments.
// It is a no-op (success) when no changed modules are found.
//
// The function signature for fn receives stdout/stderr writers to support
// prefixed output in parallel mode.
func runOnChangedModules(fn func(mod ModuleInfo, stdout, stderr io.Writer) error) error {
	if pathFlag != "" {
		return fmt.Errorf("--changed cannot be used with --path")
	}
	if exampleFlag != "" {
		return fmt.Errorf("--changed cannot be used with --example")
	}

	modules, err := detectChangedModules(refFlag)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		fmt.Println("No changed modules found")
		return nil
	}

	var parallelismCfg *config.ParallelismConfig
	if cfg != nil {
		parallelismCfg = cfg.Parallelism
	}

	return RunOnModulesParallel(modules, parallelismCfg, fn)
}

// runOnChangedModulesWithPath is a convenience wrapper for commands that need
// the module's absolute path. It wraps fn to provide the path from ModuleInfo.
func runOnChangedModulesWithPath(fn func(moduleAbsPath string, stdout, stderr io.Writer) error) error {
	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	return runOnChangedModules(func(mod ModuleInfo, stdout, stderr io.Writer) error {
		moduleAbsPath := filepath.Join(basePath, mod.Path)
		return fn(moduleAbsPath, stdout, stderr)
	})
}

// detectChangedModules returns modules that have changed compared to baseRef.
// If baseRef is empty, it auto-detects the default branch by checking origin/HEAD,
// then falling back to origin/main or origin/master.
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
