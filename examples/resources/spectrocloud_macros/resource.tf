resource "spectrocloud_macros" "project_macro" {
  macros = {
    "macro_project_1" = "val1",
    "macro_project_2" = "val2",
  }
  project = "Default"
}

resource "spectrocloud_macros" "tenant_macro" {
  macros = {
    "macro_tenant_1" = "tenant_val1",
    "macro_tenant_2" = "tenant_val2",
  }
}