resource "spectrocloud_macros" "project_macros" {
  name = "project1"
  value = "project_val2"
  project = "Default"
}

resource "spectrocloud_macros" "tenant_macros" {
  name = "tenant1"
  value = "tenant_val1"
}