# Looks up a permission definition by name and scope.
#
# Lookup keys:
#   name  - Required. Example: "App Deployment", "Cluster Management", "User Access".
#   scope - Optional, default "project". Allowed: "project", "tenant", "resource".
data "spectrocloud_permission" "example" {
  name  = "App Deployment"
  scope = "project"
}

# Computed outputs:
#   permission_details - All attributes on this data source, for reference.
#   permission_id      - ID of the permission.
#   permission_list    - List of individual permission strings granted by this permission
#                        name/scope.
output "permission_details" {
  value = data.spectrocloud_permission.example
}

output "permission_id" {
  value = data.spectrocloud_permission.example.id
}

output "permission_list" {
  value = data.spectrocloud_permission.example.permissions
}
