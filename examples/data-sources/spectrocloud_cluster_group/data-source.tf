# Retrieve details of a specific cluster group
data "spectrocloud_cluster_group" "example_group" {
  name    = "my-cluster-group" # Specify the name of the cluster group
  context = "tenant"           # Context can be "system", "tenant", or "project"
}

# Output the retrieved cluster group details
output "cluster_group_name" {
  value = data.spectrocloud_cluster_group.example_group.name
}
