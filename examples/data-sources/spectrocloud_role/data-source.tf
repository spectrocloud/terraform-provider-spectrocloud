# Looks up a role by name or by ID.
data "spectrocloud_role" "example" {
  # Lookup key, optional (conflicts with `id`), also Computed.
  name = "admin-role"
  # Lookup key, optional (conflicts with `name`), also Computed.
  # id = "63d48062b3a0c92a6f230112"
}

# Computed.
output "role_id" {
  value = data.spectrocloud_role.example.id
}

# Computed. Set of permission ID strings granted by this role.
output "role_permissions" {
  value = data.spectrocloud_role.example.permissions
}
