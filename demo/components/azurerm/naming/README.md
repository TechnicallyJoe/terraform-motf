# Naming Component

A simple component that wraps the Azure naming module to generate consistent resource names.

This component has no Azure provider dependency and can be used for testing purposes.

## Usage

```hcl
module "naming" {
  source = "../../components/azurerm/naming"
}

output "rg_name" {
  value = module.naming.resource_group_name
}
```

## Outputs

| Name | Description |
|------|-------------|
| `resource_group_name` | Generated resource group name |
| `storage_account_name` | Generated storage account name |
| `key_vault_name` | Generated key vault name |
