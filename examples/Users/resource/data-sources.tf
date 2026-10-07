
data "spectrocloud_project" "default" {
  name = "Default"
}

data "spectrocloud_project" "ranjith" {
  name = "ranjith"
}

data "spectrocloud_role" "app_roles" {
  for_each = toset(var.app_role_var)
  name     = each.key
}

data "spectrocloud_role" "tenant_roles" {
  for_each = toset(var.tenant_role_var)
  name     = each.key
}

data "spectrocloud_workspace" "workspace" {
  name = "test-ws-tf"
}

data "spectrocloud_workspace" "workspace2" {
  name = "test-ws-2"
}

data "spectrocloud_role" "workspace_roles" {
  for_each = toset(var.workspace_role_var)
  name     = each.key
}

data "spectrocloud_filter" "filter" {
  name = "test-tf"
}

data "spectrocloud_role" "resource_roles" {
  for_each = toset(var.resource_role_var)
  name     = each.key
}

data "spectrocloud_role" "resource_roles_editor" {
  for_each = toset(var.resource_role_editor_var)
  name     = each.key
}

data "spectrocloud_team" "team2" {
  name = "team2"
}