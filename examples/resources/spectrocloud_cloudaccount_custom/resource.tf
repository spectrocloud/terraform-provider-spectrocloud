data "spectrocloud_private_cloud_gateway" "nutanix_pcg" {
  name = "test-pcg"
}

resource "spectrocloud_cloudaccount_custom" "cloud_account" {
  name = "test-custom-cloud-account"
  cloud = "nutanix"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.nutanix_pcg.id
  context = "tenant"
  credentials = {
    "NUTANIX_USER" = "test_user",
    "NUTANIX_PASSWORD" = sensitive("test123"),
    "NUTANIX_ENDPOINT" = "10.12.11.22",
    "NUTANIX_PORT" = "8998",
    "NUTANIX_INSECURE" = "yes"
  }
}