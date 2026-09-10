data "spectrocloud_private_cloud_gateway" "nutanix_pcg" {
  name = "test-pcg"
}

resource "spectrocloud_cloudaccount_custom" "cloud_account" {
  # Required. Updates in place.
  name = "test-custom-cloud-account"
  # Required, ForceNew. The custom cloud provider name (e.g. "nutanix").
  cloud = "nutanix"
  # Required, ForceNew. Connects this account to the underlying infrastructure through a PCG.
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.nutanix_pcg.id
  # Optional, default "project", ForceNew. Allowed: "project", "tenant".
  context = "tenant"
  # Optional, sensitive. Provider-specific credential key/value pairs - the required keys
  # depend on the custom cloud provider (shown here: Nutanix). Updates in place.
  credentials = {
    "NUTANIX_USER"     = "test_user",
    "NUTANIX_PASSWORD" = sensitive("test123"),
    "NUTANIX_ENDPOINT" = "1.2.3.4",
    "NUTANIX_PORT"     = "8998",
    "NUTANIX_INSECURE" = "yes"
  }
}
