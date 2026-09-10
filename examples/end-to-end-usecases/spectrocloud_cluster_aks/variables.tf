# --- Azure cloud account ---
variable "azure_tenant_id" {
  description = "Azure AD tenant ID."
}
variable "azure_client_id" {
  description = "Azure AD application (client) ID."
}
variable "azure_client_secret" {
  description = "Azure AD application client secret (credential material)."
  sensitive   = true
}

# --- Cluster placement ---
variable "azure_subscription_id" {
  description = "Azure subscription ID."
}
variable "azure_resource_group" {
  description = "Azure resource group the cluster is provisioned into."
}
variable "azure_region" {
  description = "Azure region, e.g. eastus."
  default     = "eastus"
}
variable "cluster_ssh_public_key" {
  description = "Public SSH key injected into all cluster nodes."
}

# --- Backup storage location (Azure Blob) ---
variable "backup_azure_storage_name" {
  description = "Azure storage account name for the backup storage location."
}
variable "backup_azure_container_name" {
  description = "Blob container name for the backup storage location."
  default     = "backups"
}
variable "backup_azure_resource_group" {
  description = "Resource group of the storage account used for backups."
}

# --- Local kubeconfig export ---
variable "kubeconfig_output_dir" {
  description = "Local directory to write the cluster's kubeconfig files into."
  default     = "."
}
