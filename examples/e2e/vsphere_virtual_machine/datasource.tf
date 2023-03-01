data "spectrocloud_private_cloud_gateway" "gateway" {
  name = var.gateway_name
}

data "spectrocloud_ippool" "ippool" {
  name                     = var.ippool_name
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.gateway.id
}


data "spectrocloud_cluster_profile" "profile" {
  name    = "vmware-jben"
  version = "1.0.0"
  context = "project"
}