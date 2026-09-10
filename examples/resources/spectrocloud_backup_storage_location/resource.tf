# Common attributes across all backup storage locations, regardless of provider:
#   name             - Required.
#   storage_provider - Optional, ForceNew, default "aws". Allowed: "aws", "minio", "gcp", "azure".
#                       This is the only ForceNew attribute on the resource - changing it
#                       recreates the backup location under a new provider.
#   context          - Optional, default "project". Allowed: "project", "tenant".
#   is_default       - Optional, default false. Whether new backups use this location by default.
#   region           - Optional. Relevant to S3/Minio.
#   bucket_name      - Optional. Relevant to S3/Minio/GCP.
#   ca_cert          - Optional. Relevant to S3/Minio, for a self-signed storage endpoint.
# All of the above (aside from storage_provider) update in place.

// S3 backup location, using static access/secret key credentials
resource "spectrocloud_backup_storage_location" "bsl_s3" {
  name             = "project-dev-bsl-s3"
  context          = "project"
  storage_provider = "aws"
  is_default       = false
  region           = "us-east-1"
  bucket_name      = "project-backup-bucket-s3"
  # ca_cert          = "REPLACE_ME" # Optional, for a self-signed S3-compatible endpoint.

  # Required when storage_provider = "aws" or "minio". At most one block.
  s3 {
    # Required. Allowed: "secret" (static keys, shown here) or "sts" (see bsl_sts below).
    credential_type = "secret"
    # Required when credential_type = "secret". Both sensitive.
    access_key = var.aws_access_key
    secret_key = var.aws_secret_key
    # Optional. Forces path-style S3 URLs (bucket in the path, not the hostname) - needed for
    # most S3-compatible services, see bsl_minio below.
    s3_force_path_style = false
    # Optional. The S3 endpoint URL.
    s3_url = "https://s3.us-east-1.amazonaws.com"
    # arn / external_id are only used with credential_type = "sts" - see bsl_sts below.
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
  s3 {
    credential_type = "secret"
    access_key      = var.aws_access_key
    secret_key      = var.aws_secret_key
    # Minio (and most other S3-compatible services) require path-style URLs.
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

  # Required when storage_provider = "gcp". At most one block.
  gcp_storage_config {
    project_id           = "test-gcp-project-id"    # Required.
    gcp_json_credentials = var.gcp_json_credentials # Required, sensitive.
  }
}

// Azure backup location
resource "spectrocloud_backup_storage_location" "bsl_azure" {
  name             = "project-dev-azure-bsl"
  context          = "project"
  storage_provider = "azure"
  is_default       = false

  # Required when storage_provider = "azure". At most one block - every field here is Required.
  azure_storage_config {
    container_name      = "test-container"
    storage_name        = "test-storage"
    stock_keeping_unit  = "Standard_LRS" # Azure storage account SKU, e.g. "Standard_LRS".
    resource_group      = "test-resource-group"
    azure_tenant_id     = "test-azure-tenant-id"
    azure_client_id     = "test-azure-client-id"
    subscription_id     = "test-azure-subscription-id"
    azure_client_secret = var.azure_client_secret # Sensitive.
  }
}

// S3 backup location, using an assumable IAM role (STS) instead of static keys
resource "spectrocloud_backup_storage_location" "bsl_sts" {
  name        = "tenant-dev-1"
  context     = "tenant"
  is_default  = false
  region      = "us-east-2"
  bucket_name = "tenant-backup-bucket-sts"
  s3 {
    credential_type = "sts"
    # Required when credential_type = "sts".
    arn = var.aws_sts_role_arn
    # Optional, used alongside the role ARN for cross-account STS role assumption.
    external_id         = var.aws_external_id
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-2.amazonaws.com"
  }
}
