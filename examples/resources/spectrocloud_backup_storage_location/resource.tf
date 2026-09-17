# Common attributes across all backup storage locations, regardless of provider:
#   name             - Required.
#   storage_provider - Optional, ForceNew, default "aws". Allowed: "aws", "minio", "gcp", "azure".
#                       This is the only ForceNew attribute on the resource - changing it
#                       recreates the backup location under a new provider.
#   context          - Optional, default "project". Allowed: "project", "tenant".
#   is_default       - Optional, default false. Whether new backups use this location by default.
#   region           - Optional in the schema, but a CustomizeDiff (schemaValidationForLocationProvider)
#                       requires it (along with `s3`/`bucket_name`) whenever storage_provider is
#                       "aws" or "minio", and forbids it entirely for "azure"/"gcp".
#   bucket_name      - Same CustomizeDiff: required for "aws"/"minio" (with `s3`/`region`) and for
#                       "gcp" (with `gcp_storage_config`); forbidden for "azure".
#   ca_cert          - Same CustomizeDiff: only meaningful for "aws"/"minio"; forbidden for
#                       "azure"/"gcp".
# The same CustomizeDiff also forbids `azure_storage_config`/`gcp_storage_config` when
# storage_provider is "aws"/"minio", and forbids `s3` when it's "azure"/"gcp" - each provider's
# block below is mutually exclusive with the others in practice, not just by convention.
# All of the above (aside from storage_provider) update in place.

// S3 backup location, using static access/secret key credentials
resource "spectrocloud_backup_storage_location" "bsl_s3" {
  name             = "project-dev-bsl-s3"
  context          = "project"
  storage_provider = "aws"
  is_default       = false
  region           = "us-east-1"
  bucket_name      = "project-backup-bucket-s3"
  # ca_cert          = "REPLACE_ME"

  # s3 block (static access/secret key credentials, at most one):
  #   credential_type     - Required. Allowed: "secret" (static keys) or "sts" (assumable IAM role).
  #   access_key          - Required when credential_type = "secret", sensitive.
  #   secret_key          - Required when credential_type = "secret", sensitive.
  #   s3_force_path_style - Optional. Forces path-style S3 URLs (bucket in the path, not the
  #                         hostname) - needed for most S3-compatible services (e.g. Minio).
  #   s3_url              - Optional. The S3 endpoint URL.
  s3 {
    credential_type     = "secret"
    access_key          = var.aws_access_key
    secret_key          = var.aws_secret_key
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-1.amazonaws.com"
  }
}

// Minio (S3-compatible) backup location, using static credentials
resource "spectrocloud_backup_storage_location" "bsl_minio" {
  name             = "project-dev-minio-bsl"
  context          = "project"
  storage_provider = "minio"
  is_default       = false
  region           = "us-east-2"
  bucket_name      = "project-backup-bucket-minio"

  # s3 block (Minio, static credentials, at most one):
  #   credential_type     - Required. Allowed: "secret" (static keys) or "sts" (assumable IAM role).
  #   access_key          - Required when credential_type = "secret", sensitive.
  #   secret_key          - Required when credential_type = "secret", sensitive.
  #   s3_force_path_style - Optional. Minio (and most other S3-compatible services) require
  #                         path-style URLs (bucket in the path, not the hostname).
  #   s3_url              - Optional. The S3 endpoint URL.
  s3 {
    credential_type     = "secret"
    access_key          = var.aws_access_key
    secret_key          = var.aws_secret_key
    s3_force_path_style = true
    s3_url              = "http://10.90.78.23"
  }
}

// GCP backup location
resource "spectrocloud_backup_storage_location" "bsl_gcp" {
  name             = "project-dev-gcp"
  context          = "project"
  storage_provider = "gcp"
  is_default       = false
  bucket_name      = "project-backup-bucket-gcp"

  # gcp_storage_config block (Required when storage_provider = "gcp", at most one):
  #   project_id           - Required.
  #   gcp_json_credentials - Required, sensitive.
  gcp_storage_config {
    project_id           = "test-gcp-project-id"
    gcp_json_credentials = var.gcp_json_credentials
  }
}

// Azure backup location
resource "spectrocloud_backup_storage_location" "bsl_azure" {
  name             = "project-dev-azure-bsl"
  context          = "project"
  storage_provider = "azure"
  is_default       = false

  # azure_storage_config block (Required when storage_provider = "azure", at most one - every
  # field here is Required):
  #   stock_keeping_unit  - Azure storage account SKU, e.g. "Standard_LRS".
  #   azure_client_secret - Sensitive.
  #   (container_name, storage_name, resource_group, azure_tenant_id, azure_client_id,
  #   subscription_id carry no further constraints beyond being Required.)
  azure_storage_config {
    container_name      = "test-container"
    storage_name        = "test-storage"
    stock_keeping_unit  = "Standard_LRS"
    resource_group      = "test-resource-group"
    azure_tenant_id     = "test-azure-tenant-id"
    azure_client_id     = "test-azure-client-id"
    subscription_id     = "test-azure-subscription-id"
    azure_client_secret = var.azure_client_secret
  }
}

// S3 backup location, using an assumable IAM role (STS) instead of static keys
resource "spectrocloud_backup_storage_location" "bsl_sts" {
  name        = "tenant-dev-1"
  context     = "tenant"
  is_default  = false
  region      = "us-east-2"
  bucket_name = "tenant-backup-bucket-sts"

  # s3 block (STS assumable IAM role, at most one):
  #   credential_type     - Required. Allowed: "secret" (static keys) or "sts" (assumable IAM role).
  #   arn                 - Required when credential_type = "sts". The IAM role ARN to assume.
  #   external_id         - Optional. Used alongside `arn` for cross-account STS role assumption
  #                         when credential_type = "sts".
  #   s3_force_path_style - Optional. Forces path-style S3 URLs (bucket in the path, not the
  #                         hostname) - needed for most S3-compatible services (e.g. Minio).
  #   s3_url              - Optional. The S3 endpoint URL.
  s3 {
    credential_type     = "sts"
    arn                 = var.aws_sts_role_arn
    external_id         = var.aws_external_id
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-2.amazonaws.com"
  }
}
