// set permission with data source role
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

// set permission with data source permission

variable "perms" {
  type    = list(string)
  default = ["API Key", "Audit", "Filter"]
}

data "spectrocloud_permission" "app_permissions" {
  for_each = toset(var.perms)
  name     = each.key
  scope = "tenant"
}

resource "spectrocloud_role" "custom_role_permission" {
  name        = "Test Cluster Role With Custom Permission"
  type        = "tenant"
  permissions = flatten([for p in data.spectrocloud_permission.app_permissions : p.permissions])
}