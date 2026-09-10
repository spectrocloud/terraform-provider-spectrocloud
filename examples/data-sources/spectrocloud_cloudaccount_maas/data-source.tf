# Looks up an existing MAAS cloud account registered in Palette, by name or by ID.
data "spectrocloud_cloudaccount_maas" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "example-maas-account"
  # Optional. Allowed: "project", "tenant", "" (default). Required only to disambiguate when
  # more than one account shares the same `name` across scopes.
  context = "project"
}

# Alternative: look up by ID instead of name
data "spectrocloud_cloudaccount_maas" "by_id" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  id = "123e4567-e89b-12d3-a456-426614174000"
}

output "maas_account_id" {
  value = data.spectrocloud_cloudaccount_maas.example.id
}

output "maas_account_name" {
  value = data.spectrocloud_cloudaccount_maas.example.name
}

# Computed.
output "maas_api_endpoint" {
  value = data.spectrocloud_cloudaccount_maas.example.maas_api_endpoint
}

# Computed. Credential material - the schema does not mark this Sensitive, but treat it as a
# secret regardless and mark any output of it sensitive yourself, as done here.
output "maas_api_key" {
  value     = data.spectrocloud_cloudaccount_maas.example.maas_api_key
  sensitive = true
}

output "private_cloud_gateway_id" {
  value = data.spectrocloud_cloudaccount_maas.example.private_cloud_gateway_id
}
