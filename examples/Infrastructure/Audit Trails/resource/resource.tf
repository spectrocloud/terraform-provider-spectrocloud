# A tenant may have one CloudWatch audit trail and one Splunk audit trail configured concurrently.
# `type` picks which one this resource manages and is ForceNew - changing it recreates the sink
# under a new type rather than converting it in place.
#
# Attributes:
#   name - Required. Updatable in place.
#   type - Required, ForceNew. Allowed: "cloudwatch", "splunk".
#
# cloudwatch block (Required when type = "cloudwatch", at most one):
#   group           - Required. CloudWatch log group name.
#   region          - Required. AWS region.
#   stream          - Optional. CloudWatch log stream name.
#   credential_type - Optional, default "secret", ForceNew. Allowed: "secret", "sts".
#   access_key      - Required when credential_type = "secret", sensitive.
#   secret_key      - Required when credential_type = "secret", sensitive.
#   arn             - Required when credential_type = "sts". The IAM role ARN to assume.
#   external_id     - Optional, sensitive. Used alongside `arn` for STS role assumption when
#                     credential_type = "sts".
#   partition       - Optional, default "aws". Allowed: "aws", "aws-us-gov".
#
# splunk block (Required when type = "splunk", at most one):
#   hec_url - Required.
#   token   - Required, sensitive.
#   index   - Optional. Defaults to the token's own default index.
#   source  - Optional. Defaults to the token's own default source.
#
# splunk.tls_config block (Optional, at most one):
#   ca_cert_base64       - Optional. Base64-encoded CA certificate, for a self-signed Splunk
#                          endpoint.
#   insecure_skip_verify - Optional, default true (skips certificate verification by default -
#                          set to false to verify the Splunk endpoint's TLS certificate).
#                          `tls_verification` is the Computed inverse of this value - it is
#                          read-only and populated automatically.

# --- CloudWatch, using long-lived AWS access/secret keys ---
resource "spectrocloud_audit_trail" "cloudwatch_secret" {
  name = "cloudwatch-secret-example"
  type = "cloudwatch"

  cloudwatch {
    group  = "dev-hubble-audits"
    region = "us-east-1"
    # stream = "REPLACE_ME"

    credential_type = "secret"
    access_key      = var.aws_access_key
    secret_key      = var.aws_secret_key
    partition       = "aws"
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
    arn             = var.aws_sts_role_arn
    external_id     = var.aws_external_id
    partition       = "aws"
  }
}

# --- Splunk HTTP Event Collector (HEC) ---
resource "spectrocloud_audit_trail" "splunk" {
  name = "splunk-example"
  type = "splunk"

  splunk {
    hec_url = "https://http-inputs-example.splunkcloud.com:443"
    token   = var.splunk_hec_token
    index   = "main"
    source  = "palette"

    tls_config {
      # ca_cert_base64 = "REPLACE_ME"
      insecure_skip_verify = false
    }
  }
}

# Import example:
# terraform import spectrocloud_audit_trail.cloudwatch_secret "audit_trail_uid_here"
