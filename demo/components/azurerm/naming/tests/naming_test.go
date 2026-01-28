package tests

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestBasicExample(t *testing.T) {
	t.Parallel()

	terraformDir := filepath.Join("..", "examples", "basic")

	terraformOptions := &terraform.Options{
		TerraformDir: terraformDir,
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Verify outputs are not empty
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	assert.NotEmpty(t, resourceGroupName)

	storageAccountName := terraform.Output(t, terraformOptions, "storage_account_name")
	assert.NotEmpty(t, storageAccountName)

	keyVaultName := terraform.Output(t, terraformOptions, "key_vault_name")
	assert.NotEmpty(t, keyVaultName)
}
