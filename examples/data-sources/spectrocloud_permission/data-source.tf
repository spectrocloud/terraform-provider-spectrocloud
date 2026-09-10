# Looks up a permission definition by name and scope.
data "spectrocloud_permission" "example" {
  # Required lookup key. Example: "App Deployment", "Cluster Management", "User Access".
  name = "App Deployment"
  # Optional lookup key, default "project". Allowed: "project", "tenant", "resource".
  scope = "project"
}

# Computed. All attributes on this data source, for reference.
output "permission_details" {
  value = data.spectrocloud_permission.example
}

# Computed.
output "permission_id" {
  value = data.spectrocloud_permission.example.id
}

# Computed. List of individual permission strings granted by this permission name/scope.
output "permission_list" {
  value = data.spectrocloud_permission.example.permissions
}
