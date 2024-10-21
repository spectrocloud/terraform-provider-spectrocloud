data "spectrocloud_private_cloud_gateway" "gateway" {
  name = "test-vm-pcg"
}

resource "spectrocloud_privatecloudgateway_dns_map" "dns_map_test" {
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.gateway.id
  search_domain_name       = "test1.spectro.com"
  data_center              = "DataCenterTest"
  network                  = "TEST-VM-NETWORK"
}
