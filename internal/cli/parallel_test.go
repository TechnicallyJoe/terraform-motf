package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunOnModules_Empty(t *testing.T) {
	var buf bytes.Buffer
	called := false

	err := runOnModules([]ModuleInfo{}, false, 4, &buf, &buf, func(mod ModuleInfo, stdout, stderr io.Writer) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if called {
		t.Error("function should not be called for empty module list")
	}
}

func TestRunOnModules_Sequential(t *testing.T) {
	var buf bytes.Buffer
	modules := []ModuleInfo{
		{Name: "mod-a", Path: "path/to/a"},
		{Name: "mod-b", Path: "path/to/b"},
		{Name: "mod-c", Path: "path/to/c"},
	}

	var order []string

	err := runOnModules(modules, false, 4, &buf, &buf, func(mod ModuleInfo, stdout, stderr io.Writer) error {
		order = append(order, mod.Name)
		_, _ = stdout.Write([]byte("processing " + mod.Name + "\n"))
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Sequential should maintain order
	expected := []string{"mod-a", "mod-b", "mod-c"}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("order[%d] = %s, want %s", i, order[i], name)
		}
	}
}

func TestRunOnModules_Parallel(t *testing.T) {
	var buf bytes.Buffer
	modules := []ModuleInfo{
		{Name: "mod-a", Path: "path/to/a"},
		{Name: "mod-b", Path: "path/to/b"},
		{Name: "mod-c", Path: "path/to/c"},
		{Name: "mod-d", Path: "path/to/d"},
	}

	var count atomic.Int32

	err := runOnModules(modules, true, 4, &buf, &buf, func(mod ModuleInfo, stdout, stderr io.Writer) error {
		count.Add(1)
		_, _ = stdout.Write([]byte("processing " + mod.Name + "\n"))
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if count.Load() != 4 {
		t.Errorf("expected 4 modules processed, got %d", count.Load())
	}
}

func TestRunOnModules_CollectsAllErrors(t *testing.T) {
	var buf bytes.Buffer
	modules := []ModuleInfo{
		{Name: "mod-a", Path: "path/to/a"},
		{Name: "mod-b", Path: "path/to/b"},
		{Name: "mod-c", Path: "path/to/c"},
	}

	err := runOnModules(modules, false, 4, &buf, &buf, func(mod ModuleInfo, stdout, stderr io.Writer) error {
		if mod.Name == "mod-a" || mod.Name == "mod-c" {
			return errors.New("failed")
		}
		return nil
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should contain both module names in error
	errStr := err.Error()
	if !strings.Contains(errStr, "mod-a") {
		t.Errorf("error should mention mod-a: %s", errStr)
	}
	if !strings.Contains(errStr, "mod-c") {
		t.Errorf("error should mention mod-c: %s", errStr)
	}
}

func TestRunOnModules_ParallelCollectsAllErrors(t *testing.T) {
	var buf bytes.Buffer
	modules := []ModuleInfo{
		{Name: "mod-a", Path: "path/to/a"},
		{Name: "mod-b", Path: "path/to/b"},
		{Name: "mod-c", Path: "path/to/c"},
	}

	err := runOnModules(modules, true, 4, &buf, &buf, func(mod ModuleInfo, stdout, stderr io.Writer) error {
		if mod.Name == "mod-a" || mod.Name == "mod-c" {
			return errors.New("failed")
		}
		return nil
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should contain both module names in error
	errStr := err.Error()
	if !strings.Contains(errStr, "mod-a") {
		t.Errorf("error should mention mod-a: %s", errStr)
	}
	if !strings.Contains(errStr, "mod-c") {
		t.Errorf("error should mention mod-c: %s", errStr)
	}
}

func TestRunOnModules_BoundedParallelism(t *testing.T) {
	var buf bytes.Buffer
	modules := make([]ModuleInfo, 10)
	for i := range modules {
		modules[i] = ModuleInfo{Name: fmt.Sprintf("mod-%d", i), Path: fmt.Sprintf("path/%d", i)}
	}

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32
	maxJobs := 3

	err := runOnModules(modules, true, maxJobs, &buf, &buf, func(mod ModuleInfo, stdout, stderr io.Writer) error {
		current := concurrent.Add(1)
		// Track max concurrent
		for {
			old := maxConcurrent.Load()
			if current <= old || maxConcurrent.CompareAndSwap(old, current) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond) // Simulate work
		concurrent.Add(-1)
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if maxConcurrent.Load() > int32(maxJobs) {
		t.Errorf("max concurrent %d exceeded limit %d", maxConcurrent.Load(), maxJobs)
	}
}

func TestModuleError(t *testing.T) {
	originalErr := errors.New("original error")
	modErr := &moduleError{
		module: ModuleInfo{Name: "test-mod", Path: "path/to/test"},
		err:    originalErr,
	}

	// Test Error()
	errStr := modErr.Error()
	if !strings.Contains(errStr, "test-mod") {
		t.Errorf("error should contain module name: %s", errStr)
	}
	if !strings.Contains(errStr, "path/to/test") {
		t.Errorf("error should contain module path: %s", errStr)
	}
	if !strings.Contains(errStr, "original error") {
		t.Errorf("error should contain original error: %s", errStr)
	}

	// Test Unwrap()
	unwrapped := modErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() should return original error")
	}
}
