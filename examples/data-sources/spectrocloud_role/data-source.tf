data "spectrocloud_role" "role" {
  name = "Resource Cluster Admin"

  # (alternatively)
  # id =  "66fbea622947f81fb62294ac"
}

output "role_id" {
  value = data.spectrocloud_role.role.id
}

output "role_permissions" {
  value = data.spectrocloud_role.role.permissions
}