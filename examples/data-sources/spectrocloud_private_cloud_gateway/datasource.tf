data "spectrocloud_private_cloud_gateway" "gateway" {
  name = "pcg-benoitcamp-vcenter"
}

output "same" {
  value = data.spectrocloud_private_cloud_gateway.gateway.id
}
