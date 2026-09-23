# Backup storage location, created (not looked up). Uses Azure Blob storage, matching this
# cluster's cloud provider.
#
# Day-2 mutability: only `storage_provider` is ForceNew - changing the backend storage type
# recreates the location. Everything else, including the nested `azure_storage_config` block,
# updates in place.
resource "spectrocloud_backup_storage_location" "bsl" {
  name = "e2e-aks-backup-location"
  # Optional, ForceNew. Allowed: "aws", "minio", "gcp", "azure".
  storage_provider = "azure"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional, default false. Marks this as the tenant/project default backup destination.
  is_default = false

  azure_storage_config {
    container_name      = var.backup_azure_container_name
    storage_name        = var.backup_azure_storage_name
    stock_keeping_unit  = "Standard_LRS"
    resource_group      = var.backup_azure_resource_group
    azure_tenant_id     = var.azure_tenant_id
    azure_client_id     = var.azure_client_id
    subscription_id     = var.azure_subscription_id
    azure_client_secret = var.azure_client_secret
  }
}
