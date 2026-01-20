package finder

import (
	"os"
	"path/filepath"
	"strings"
)

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
