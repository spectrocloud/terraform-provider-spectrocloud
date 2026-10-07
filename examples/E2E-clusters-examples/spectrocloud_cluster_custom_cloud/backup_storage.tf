# Creates a dedicated backup storage location for this end-to-end example. Nutanix, like other
# on-prem infrastructure, commonly backs up to a self-hosted, S3-compatible object store (Minio)
# rather than a public cloud bucket.
#
# Day-2 mutability: `storage_provider` is ForceNew. Everything else - name, is_default, region,
# bucket_name, ca_cert, and the s3 credentials - updates in place.
resource "spectrocloud_backup_storage_location" "bsl" {
  name             = "e2e-custom-cloud-backup-location"
  storage_provider = "minio"
  context          = "project"
  is_default       = false

  region      = "minio-default"
  bucket_name = "e2e-custom-cloud-backups"

  s3 {
    s3_url              = var.backup_minio_endpoint
    s3_force_path_style = true
    credential_type     = "secret"
    access_key          = var.backup_minio_access_key
    secret_key          = var.backup_minio_secret_key
  }

  # GCP Cloud Storage alternative (used by the cluster_gcp/cluster_gke examples):
  # storage_provider = "gcp"
  # gcp_storage_config {
  #   project_id            = var.backup_gcp_project_id
  #   gcp_json_credentials  = var.backup_gcp_json_credentials
  # }

  # Azure Blob Storage alternative (used by the cluster_azure/cluster_aks examples):
  # storage_provider = "azure"
  # azure_storage_config {
  #   container_name      = "e2e-backups"
  #   storage_name        = var.backup_azure_storage_account
  #   stock_keeping_unit  = "Standard_LRS"
  #   resource_group      = var.backup_azure_resource_group
  #   azure_tenant_id     = var.azure_tenant_id
  #   azure_client_id     = var.azure_client_id
  #   subscription_id     = var.backup_azure_subscription_id
  #   azure_client_secret = var.azure_client_secret
  # }
}
