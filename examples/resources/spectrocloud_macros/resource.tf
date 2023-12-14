resource "spectrocloud_macros" "project_macro" {
  macros = {
    "tfpjt"="test_value",
    "mutiplepjt"="macros_value",
  }
  context = "project"
}

#resource "spectrocloud_macro" "tenant_macro" {
#  name  = "tenant1"
#  value = "tenant_val1"
#}