package terraform

import (
	"fmt"
	"os"
	"os/exec"
)

// Binary holds the terraform/tofu binary name
var Binary = "terraform"

// SetBinary sets the binary to use for terraform commands
func SetBinary(binary string) {
	Binary = binary
}

// RunInit executes terraform/tofu init in the specified directory
func RunInit(dir string) error {
	cmd := exec.Command(Binary, "init")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s init in %s\n", Binary, dir)
	return cmd.Run()
}

// RunFmt executes terraform/tofu fmt in the specified directory
func RunFmt(dir string) error {
	cmd := exec.Command(Binary, "fmt")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s fmt in %s\n", Binary, dir)
	return cmd.Run()
}

// RunValidate executes terraform/tofu validate in the specified directory
func RunValidate(dir string) error {
	cmd := exec.Command(Binary, "validate")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Running %s validate in %s\n", Binary, dir)
	return cmd.Run()
}
