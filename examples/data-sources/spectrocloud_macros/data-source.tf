data "spectrocloud_macros" "macros" {
  context = "project"
  #  macro_name = "MACRO_PROJECT_PODCIDR"
}

output "macro_eg_name" {
  value = data.spectrocloud_macros.macros.macro_value
}

output "macros_map" {
  value = data.spectrocloud_macros.macros.macros_map
}
