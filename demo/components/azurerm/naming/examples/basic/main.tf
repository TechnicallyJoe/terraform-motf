terraform {
  required_version = ">= 1.5.7"
}

module "naming" {
  source = "../../"
}

output "resource_group_name" {
  description = "Generated resource group name"
  value       = module.naming.resource_group_name
}

output "storage_account_name" {
  description = "Generated storage account name"
  value       = module.naming.storage_account_name
}

output "key_vault_name" {
  description = "Generated key vault name"
  value       = module.naming.key_vault_name
}
