# Resource-scoped custom role (applies to a specific resource instance rather than an entire
# project or tenant), composed from existing resource-level permissions looked up by name.
#
# Day-2 mutability: nothing on this resource is ForceNew - name, type, and permissions all
# update in place.
#
# Attributes:
#   name        - Required.
#   type        - Optional, default "project". Allowed: "project", "tenant", "resource".
#   permissions - Required. Set of permission ID strings.

variable "resource_perms" {
  type    = list(string)
  default = ["Cluster View", "Cluster Edit"]
}

data "spectrocloud_permission" "resource_permissions" {
  for_each = toset(var.resource_perms)
  name     = each.key
  scope    = "resource"
}

resource "spectrocloud_role" "resource_role" {
  name        = "Test Resource Role"
  type        = "resource"
  permissions = flatten([for p in data.spectrocloud_permission.resource_permissions : p.permissions])
}

# terraform import spectrocloud_role.resource_role "<role-uid>"
