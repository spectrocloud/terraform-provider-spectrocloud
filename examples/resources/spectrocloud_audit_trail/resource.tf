# A tenant may have one CloudWatch audit trail and one Splunk audit trail configured concurrently.
# `type` picks which one this resource manages and is ForceNew - changing it recreates the sink
# under a new type rather than converting it in place.

# --- CloudWatch, using long-lived AWS access/secret keys ---
resource "spectrocloud_audit_trail" "cloudwatch_secret" {
  # Required. Updatable in place.
  name = "cloudwatch-secret-example"
  # Required, ForceNew. Allowed: "cloudwatch", "splunk".
  type = "cloudwatch"

  # Required when type = "cloudwatch". At most one block.
  cloudwatch {
    group  = "dev-hubble-audits" # Required. CloudWatch log group name.
    region = "us-east-1"         # Required. AWS region.
    # stream = "REPLACE_ME"      # Optional. CloudWatch log stream name.

    # Optional, default "secret", ForceNew. Allowed: "secret", "sts".
    credential_type = "secret"
    # Required when credential_type = "secret". Both sensitive.
    access_key = var.aws_access_key
    secret_key = var.aws_secret_key
    # Optional, default "aws". Allowed: "aws", "aws-us-gov".
    partition = "aws"
  }
}

# --- CloudWatch, using an assumable IAM role (STS) instead of long-lived keys ---
resource "spectrocloud_audit_trail" "cloudwatch_sts" {
  name = "cloudwatch-sts-example"
  type = "cloudwatch"

  cloudwatch {
    group  = "dev-hubble-audits"
    region = "us-east-1"

    credential_type = "sts"
    # Required when credential_type = "sts".
    arn = var.aws_sts_role_arn
    # Optional, sensitive. Used alongside the role ARN for STS role assumption.
    external_id = var.aws_external_id
    partition   = "aws"
  }
}

# --- Splunk HTTP Event Collector (HEC) ---
resource "spectrocloud_audit_trail" "splunk" {
  name = "splunk-example"
  type = "splunk"

  # Required when type = "splunk". At most one block.
  splunk {
    hec_url = "https://http-inputs-example.splunkcloud.com:443" # Required.
    token   = var.splunk_hec_token                              # Required, sensitive.
    index   = "main"                                            # Optional. Defaults to the token's own default index.
    source  = "palette"                                         # Optional. Defaults to the token's own default source.

    # Optional, at most one block.
    tls_config {
      # Optional. Base64-encoded CA certificate, for a self-signed Splunk endpoint.
      # ca_cert_base64 = "REPLACE_ME"
      # Optional, default true (skips certificate verification by default - set to false to
      # verify the Splunk endpoint's TLS certificate). `tls_verification` is the Computed
      # inverse of this value - it is read-only and populated automatically.
      insecure_skip_verify = false
    }
  }
}

# Import example:
# terraform import spectrocloud_audit_trail.cloudwatch_secret "audit_trail_uid_here"
