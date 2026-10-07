# Looks up an existing GCP cloud account registered in Palette, by name or by ID.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   context - Optional. Allowed: "project", "tenant", "" (default). Required only to
#             disambiguate when more than one account shares the same `name` across scopes.
data "spectrocloud_cloudaccount_gcp" "example" {
  name    = "example-gcp-account"
  context = "project"
}

# Alternative: look up by ID instead of name.
#
# Lookup keys:
#   id - Exactly one of `id`/`name` required, also Computed.
data "spectrocloud_cloudaccount_gcp" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000"
}

output "gcp_account_id" {
  value = data.spectrocloud_cloudaccount_gcp.example.id
}

output "gcp_account_name" {
  value = data.spectrocloud_cloudaccount_gcp.example.name
}
