# Attributes:
#   name                     - Required. Updates in place.
#   cloud                    - Required, ForceNew. The custom cloud provider name (e.g.
#                              "nutanix").
#   private_cloud_gateway_id - Required, ForceNew. Connects this account to the underlying
#                              infrastructure through a PCG.
#   context                  - Optional, default "project", ForceNew. Allowed: "project",
#                              "tenant".
#   credentials              - Optional, sensitive. Provider-specific credential key/value pairs
#                              - the required keys depend on the custom cloud provider (shown
#                              here: Nutanix). Updates in place.

data "spectrocloud_private_cloud_gateway" "nutanix_pcg" {
  name = "test-pcg"
}

resource "spectrocloud_cloudaccount_custom" "cloud_account" {
  name                     = "test-custom-cloud-account"
  cloud                    = "nutanix"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.nutanix_pcg.id
  context                  = "tenant"
  credentials = {
    "NUTANIX_USER"     = "test_user",
    "NUTANIX_PASSWORD" = var.nutanix_password,
    "NUTANIX_ENDPOINT" = "1.2.3.4",
    "NUTANIX_PORT"     = "8998",
    "NUTANIX_INSECURE" = "yes"
  }
}
