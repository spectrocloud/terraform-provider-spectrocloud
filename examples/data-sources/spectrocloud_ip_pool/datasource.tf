data "spectrocloud_private_cloud_gateway" "gateway" {
  name = "pcg-benoitcamp-vcenter"
}

data "spectrocloud_ippool" "ippool" {
   name                 = "IP Pool Jesse"
   private_cloud_gateway_id      = data.spectrocloud_private_cloud_gateway.gateway.id
}

variable "private_cloud_gateway_id" {
  type = string
}

output "same" {
  value = data.spectrocloud_ippool.ippool.id
}
