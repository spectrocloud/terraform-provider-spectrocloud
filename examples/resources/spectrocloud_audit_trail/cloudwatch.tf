# CloudWatch audit trail using secret credentials.
# Maps to POST /v1/tenants/{tenantUID}/assets/dataSinks with type "cloudwatch".
# A tenant may have one CloudWatch audit trail and one Splunk audit trail concurrently.
resource "spectrocloud_audit_trail" "cloudwatch" {
  name = "rag"
  type = "cloudwatch"

  cloudwatch {
    group  = "dev-hubble-audits"
    region = "us-east-1"

    credential_type = "secret"
    access_key      = var.aws_access_key
    secret_key      = var.aws_secret_key
    partition       = "aws"
  }
}

resource "spectrocloud_audit_trail" "cloudwatch_sts" {
  name = "rag-sts"
  type = "cloudwatch"

  cloudwatch {
    group  = "dev-hubble-audits"
    region = "us-east-1"

    credential_type = "sts"
    arn             = var.aws_sts_role_arn
    external_id     = var.aws_external_id
    partition       = "aws"
  }
}
