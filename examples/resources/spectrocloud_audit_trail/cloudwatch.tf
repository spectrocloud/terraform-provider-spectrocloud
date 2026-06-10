resource "spectrocloud_audit_trail" "cloudwatch" {
  name = "test-tf"
  type = "cloudwatch"

  cloudwatch {
    group  = "logs"
    region = "us-east-1"
    stream = "test"

    credential_type = "secret"
    access_key      = var.aws_access_key
    secret_key      = var.aws_secret_key
  }
}

resource "spectrocloud_audit_trail" "cloudwatch_sts" {
  name = "test-tf-sts"
  type = "cloudwatch"

  cloudwatch {
    group  = "logs"
    region = "us-east-1"

    credential_type = "sts"
    arn             = "arn:aws:iam::123456789012:role/SpectroCloudRole"
    external_id     = var.external_id
  }
}
