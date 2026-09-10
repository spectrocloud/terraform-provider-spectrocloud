# Backup storage location, created (not looked up). Uses GCP Cloud Storage, matching this
# cluster's cloud provider.
#
# Day-2 mutability: only `storage_provider` is ForceNew - changing the backend storage type
# recreates the location. Everything else, including the nested `gcp_storage_config` block,
# updates in place.
resource "spectrocloud_backup_storage_location" "bsl" {
  name = "e2e-gcp-backup-location"
  # Optional, ForceNew. Allowed: "aws", "minio", "gcp", "azure" - "gcp" here to match this
  # cluster's cloud.
  storage_provider = "gcp"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional, default false. Marks this as the tenant/project default backup destination.
  is_default = false

  gcp_storage_config {
    project_id           = var.backup_gcp_project_id
    gcp_json_credentials = var.backup_gcp_json_credentials
  }
}
