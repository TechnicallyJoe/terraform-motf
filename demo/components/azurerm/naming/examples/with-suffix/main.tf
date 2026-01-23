terraform {
  required_version = ">= 1.5.7"
}

module "naming" {
  source = "../../"
}

# Custom suffix example
module "naming_with_suffix" {
  source = "Azure/naming/azurerm"
  version = "0.4.3"

  suffix = ["dev", "westeu"]
}

output "default_names" {
  description = "Default generated names"
  value = {
    resource_group  = module.naming.resource_group_name
    storage_account = module.naming.storage_account_name
    key_vault       = module.naming.key_vault_name
  }
}

output "custom_names" {
  description = "Names with custom suffix"
  value = {
    resource_group  = module.naming_with_suffix.resource_group.name
    storage_account = module.naming_with_suffix.storage_account.name
    key_vault       = module.naming_with_suffix.key_vault.name
  }
}
