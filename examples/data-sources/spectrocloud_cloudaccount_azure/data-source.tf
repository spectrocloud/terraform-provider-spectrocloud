# Looks up an existing Azure cloud account registered in Palette, by name or by ID.
data "spectrocloud_cloudaccount_azure" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "example-azure-account"
  # Optional. Allowed: "project", "tenant", "" (default). Required only to disambiguate when
  # more than one account shares the same `name` across scopes.
  context = "project"
}

# Alternative: look up by ID instead of name
data "spectrocloud_cloudaccount_azure" "by_id" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
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

# Computed, but not currently populated by the provider's read implementation - expect these to
# come back empty regardless of the account's actual configuration.
output "azure_tenant_name" {
  value = data.spectrocloud_cloudaccount_azure.example.tenant_name
}

output "azure_disable_properties_request" {
  value = data.spectrocloud_cloudaccount_azure.example.disable_properties_request
}
