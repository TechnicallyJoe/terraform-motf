package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/TechnicallyJoe/sturdy-parakeet/internal/config"
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
	cmd := exec.Command(r.config.Binary, args...)
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
	cmd := exec.Command(r.config.Binary, args...)
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
	cmd := exec.Command(r.config.Binary, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s %s in %s\n", r.config.Binary, strings.Join(args, " "), dir)
	return cmd.Run()
}

// BuildArgs constructs the argument list for a command with extra arguments
func BuildArgs(command string, extraArgs ...string) []string {
	return append([]string{command}, extraArgs...)
}
