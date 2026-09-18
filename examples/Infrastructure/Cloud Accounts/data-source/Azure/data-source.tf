# Looks up an existing Azure cloud account registered in Palette, by name or by ID.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   context - Optional. Allowed: "project", "tenant", "" (default). Required only to
#             disambiguate when more than one account shares the same `name` across scopes.
data "spectrocloud_cloudaccount_azure" "example" {
  name    = "example-azure-account"
  context = "project"
}

# Alternative: look up by ID instead of name.
#
# Lookup keys:
#   id - Exactly one of `id`/`name` required, also Computed.
data "spectrocloud_cloudaccount_azure" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000"
}

output "azure_account_id" {
  value = data.spectrocloud_cloudaccount_azure.example.id
}

output "azure_account_name" {
  value = data.spectrocloud_cloudaccount_azure.example.name
}

# Computed.
output "azure_tenant_id" {
  value = data.spectrocloud_cloudaccount_azure.example.azure_tenant_id
}

output "azure_client_id" {
  value = data.spectrocloud_cloudaccount_azure.example.azure_client_id
}

# Computed.
output "azure_tenant_name" {
  value = data.spectrocloud_cloudaccount_azure.example.tenant_name
}

output "azure_disable_properties_request" {
  value = data.spectrocloud_cloudaccount_azure.example.disable_properties_request
}
