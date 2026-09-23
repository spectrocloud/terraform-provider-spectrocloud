# Required for static IP placement.
resource "spectrocloud_privatecloudgateway_ippool" "ippool" {
  count                    = var.deploy-vmware-static ? 1 : 0
  gateway                  = var.network_gateway
  name                     = "vsphere-vmware-ippool"
  network_type             = "range"
  prefix                   = var.network_prefix
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg[0].id
  ip_start_range           = var.ip_range_start
  ip_end_range             = var.ip_range_end
  nameserver_addresses     = var.nameserver_addr
}
