package finder

import (
	"os"
	"path/filepath"
	"strings"
)

// skipDirs contains directory names that should be skipped during module discovery
var skipDirs = map[string]bool{
	".terraform":   true,
	".git":         true,
	"node_modules": true,
	"examples":     true,
	"modules":      true,
	"tests":        true,
	".spacelift":   true,
}

// FindModule searches for a module with the given name in the specified search path
// It recursively searches subdirectories and returns all matching directories
// Only directories containing .tf or .tf.json files are considered valid modules
func FindModule(searchPath, moduleName string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Skip excluded directories
		if skipDirs[d.Name()] {
			return filepath.SkipDir
		}

		// Check if directory name matches the module name
		if d.Name() == moduleName {
			// Verify it contains .tf or .tf.json files
			if hasTerraformFiles(path) {
				matches = append(matches, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

// hasTerraformFiles checks if a directory contains any .tf or .tf.json files
func hasTerraformFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".tf" || strings.HasSuffix(name, ".tf.json") {
			return true
		}
	}

	return false
}

// ListAllModules finds all modules in the specified search path
// Returns a map of module names to their paths
func ListAllModules(searchPath string) (map[string]string, error) {
	modules := make(map[string]string)

	err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Skip excluded directories
		if skipDirs[d.Name()] {
			return filepath.SkipDir
		}

		// Check if this directory contains terraform files
		if hasTerraformFiles(path) {
			// Use the directory name as the module name
			moduleName := d.Name()
			// Store the path (only if not already seen or if it's a shorter path)
			if _, exists := modules[moduleName]; !exists {
				modules[moduleName] = path
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil
}

// MatchesWildcard checks if a name matches a wildcard pattern
// Supports * as a wildcard for any number of characters
func MatchesWildcard(name, pattern string) bool {
	// Convert wildcard pattern to a simple match
	// Split by * to get the parts that must be present
	parts := strings.Split(pattern, "*")

	// If no wildcards, do exact match
	if len(parts) == 1 {
		return name == pattern
	}

	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}

		idx := strings.Index(name[pos:], part)
		if idx == -1 {
			return false
		}

		// For the first part, it must be at the start (unless pattern starts with *)
		if i == 0 && pattern[0] != '*' && idx != 0 {
			return false
		}

		pos += idx + len(part)
	}

	// For the last part, check if pattern ends with *
	if len(parts) > 0 && parts[len(parts)-1] != "" && !strings.HasSuffix(pattern, "*") {
		// Pattern doesn't end with *, so the last part must be at the end
		return strings.HasSuffix(name, parts[len(parts)-1])
	}

	return true
}
