package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

// runOnChangedModules detects changed modules and runs fn on each module.
// When parallelFlag is set, modules are processed concurrently.
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
