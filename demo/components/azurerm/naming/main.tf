terraform {
  required_version = ">= 1.5.7"
}

module "naming" {
  source  = "Azure/naming/azurerm"
  version = "0.4.3"
}

output "resource_group_name" {
  description = "Generated resource group name"
  value       = module.naming.resource_group.name
}

output "storage_account_name" {
  description = "Generated storage account name"
  value       = module.naming.storage_account.name
}

output "key_vault_names" {
  description = "Generated key vault name"
  value       = module.naming.key_vault.name
}
