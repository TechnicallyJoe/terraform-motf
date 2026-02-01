package cli

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

// ModuleRunner is a function that runs a command on a module
// with the given stdout and stderr writers.
type ModuleRunner func(mod ModuleInfo, stdout, stderr io.Writer) error

// runOnModules executes fn on each module, either sequentially or in parallel
// based on the parallel flag. When parallel is true, it uses a worker pool
// with bounded concurrency.
//
// Parameters:
//   - modules: list of modules to process
//   - parallel: whether to run in parallel
//   - maxJobs: maximum concurrent jobs (use resolveMaxJobs to compute)
//   - out: output writer for prefixed output (typically os.Stdout)
//   - errOut: error output writer (typically os.Stderr)
//   - fn: function to run on each module
//
// Returns combined errors from all failed modules (does not fail fast).
func runOnModules(modules []ModuleInfo, parallel bool, maxJobs int, out, errOut io.Writer, fn ModuleRunner) error {
	if len(modules) == 0 {
		return nil
	}

	// Calculate max name length for alignment
	maxNameLen := 0
	for _, mod := range modules {
		if len(mod.Name) > maxNameLen {
			maxNameLen = len(mod.Name)
		}
	}

	if !parallel {
		return runSequential(modules, maxNameLen, out, errOut, fn)
	}

	return runParallel(modules, maxJobs, maxNameLen, out, errOut, fn)
}

// runSequential runs fn on each module one at a time
func runSequential(modules []ModuleInfo, maxNameLen int, out, errOut io.Writer, fn ModuleRunner) error {
	var errs []error
	mu := &sync.Mutex{} // For consistent output even in sequential mode

	for i, mod := range modules {
		writers := newPrefixedWriterPair(mod.Name, maxNameLen, i, out, errOut, mu)
		if err := fn(mod, writers.stdout, writers.stderr); err != nil {
			errs = append(errs, &moduleError{module: mod, err: err})
		}
		_ = writers.Flush()
	}

	return errors.Join(errs...)
}

// runParallel runs fn on modules concurrently with bounded parallelism
func runParallel(modules []ModuleInfo, maxJobs int, maxNameLen int, out, errOut io.Writer, fn ModuleRunner) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	// Semaphore channel for bounded concurrency
	sem := make(chan struct{}, maxJobs)

	// Shared mutex for output synchronization
	outputMu := &sync.Mutex{}

	for i, mod := range modules {
		wg.Add(1)
		go func(index int, m ModuleInfo) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			writers := newPrefixedWriterPair(m.Name, maxNameLen, index, out, errOut, outputMu)
			if err := fn(m, writers.stdout, writers.stderr); err != nil {
				mu.Lock()
				errs = append(errs, &moduleError{module: m, err: err})
				mu.Unlock()
			}
			_ = writers.Flush()
		}(i, mod)
	}

	wg.Wait()
	return errors.Join(errs...)
}

// moduleError wraps an error with module context
type moduleError struct {
	module ModuleInfo
	err    error
}

func (e *moduleError) Error() string {
	return e.module.Name + " (" + e.module.Path + "): " + e.err.Error()
}

func (e *moduleError) Unwrap() error {
	return e.err
}

// RunOnModulesParallel is a convenience function that uses the global
// parallelFlag along with config to run on modules.
// This is the primary entry point for commands using --changed.
//
// Note: CLI flags are merged into config during PersistentPreRunE,
// so parallelismCfg already reflects any --max-parallel override.
func RunOnModulesParallel(modules []ModuleInfo, parallelismCfg *config.ParallelismConfig, fn ModuleRunner) error {
	return runOnModules(modules, parallelFlag, parallelismCfg.GetMaxJobs(), os.Stdout, os.Stderr, fn)
}
