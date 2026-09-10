# Looks up an existing vSphere cloud account registered in Palette, by name or by ID.
data "spectrocloud_cloudaccount_vsphere" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "example-vsphere-account"
  # Optional. Allowed: "project", "tenant", "" (default). Required only to disambiguate when
  # more than one account shares the same `name` across scopes.
  context = "project"
}

# Alternative: look up by ID instead of name
data "spectrocloud_cloudaccount_vsphere" "by_id" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
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
