# Looks up an existing GCP cloud account registered in Palette, by name or by ID.
data "spectrocloud_cloudaccount_gcp" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "example-gcp-account"
  # Optional. Allowed: "project", "tenant", "" (default). Required only to disambiguate when
  # more than one account shares the same `name` across scopes.
  context = "project"
}

# Alternative: look up by ID instead of name
data "spectrocloud_cloudaccount_gcp" "by_id" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  id = "123e4567-e89b-12d3-a456-426614174000"
}

output "gcp_account_id" {
  value = data.spectrocloud_cloudaccount_gcp.example.id
}

output "gcp_account_name" {
  value = data.spectrocloud_cloudaccount_gcp.example.name
}
