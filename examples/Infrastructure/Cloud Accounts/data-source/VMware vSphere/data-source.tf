# Looks up an existing vSphere cloud account registered in Palette, by name or by ID.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   context - Optional. Allowed: "project", "tenant", "" (default). Required only to
#             disambiguate when more than one account shares the same `name` across scopes.
data "spectrocloud_cloudaccount_vsphere" "example" {
  name    = "example-vsphere-account"
  context = "project"
}

# Alternative: look up by ID instead of name.
#
# Lookup keys:
#   id - Exactly one of `id`/`name` required, also Computed.
data "spectrocloud_cloudaccount_vsphere" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000"
}

output "vsphere_account_id" {
  value = data.spectrocloud_cloudaccount_vsphere.example.id
}

output "vsphere_account_name" {
  value = data.spectrocloud_cloudaccount_vsphere.example.name
}

# Computed.
output "private_cloud_gateway_id" {
  value = data.spectrocloud_cloudaccount_vsphere.example.private_cloud_gateway_id
}
