variable "roles" {
  type    = list(string)
  default = ["Cluster Admin", "Cluster Profile Editor"]
}

# Data source loop to retrieve multiple roles
data "spectrocloud_role" "roles" {
  for_each = toset(var.roles)
  name     = each.key
}

resource "spectrocloud_role" "custom_role" {
  name        = "Test Cluster Role"
  type        = "project"
  permissions = flatten([for role in data.spectrocloud_role.roles : role.permissions])
}