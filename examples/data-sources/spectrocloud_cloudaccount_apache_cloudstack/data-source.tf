# Looks up an existing Apache CloudStack cloud account registered in Palette, by name or by ID.

# Retrieve details of an Apache CloudStack cloud account using name.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   context - Optional. Allowed: "project", "tenant", "" (default). Required only to
#             disambiguate when more than one account shares the same `name` across scopes.
data "spectrocloud_cloudaccount_apache_cloudstack" "example" {
  name    = "apache-cloudstack-account-1"
  context = "project"
}

# Retrieve details of an Apache CloudStack cloud account using ID.
#
# Lookup keys:
#   id - Exactly one of `id`/`name` required, also Computed.
data "spectrocloud_cloudaccount_apache_cloudstack" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000"
}

# Output cloud account details
output "cloudstack_account_id" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.id
  description = "Apache CloudStack cloud account ID"
}

output "cloudstack_account_name" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.name
  description = "Apache CloudStack cloud account name"
}

output "cloudstack_account_context" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.context
  description = "Context scope of the cloud account (project/tenant)"
}

# Computed. There is no api_url or domain attribute on this data source - only the Private
# Cloud Gateway the account routes through is exposed.
output "private_cloud_gateway_id" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.private_cloud_gateway_id
  description = "Private Cloud Gateway ID associated with this account"
}
