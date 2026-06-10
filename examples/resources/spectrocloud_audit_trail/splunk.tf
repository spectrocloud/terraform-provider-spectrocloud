resource "spectrocloud_audit_trail" "splunk" {
  name = "test-tf"
  type = "splunk"

  splunk {
    hec_url = "https://http-inputs-example.splunkcloud.com:443"
    token   = var.splunk_hec_token
    index   = "main"
    source  = "palette"

    tls_config {
      tls_verification = true
    }
  }
}
