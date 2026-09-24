# Backup storage location, created (not looked up). Uses AWS S3 with static access/secret key
# credentials, matching this cluster's cloud provider.
#
# Day-2 mutability: only `storage_provider` is ForceNew - changing the backend storage type
# recreates the location. Everything else, including the nested `s3` block, updates in place.
resource "spectrocloud_backup_storage_location" "bsl" {
  name             = "e2e-eks-backup-location"
  storage_provider = "aws"
  context          = "project"
  is_default       = false
  region           = var.aws_region
  bucket_name      = var.backup_s3_bucket_name

  s3 {
    credential_type = "secret"
    access_key      = var.backup_s3_access_key
    secret_key      = var.backup_s3_secret_key
  }
}
