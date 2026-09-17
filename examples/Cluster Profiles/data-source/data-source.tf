# Looks up a cluster profile by name (with optional version) or by ID.

# Retrieve details of a specific cluster profile using name.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   version - Optional, also Computed. Defaults to "1.0.0" when omitted.
#   context - Optional, default "project". Allowed: "project", "tenant", "system".
data "spectrocloud_cluster_profile" "example" {
  name    = "example-cluster-profile"
  version = "1.0.0"
  context = "project"
}

# Retrieve details of a cluster profile using ID.
#
# Lookup keys:
#   id - Exactly one of `id`/`name` required, also Computed.
data "spectrocloud_cluster_profile" "by_id" {
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
