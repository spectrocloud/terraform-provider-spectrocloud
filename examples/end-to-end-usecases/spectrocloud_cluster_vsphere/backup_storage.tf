# Backup storage location, created (not looked up) so every attribute is exercised here too.
#
# Day-2 mutability: only `storage_provider` is ForceNew - changing the backend storage type
# recreates the location. `region`, `bucket_name`, `ca_cert`, `is_default`, and the nested `s3`
# block all update in place.
#
# This example uses the default "aws" storage_provider with S3 "secret" (access/secret key)
# credentials - the most common path. "minio" (S3-compatible), "gcp", and "azure" are also
# supported via storage_provider plus the corresponding gcp_storage_config/azure_storage_config
# block instead of `s3` (mutually exclusive with `s3` and with each other - see commented
# alternatives below).
resource "spectrocloud_backup_storage_location" "bsl" {
  name = "e2e-vsphere-backup-location"
  # Optional, default "aws", ForceNew. Allowed: "aws", "minio", "gcp", "azure".
  storage_provider = "aws"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional, default false. Marks this as the tenant/project default backup destination.
  is_default = true
  # Optional. AWS region the bucket lives in.
  region = var.backup_s3_region
  # Optional. Name of the S3 bucket backups are written to.
  bucket_name = var.backup_s3_bucket_name
  # Optional. CA certificate for validating the storage endpoint over TLS - only needed for
  # self-signed/private endpoints (e.g. Minio); left unset for plain AWS S3.
  # ca_cert = file("s3-ca.pem")

  s3 {
    # Optional. Only needed for S3-compatible endpoints (e.g. Minio) - leave unset for AWS S3.
    # s3_url               = "https://minio.corp.example.com"
    # s3_force_path_style  = true

    # Required. "secret" (static access/secret key, used here) or "sts" (assumable role).
    credential_type = "secret"
    # Required when credential_type = "secret" (credential material).
    access_key = var.backup_s3_access_key
    secret_key = var.backup_s3_secret_key
    # Required when credential_type = "sts" instead (omit access_key/secret_key in that case):
    # arn         = var.backup_s3_role_arn
    # external_id = var.backup_s3_external_id
  }

  # GCP alternative (mutually exclusive with `s3` - set storage_provider = "gcp" and use this
  # block instead):
  # gcp_storage_config {
  #   project_id            = var.gcp_project_id
  #   gcp_json_credentials  = file("gcp-credentials.json")
  # }

  # Azure alternative (mutually exclusive with `s3` - set storage_provider = "azure" and use
  # this block instead):
  # azure_storage_config {
  #   container_name      = "backups"
  #   storage_name        = "e2ebackupstorage"
  #   stock_keeping_unit  = "Standard_LRS"
  #   resource_group      = var.azure_resource_group
  #   azure_tenant_id     = var.azure_tenant_id
  #   azure_client_id     = var.azure_client_id
  #   subscription_id     = var.azure_subscription_id
  #   azure_client_secret = var.azure_client_secret
  # }
}
