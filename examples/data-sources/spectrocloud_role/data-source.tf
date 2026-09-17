# Looks up a role by name or by ID.
#
# Lookup keys:
#   name - Optional (conflicts with `id`), also Computed.
#   id   - Optional (conflicts with `name`), also Computed.
data "spectrocloud_role" "example" {
  name = "admin-role"
  # id = "63d48062b3a0c92a6f230112"
}

# Computed outputs:
#   role_id
#   role_permissions - Set of permission ID strings granted by this role.
output "role_id" {
  value = data.spectrocloud_role.example.id
}

output "role_permissions" {
  value = data.spectrocloud_role.example.permissions
}
