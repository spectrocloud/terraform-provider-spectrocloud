# Looks up an existing custom cloud account (e.g. Nutanix) registered in Palette, by name or by
# ID.
data "spectrocloud_cloudaccount_custom" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "example-nutanix-account"
  # Required. The custom cloud provider name, e.g. "nutanix".
  cloud = "nutanix"
  # Optional. Allowed: "project", "tenant", "" (default). Required only to disambiguate when
  # more than one account shares the same `name` across scopes.
  context = "project"
}

# Alternative: look up by ID instead of name
data "spectrocloud_cloudaccount_custom" "by_id" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  id = "123e4567-e89b-12d3-a456-426614174000"
  # Still required even when looking up by id.
  cloud = "nutanix"
}

output "custom_account_id" {
  value = data.spectrocloud_cloudaccount_custom.example.id
}

output "custom_account_name" {
  value = data.spectrocloud_cloudaccount_custom.example.name
}
