# Looks up a cluster group by name within a given context.
#
# Lookup keys:
#   name    - Required.
#   context - Optional, default "tenant". Allowed: "system", "tenant", "project".
data "spectrocloud_cluster_group" "example_group" {
  name    = "my-cluster-group"
  context = "tenant"
}

# Computed (also echoes the lookup key on success).
output "cluster_group_id" {
  value = data.spectrocloud_cluster_group.example_group.id
}

output "cluster_group_name" {
  value = data.spectrocloud_cluster_group.example_group.name
}
