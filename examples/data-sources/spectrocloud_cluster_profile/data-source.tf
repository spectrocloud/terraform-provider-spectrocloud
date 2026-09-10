# Looks up a cluster profile by name (with optional version) or by ID.
#
# Retrieve details of a specific cluster profile using name
data "spectrocloud_cluster_profile" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "example-cluster-profile"
  # Optional lookup key, also Computed. Defaults to "1.0.0" when omitted.
  version = "1.0.0"
  # Optional lookup key, default "project". Allowed: "project", "tenant", "system".
  context = "project"
}

# Retrieve details of a cluster profile using ID
data "spectrocloud_cluster_profile" "by_id" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  id = "123e4567e89ba426614174000"
}

# Output cluster profile details
output "cluster_profile_id" {
  value = data.spectrocloud_cluster_profile.example.id
}

output "cluster_profile_version" {
  value = data.spectrocloud_cluster_profile.example.version
}

# Computed. List of packs published on this profile version - each with type, registry_uid,
# uid, name, tag, values, and any manifest entries (uid/name/content).
output "cluster_profile_packs" {
  value = data.spectrocloud_cluster_profile.example.pack
}
