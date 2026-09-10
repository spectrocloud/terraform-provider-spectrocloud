# Looks up macros/service-variable-output values defined at the project or tenant level.
data "spectrocloud_macros" "macros" {
  # Optional lookup key, default "tenant". Allowed: "project", "tenant".
  context = "project"
  # Optional lookup key. When set, `macro_value` below is populated with this macro's value.
  macro_name = "MACRO_PROJECT_PODCIDR"
}

# Computed. Populated only when macro_name above is set.
output "macro_value" {
  value = data.spectrocloud_macros.macros.macro_value
}

# Computed. Every macro in this context, as a name -> value map.
output "macros_map" {
  value = data.spectrocloud_macros.macros.macros_map
}

# Computed. UID of the project or tenant this macro set belongs to.
output "macros_id" {
  value = data.spectrocloud_macros.macros.id
}
