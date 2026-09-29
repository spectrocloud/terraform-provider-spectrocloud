# Looks up an existing custom cloud account (e.g. Nutanix) registered in Palette, by name or by
# ID.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   cloud   - Required. The custom cloud provider name, e.g. "nutanix".
#   context - Optional. Allowed: "project", "tenant", "" (default). Required only to
#             disambiguate when more than one account shares the same `name` across scopes.
data "spectrocloud_cloudaccount_custom" "example" {
  name    = "example-nutanix-account"
  cloud   = "nutanix"
  context = "project"
}

# Alternative: look up by ID instead of name.
#
# Lookup keys:
#   id    - Exactly one of `id`/`name` required, also Computed.
#   cloud - Required. Still required even when looking up by id.
data "spectrocloud_cloudaccount_custom" "by_id" {
  id    = "123e4567-e89b-12d3-a456-426614174000"
  cloud = "nutanix"
}

output "custom_account_id" {
  value = data.spectrocloud_cloudaccount_custom.example.id
}

output "custom_account_name" {
  value = data.spectrocloud_cloudaccount_custom.example.name
}
