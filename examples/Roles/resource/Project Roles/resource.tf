# Project-scoped custom role, composed from other roles' permission sets.
#
# Day-2 mutability: nothing on this resource is ForceNew - name, type, and permissions all
# update in place.
#
# Attributes:
#   name        - Required.
#   type        - Optional, default "project". Allowed: "project", "tenant", "resource".
#   permissions - Required. Set of permission ID strings.

variable "roles" {
  type    = list(string)
  default = ["Cluster Admin", "Cluster Profile Editor"]
}

# Data source loop to retrieve multiple existing roles, whose permissions we combine below.
data "spectrocloud_role" "roles" {
  for_each = toset(var.roles)
  name     = each.key
}

resource "spectrocloud_role" "project_role" {
  name        = "Test Project Role"
  type        = "project"
  permissions = flatten([for role in data.spectrocloud_role.roles : role.permissions])
}

# terraform import spectrocloud_role.project_role "<role-uid>"
