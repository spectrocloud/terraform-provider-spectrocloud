# Retrieve details of a specific role by name
data "spectrocloud_role" "example" {
  name = "admin-role"
}

# Output role ID
output "role_id" {
  value = data.spectrocloud_role.example.id
}

# Output permissions associated with the role
output "role_permissions" {
  value = data.spectrocloud_role.example.permissions
}
