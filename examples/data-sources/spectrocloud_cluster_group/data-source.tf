# Looks up a cluster group by name within a given context.
data "spectrocloud_cluster_group" "example_group" {
  # Required lookup key.
  name = "my-cluster-group"
  # Optional lookup key, default "tenant". Allowed: "system", "tenant", "project".
  context = "tenant"
}

# Computed (also echoes the lookup key on success).
output "cluster_group_id" {
  value = data.spectrocloud_cluster_group.example_group.id
}

output "cluster_group_name" {
  value = data.spectrocloud_cluster_group.example_group.name
}
