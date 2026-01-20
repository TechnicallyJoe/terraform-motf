# Demo Storage Account Component
terraform {
  required_version = ">= 1.0.0"
}

provider "azurerm" {
  features {}
}

variable "name" {
  type        = string
  description = "The name of the storage account"
  default     = "demostorageaccount"
}

variable "resource_group_name" {
  type        = string
  description = "The name of the resource group"
  default     = "demo-rg"
}

variable "location" {
  type        = string
  description = "The Azure region"
  default     = "eastus"
}

variable "account_tier" {
  type        = string
  description = "The storage account tier"
  default     = "Standard"
}

variable "account_replication_type" {
  type        = string
  description = "The replication type"
  default     = "LRS"
}

output "name" {
  value       = var.name
  description = "The storage account name"
}

output "resource_group_name" {
  value       = var.resource_group_name
  description = "The resource group name"
}
