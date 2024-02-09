data "spectrocloud_private_cloud_gateway" "nutanix_pcg" {
  name = "sh-nut-feb5a"
}

resource "spectrocloud_custom_cloud_account" "cloud_account" {
  name = "test-custom-cloud-account"
  cloud = "nutanix"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.nutanix_pcg.id
  context = "tenant"
  credentials = {
    "NUTANIX_USER" = "arvind1",
    "NUTANIX_PASSWORD" = sensitive("Pni1NN985QF$"),
    "NUTANIX_ENDPOINT" = "10.11.136.220",
    "NUTANIX_PORT" = "9443",
    "NUTANIX_INSECURE" = "yes"
  }
}