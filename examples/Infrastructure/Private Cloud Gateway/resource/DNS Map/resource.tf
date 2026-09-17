data "spectrocloud_private_cloud_gateway" "gateway" {
  name = "test-vm-pcg"
}

# Day-2 mutability: `private_cloud_gateway_id`, `data_center`, and `network` are ForceNew -
# changing any of them recreates the mapping. `search_domain_name` updates in place.
#
# Attributes:
#   private_cloud_gateway_id - Required, ForceNew. ID of the private cloud gateway this DNS
#                              mapping applies to.
#   search_domain_name       - Required. Domain name used for DNS search queries within the
#                              private cloud. Must be a valid domain name (e.g. "example.com").
#   data_center              - Required, ForceNew. The vSphere datacenter this mapping applies
#                              to, as it appears in vSphere.
#   network                  - Required, ForceNew. The vSphere network this mapping is bound to,
#                              as it appears in vSphere.
resource "spectrocloud_privatecloudgateway_dns_map" "dns_map_test" {
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.gateway.id
  search_domain_name       = "test1.spectro.com"
  data_center              = "DataCenterTest"
  network                  = "TEST-VM-NETWORK"
}
