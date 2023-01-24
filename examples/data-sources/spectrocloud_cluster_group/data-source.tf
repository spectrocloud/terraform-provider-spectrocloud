data "spectrocloud_cluster_group" "beehive" {
  name = "beehive"
  context = "system"
}

output "out_beehive" {
  value = data.spectrocloud_cluster_group.beehive.id
}

data "spectrocloud_cluster_group" "tenant_cl" {
  name = "tenant_cl"
  context = "system"
}

output "out_tenant_cl" {
  value = data.spectrocloud_cluster_group.tenant_cl.id
}