# Looks up macros/service-variable-output values defined at the project or tenant level.
#
# Lookup keys:
#   context    - Optional, default "tenant". Allowed: "project", "tenant".
#   macro_name - Optional. When set, `macro_value` below is populated with this macro's value.
data "spectrocloud_macros" "macros" {
  context    = "project"
  macro_name = "MACRO_PROJECT_PODCIDR"
}

# Computed outputs:
#   macro_value - Populated only when macro_name above is set.
#   macros_map  - Every macro in this context, as a name -> value map.
#   macros_id   - UID of the project or tenant this macro set belongs to.
output "macro_value" {
  value = data.spectrocloud_macros.macros.macro_value
}

output "macros_map" {
  value = data.spectrocloud_macros.macros.macros_map
}

output "macros_id" {
  value = data.spectrocloud_macros.macros.id
}
