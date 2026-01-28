package tests

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestBasicExample(t *testing.T) {
	t.Parallel()

	// Define the path to the Terraform code that will be tested.
	terraformDir := filepath.Join("..", "examples", "basic")

	// Configure Terraform options with default retryable errors to handle the most common retryable errors in terraform testing.
	terraformOptions := &terraform.Options{
		// The path to where your Terraform code is located
		TerraformDir: terraformDir,
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created.
	// defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform validate` and fail the test if there are any errors.
	terraform.InitAndValidate(t, terraformOptions)

	// Example output check (customize as needed) - Only works if apply is run.
	// output := terraform.Output(t, terraformOptions, "resource_group_name")
	// assert.NotEmpty(t, output)
}
