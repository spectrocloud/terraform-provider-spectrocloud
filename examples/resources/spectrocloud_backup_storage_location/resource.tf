// S3 Back up location with secret credential type example
resource "spectrocloud_backup_storage_location" "bsl-s3" {
  name              = "project-dev-bsl"
  context           = "project"
  location_provider = "aws"
  is_default        = false
  region            = "us-east-1"
  bucket_name       = "project-backup-2"
  s3 {
    credential_type     = "secret"
    access_key          = "access_key"
    secret_key          = "secret_key"
    s3_force_path_style = false
    s3_url              = "http://10.90.78.23"
  }
}

// Minio Back up location with secret credential type example
resource "spectrocloud_backup_storage_location" "bsl-minio" {
  name              = "project-dev-minio-bsl"
  context           = "project"
  location_provider = "minio"
  is_default        = false
  region            = "us-east-2"
  bucket_name       = "project-backup-2"
  s3 {
    credential_type     = "secret"
    access_key          = "access_key"
    secret_key          = "secret_key"
    s3_force_path_style = false
    s3_url              = "http://10.90.78.23"
  }
}

// GCP Back up location example
resource "spectrocloud_backup_storage_location" "bsl-gcp" {
  name              = "project-dev-gcp"
  context           = "project"
  location_provider = "gcp"
  is_default        = false
  bucket_name       = "project-backup-2"
  gcp_storage_config {
    project_id           = "test_id"
    gcp_json_credentials = "test test"
  }
}

// Azure Back up location example
resource "spectrocloud_backup_storage_location" "bsl-azure" {
  name              = "project-dev-bsl-azure"
  context           = "project"
  location_provider = "azure"
  is_default        = false
  azure_storage_config {
    container_name      = "test-storage-container"
    storage_name        = "test-backup-storage"
    stock_keeping_unit  = "test1"
    resource_group      = "test-group"
    azure_tenant_id     = "test-tenant-id"
    azure_client_id     = "test-client-id"
    subscription_id     = "test-sub-id"
    azure_client_secret = "test-client-service"
  }
}

// S3 Back up location with STS credential type
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