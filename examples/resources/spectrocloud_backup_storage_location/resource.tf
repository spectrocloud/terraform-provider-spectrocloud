// S3 Backup location with secret credential type example
resource "spectrocloud_backup_storage_location" "bsl_s3" {
  name              = "project-dev-bsl-s3"
  context           = "project"
  location_provider = "aws"
  is_default        = false
  region            = "us-east-1"
  bucket_name       = "project-backup-bucket-s3"
  s3 {
    credential_type     = "secret"
    access_key          = "test-access-key-s3"
    secret_key          = "test-secret-key-s3"
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-1.amazonaws.com"
  }
}

// Minio Backup location with secret credential type example
resource "spectrocloud_backup_storage_location" "bsl_minio" {
  name              = "project-dev-minio-bsl"
  context           = "project"
  location_provider = "minio"
  is_default        = false
  region            = "us-east-2"
  bucket_name       = "project-backup-bucket-minio"
  s3 {
    credential_type     = "secret"
    access_key          = "test-access-key-minio"
    secret_key          = "test-secret-key-minio"
    s3_force_path_style = true
    s3_url              = "http://10.90.78.23"
  }
}

// GCP Backup location example
resource "spectrocloud_backup_storage_location" "bsl_gcp" {
  name              = "project-dev-gcp"
  context           = "project"
  location_provider = "gcp"
  is_default        = false
  bucket_name       = "project-backup-bucket-gcp"
  gcp_storage_config {
    project_id           = "test-gcp-project-id"
    gcp_json_credentials = <<EOT
                          {
                            "type": "service_account",
                            "project_id": "test-gcp-project-id",
                            "private_key_id": "test-private-key-id",
                            "private_key": "-----BEGIN PRIVATE KEY-----\nPRIVATE_KEY_HERE\n-----END PRIVATE KEY-----\n",
                            "client_email": "test-client-email@gcp-project-id.iam.gserviceaccount.com",
                            "client_id": "test-client-id",
                            "auth_uri": "https://accounts.google.com/o/oauth2/auth",
                            "token_uri": "https://oauth2.googleapis.com/token",
                            "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
                            "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test-client-email%40gcp-project-id.iam.gserviceaccount.com"
                          }
                          EOT
  }
}

// Azure Backup location example
resource "spectrocloud_backup_storage_location" "bsl_azure" {
  name              = "project-dev-azure-bsl"
  context           = "project"
  location_provider = "azure"
  is_default        = false
  azure_storage_config {
    container_name      = "test-container"
    storage_name        = "test-storage"
    stock_keeping_unit  = "Standard_LRS"
    resource_group      = "test-resource-group"
    azure_tenant_id     = "test-azure-tenant-id"
    azure_client_id     = "test-azure-client-id"
    subscription_id     = "test-azure-subscription-id"
    azure_client_secret = "test-azure-client-secret"
  }
}

// S3 Backup location with STS credential type
resource "spectrocloud_backup_storage_location" "bsl_sts" {
  name        = "tenant-dev-1"
  context     = "tenant"
  is_default  = false
  region      = "us-east-2"
  bucket_name = "tenant-backup-bucket-sts"
  s3 {
    credential_type     = "sts"
    arn                 = "arn:aws:iam::123456789012:role/TestRole"
    external_id         = "test-external-id"
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-2.amazonaws.com"
  }
}