package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/TechnicallyJoe/terraform-motf/internal/config"
)

// Runner executes terraform/tofu commands using configuration
type Runner struct {
	config *config.Config
}

// NewRunner creates a new Runner with the given configuration
func NewRunner(cfg *config.Config) *Runner {
	return &Runner{config: cfg}
}

// Binary returns the configured binary name
func (r *Runner) Binary() string {
	return r.config.Binary
}

// RunInit executes terraform/tofu init in the specified directory
func (r *Runner) RunInit(dir string, extraArgs ...string) error {
	args := append([]string{"init"}, extraArgs...)
	cmd := exec.Command(r.config.Binary, args...) //nolint:gosec // Binary is validated to be terraform or tofu
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s %s in %s\n", r.config.Binary, strings.Join(args, " "), dir)
	return cmd.Run()
}

// RunFmt executes terraform/tofu fmt in the specified directory
func (r *Runner) RunFmt(dir string, extraArgs ...string) error {
	args := append([]string{"fmt"}, extraArgs...)
	cmd := exec.Command(r.config.Binary, args...) //nolint:gosec // Binary is validated to be terraform or tofu
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s %s in %s\n", r.config.Binary, strings.Join(args, " "), dir)
	return cmd.Run()
}

// RunValidate executes terraform/tofu validate in the specified directory
func (r *Runner) RunValidate(dir string, extraArgs ...string) error {
	args := append([]string{"validate"}, extraArgs...)
	cmd := exec.Command(r.config.Binary, args...) //nolint:gosec // Binary is validated to be terraform or tofu
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s %s in %s\n", r.config.Binary, strings.Join(args, " "), dir)
	return cmd.Run()
}

// RunPlan executes terraform/tofu plan in the specified directory
func (r *Runner) RunPlan(dir string, extraArgs ...string) error {
	args := append([]string{"plan"}, extraArgs...)
	cmd := exec.Command(r.config.Binary, args...) //nolint:gosec // Binary is validated to be terraform or tofu
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s %s in %s\n", r.config.Binary, strings.Join(args, " "), dir)
	return cmd.Run()
}

// RunTest executes tests based on the configured test engine
func (r *Runner) RunTest(dir string, extraArgs ...string) error {
	var cmd *exec.Cmd
	var cmdArgs []string

	switch r.config.Test.Engine {
	case "terratest":
		// Terratest uses Go test
		cmdArgs = []string{"test", "./..."}

		// Add config args if present
		if r.config.Test.Args != "" {
			configArgs := strings.Fields(r.config.Test.Args)
			cmdArgs = append(cmdArgs, configArgs...)
		}

		// Add extra args from command line
		cmdArgs = append(cmdArgs, extraArgs...)

		cmd = exec.Command("go", cmdArgs...) //nolint:gosec // cmdArgs are constructed from validated config
		fmt.Printf("Running go %s in %s\n", strings.Join(cmdArgs, " "), dir)
	case "terraform", "tofu":
		// Terraform/Tofu native test command
		cmdArgs = []string{"test"}

		// Add config args if present
		if r.config.Test.Args != "" {
			configArgs := strings.Fields(r.config.Test.Args)
			cmdArgs = append(cmdArgs, configArgs...)
		}

		// Add extra args from command line
		cmdArgs = append(cmdArgs, extraArgs...)

		binary := r.config.Test.Engine
		cmd = exec.Command(binary, cmdArgs...) //nolint:gosec // binary is validated to be terraform or tofu
		fmt.Printf("Running %s %s in %s\n", binary, strings.Join(cmdArgs, " "), dir)
	default:
		return fmt.Errorf("unsupported test engine: %s", r.config.Test.Engine)
	}

	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
