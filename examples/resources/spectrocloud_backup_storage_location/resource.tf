resource "spectrocloud_backup_storage_location" "bsl1" {
  name        = "aaa-project-dev-1"
  context     = "project"
  is_default  = false
  region      = "us-east-1"
  bucket_name = "project-backup-2"
  s3 {
    credential_type     = "secret"
    access_key          = "access_key"
    secret_key          = "secret_key"
    s3_force_path_style = false
    s3_url              = "http://10.90.78.23"
  }
}

#resource "spectrocloud_backup_storage_location" "bsl2" {
#  name        = "tenant-dev-1"
#  context     = "tenant"
#  is_default  = false
#  region      = "us-east-2"
#  bucket_name = "tenant-backup-2"
#  s3 {
#    credential_type     = "sts"
#    arn                 = "arn_role"
#    external_id         = "external_id"
#    s3_force_path_style = false
#    #s3_url             = "http://10.90.78.23"
#  }
#}