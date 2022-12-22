resource "spectrocloud_backup_storage_location" "bsl1" {
  name        = "dev-backup-s3"
  is_default  = false
  region      = "us-east-2"
  bucket_name = "dev-backup"
  s3 {
    credential_type     = "secret"
    access_key          = "access_key"
    secret_key          = "secret_key"
    s3_force_path_style = false

    #s3_url             = "http://10.90.78.23"
  }
}

resource "spectrocloud_backup_storage_location" "bsl2" {
  name        = "prod-backup-s3"
  is_default  = false
  region      = "us-east-2"
  bucket_name = "prod-backup"
  s3 {
    credential_type     = "sts"
    arn                 = "arn_role"
    external_id         = "external_id"
    s3_force_path_style = false
    #s3_url             = "http://10.90.78.23"
  }
}