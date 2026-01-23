package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TaskConfig represents a custom task definition
type TaskConfig struct {
	Description string `yaml:"description"`
	Shell       string `yaml:"shell"`
	Command     string `yaml:"command"`
}

// ShellConfig defines how to invoke a shell
type ShellConfig struct {
	Binary string
	Args   []string // args before the script
}

// DefaultShell is the default shell when none is specified
const DefaultShell = "sh"

// Shells is the registry of supported shells
var Shells = map[string]ShellConfig{
	"sh":   {Binary: "sh", Args: []string{"-c"}},
	"bash": {Binary: "bash", Args: []string{"-c"}},
	"pwsh": {Binary: "pwsh", Args: []string{"-Command"}},
	"cmd":  {Binary: "cmd", Args: []string{"/C"}},
}

// SupportedShells returns a sorted list of supported shell names
func SupportedShells() []string {
	return []string{"bash", "cmd", "pwsh", "sh"}
}

// GetShellArgs returns the binary and arguments needed to execute a script with the given shell.
// Returns an error if the shell is not supported.
func GetShellArgs(shell, script string) (binary string, args []string, err error) {
	if shell == "" {
		shell = DefaultShell
	}

	shellCfg, ok := Shells[shell]
	if !ok {
		return "", nil, fmt.Errorf("unknown shell '%s', supported: %s", shell, strings.Join(SupportedShells(), ", "))
	}

	args = append(shellCfg.Args, script)
	return shellCfg.Binary, args, nil
}

// Runner executes custom tasks
type Runner struct {
	Tasks map[string]*TaskConfig
}

// NewRunner creates a new task runner with the given task definitions
func NewRunner(tasks map[string]*TaskConfig) *Runner {
	if tasks == nil {
		tasks = make(map[string]*TaskConfig)
	}
	return &Runner{Tasks: tasks}
}

// GetTask returns the task config for the given name, or nil if not found
func (r *Runner) GetTask(name string) *TaskConfig {
	return r.Tasks[name]
}

// ListTasks returns all task names
func (r *Runner) ListTasks() []string {
	names := make([]string, 0, len(r.Tasks))
	for name := range r.Tasks {
		names = append(names, name)
	}
	return names
}

// Run executes a task by name in the given working directory
func (r *Runner) Run(taskName, workDir string) error {
	task := r.GetTask(taskName)
	if task == nil {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	if task.Command == "" {
		return fmt.Errorf("task '%s' has no command defined", taskName)
	}

	binary, args, err := GetShellArgs(task.Shell, task.Command)
	if err != nil {
		return fmt.Errorf("task '%s': %w", taskName, err)
	}

	fmt.Printf("Running task '%s' in %s\n", taskName, workDir)
	fmt.Printf("$ %s\n", task.Command)

	cmd := exec.Command(binary, args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
