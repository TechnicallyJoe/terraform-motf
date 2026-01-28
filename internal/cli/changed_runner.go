package cli

import (
	"fmt"
	"path/filepath"
)

// runOnChangedModules detects changed modules and runs fn on each module's absolute path.
// It is a no-op (success) when no changed modules are found.
func runOnChangedModules(fn func(moduleAbsPath string) error) error {
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

	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	for _, mod := range modules {
		moduleAbsPath := filepath.Join(basePath, mod.Path)
		if err := fn(moduleAbsPath); err != nil {
			return fmt.Errorf("%s (%s): %w", mod.Name, mod.Path, err)
		}
	}

	return nil
}
