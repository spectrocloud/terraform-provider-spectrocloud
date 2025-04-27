resource "spectrocloud_macros" "project_macro" {
  macros = {
    "project_macro_1" = "val1",
    "project_macro_2" = "val2",
  }
  context = "project"
}

resource "spectrocloud_macros" "tenant_macro" {
  macros = {
    "tenant_macro_1" = "tenant_val1",
    "tenant_macro_2" = "tenant_val2",
  }
}