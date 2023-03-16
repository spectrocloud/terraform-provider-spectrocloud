data "spectrocloud_cluster_group" "beehive" {
  name    = "beehive"
  context = "system"
}

output "out_beehive" {
  value = data.spectrocloud_cluster_group.beehive.id
}

data "spectrocloud_cluster_group" "tenant_cl" {
  name    = "sanfrancisco"
  context = "tenant"
}

output "out_tenant_cl" {
  value = data.spectrocloud_cluster_group.tenant_cl.id
}

data "spectrocloud_cluster_group" "project_sc" {
  name    = "cg-1"
  context = "project"
}

output "out_project_sc" {
  value = data.spectrocloud_cluster_group.project_sc.id
}