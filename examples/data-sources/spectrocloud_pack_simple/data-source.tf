# Looks up a single pack by name/version/type - a simpler alternative to spectrocloud_pack.
data "spectrocloud_pack_simple" "example" {
  # Required lookup key.
  name = "nginx-pack"
  # Optional lookup key. Defaults to "1.0.0" when omitted.
  version = "1.2.3"
  # Optional lookup key, default "project". Allowed: "system", "project", "tenant".
  context = "project"
  # Optional lookup key - but effectively required whenever `type` is not "manifest" (the read
  # fails without it in that case).
  registry_uid = "5ee9c5adc172449eeb9c30cf"
  # Required lookup key. Allowed: "helm", "manifest", "container", "operator-instance".
  type = "helm"
}

# Output pack details
output "pack_id" {
  value = data.spectrocloud_pack_simple.example.id
}

output "pack_version" {
  value = data.spectrocloud_pack_simple.example.version
}

# Computed. Stringified YAML pack configuration.
output "pack_values" {
  value = data.spectrocloud_pack_simple.example.values
}
