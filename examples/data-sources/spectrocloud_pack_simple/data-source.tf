# Retrieve details of a specific pack
data "spectrocloud_pack_simple" "example" {
  name         = "nginx-pack"   # Required: Name of the pack
  version      = "1.2.3"        # Optional: Version of the pack
  context      = "project"      # Optional: Allowed values: "system", "project", "tenant". Defaults to "project".
  registry_uid = "5ee9c5adc172449eeb9c30cf"  # Optional: Unique identifier of the registry
  type         = "helm"         # Required: Allowed values: "helm", "manifest", "container", "operator-instance"
}

# Output pack details
output "pack_id" {
  value = data.spectrocloud_pack_simple.example.id
}

output "pack_version" {
  value = data.spectrocloud_pack_simple.example.version
}

output "pack_values" {
  value = data.spectrocloud_pack_simple.example.values
}
