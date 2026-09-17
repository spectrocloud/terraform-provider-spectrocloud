# Looks up a single pack by name/version/type - a simpler alternative to spectrocloud_pack.
#
# Lookup keys:
#   name         - Required.
#   version      - Optional. Defaults to "1.0.0" when omitted.
#   context      - Optional, default "project". Allowed: "system", "project", "tenant".
#   registry_uid - Optional, but effectively required whenever `type` is not "manifest" (the
#                  read fails without it in that case).
#   type         - Required. Allowed: "helm", "manifest", "container", "operator-instance".
data "spectrocloud_pack_simple" "example" {
  name         = "nginx-pack"
  version      = "1.2.3"
  context      = "project"
  registry_uid = "5ee9c5adc172449eeb9c30cf"
  type         = "helm"
}

# Output pack details (Computed):
#   pack_id
#   pack_version
#   pack_values - Stringified YAML pack configuration.
output "pack_id" {
  value = data.spectrocloud_pack_simple.example.id
}

output "pack_version" {
  value = data.spectrocloud_pack_simple.example.version
}

output "pack_values" {
  value = data.spectrocloud_pack_simple.example.values
}
