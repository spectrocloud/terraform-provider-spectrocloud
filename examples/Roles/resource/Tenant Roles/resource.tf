# Tenant-scoped custom role, composed from existing permissions looked up by name.
#
# Day-2 mutability: nothing on this resource is ForceNew - name, type, and permissions all
# update in place.
#
# Attributes:
#   name        - Required.
#   type        - Optional, default "project". Allowed: "project", "tenant", "resource".
#   permissions - Required. Set of permission ID strings.

variable "perms" {
  type    = list(string)
  default = ["User", "Team", "Role"]
}

data "spectrocloud_permission" "app_permissions" {
  for_each = toset(var.perms)
  name     = each.key
  scope    = "tenant"
}

resource "spectrocloud_role" "tenant_role" {
  name        = "Test Tenant Role"
  type        = "tenant"
  permissions = flatten([for p in data.spectrocloud_permission.app_permissions : p.permissions])
}

# terraform import spectrocloud_role.tenant_role "<role-uid>"
