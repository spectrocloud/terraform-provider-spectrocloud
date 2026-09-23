# Backup storage location, created (not looked up). Uses AWS S3 with static access/secret key
# credentials, matching this cluster's cloud provider.
#
# Day-2 mutability: only `storage_provider` is ForceNew - changing the backend storage type
# recreates the location. Everything else, including the nested `s3` block, updates in place.
resource "spectrocloud_backup_storage_location" "bsl" {
  name = "e2e-aws-backup-location"
  # Optional, default "aws", ForceNew. Allowed: "aws", "minio", "gcp", "azure".
  storage_provider = "aws"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional, default false. Marks this as the tenant/project default backup destination.
  is_default = true
  # Optional. AWS region the bucket lives in.
  region = var.aws_region
  # Optional. Name of the S3 bucket backups are written to.
  bucket_name = var.backup_s3_bucket_name

  s3 {
    # Required. "secret" (static access/secret key, used here) or "sts" (assumable role).
    credential_type = "secret"
    access_key      = var.backup_s3_access_key
    secret_key      = var.backup_s3_secret_key
    # Required when credential_type = "sts" instead:
    # arn         = var.backup_s3_role_arn
    # external_id = var.backup_s3_external_id
  }
}
