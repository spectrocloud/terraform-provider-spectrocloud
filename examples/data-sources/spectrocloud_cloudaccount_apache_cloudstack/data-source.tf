# Retrieve details of an Apache CloudStack cloud account using name
data "spectrocloud_cloudaccount_apache_cloudstack" "example" {
  name    = "apache-cloudstack-account-1" # Required if 'id' is not provided
  context = "project"                     # Optional: Allowed values are "project", "tenant", or "" (default)
}

# Retrieve details of an Apache CloudStack cloud account using ID
data "spectrocloud_cloudaccount_apache_cloudstack" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000" # Required if 'name' is not provided
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

output "cloudstack_api_url" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.api_url
  description = "CloudStack API endpoint URL"
}

output "cloudstack_domain" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.domain
  description = "CloudStack domain for the account"
}

output "private_cloud_gateway_id" {
  value       = data.spectrocloud_cloudaccount_apache_cloudstack.example.private_cloud_gateway_id
  description = "Private Cloud Gateway ID associated with this account"
}

