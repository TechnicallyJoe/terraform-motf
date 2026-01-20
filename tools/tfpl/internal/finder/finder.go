package finder

import (
	"os"
	"path/filepath"
	"strings"
)

// FindModule recursively searches for directories matching moduleName in searchPath
// Only returns directories that contain .tf or .tf.json files
func FindModule(searchPath, moduleName string) ([]string, error) {
	var matches []string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			return nil
		}

		// Check if this is a directory with the matching name
		if info.IsDir() && info.Name() == moduleName {
			// Check if this directory contains .tf or .tf.json files
			if hasTerraformFiles(path) {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
}

// hasTerraformFiles checks if a directory contains .tf or .tf.json files
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
		if filepath.Ext(name) == ".tf" {
			return true
		}
		// Check for .tf.json
		if strings.HasSuffix(name, ".tf.json") {
			return true
		}
	}

	return false
}
