# Retrieve details of a specific cluster profile using name
data "spectrocloud_cluster_profile" "example" {
  name    = "example-cluster-profile"  # Required if 'id' is not provided
  version = "1.0.0"                    # Optional: Version of the cluster profile
  context = "project"                   # Optional: Allowed values: "project", "tenant", "system" (Defaults to "project")
}

# Retrieve details of a cluster profile using ID
data "spectrocloud_cluster_profile" "by_id" {
  id = "123e4567e89ba426614174000"  # Required if 'name' is not provided
}

# Output cluster profile details
output "cluster_profile_id" {
  value = data.spectrocloud_cluster_profile.example.id
}

output "cluster_profile_version" {
  value = data.spectrocloud_cluster_profile.example.version
}

# Retrieve packs associated with a cluster profile
output "cluster_profile_packs" {
  value = data.spectrocloud_cluster_profile.example.pack
}