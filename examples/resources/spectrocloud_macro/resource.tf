resource "spectrocloud_macro" "project_macro" {
  name    = "project1"
  value   = "project_val2"
  project = "Default"
}

resource "spectrocloud_macro" "tenant_macro" {
  name  = "tenant1"
  value = "tenant_val1"
}