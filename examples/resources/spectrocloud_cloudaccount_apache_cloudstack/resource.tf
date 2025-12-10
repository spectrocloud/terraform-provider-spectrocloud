# # Example 1: Apache CloudStack Account with API Credentials
# resource "spectrocloud_cloudaccount_apache_cloudstack" "cloudstack_account" {
#   name = "apache-cloudstack-account-1"

#   # CloudStack API endpoint
#   api_url = var.cloudstack_api_url

#   # CloudStack API credentials
#   api_key    = var.cloudstack_api_key
#   secret_key = var.cloudstack_secret_key

#   # Optional: CloudStack domain (defaults to ROOT domain if not specified)
#   domain = var.cloudstack_domain

#   # Optional: Skip TLS certificate verification (not recommended for production)
#   # insecure = false

#   # Private Cloud Gateway ID (required for private cloud deployments)
#   private_cloud_gateway_id = var.private_cloud_gateway_id
# }

# # Example 2: Apache CloudStack Account with Insecure Connection
# resource "spectrocloud_cloudaccount_apache_cloudstack" "cloudstack_account_insecure" {
#   name = "apache-cloudstack-account-insecure"

#   api_url    = var.cloudstack_api_url
#   api_key    = var.cloudstack_api_key
#   secret_key = var.cloudstack_secret_key
#   domain     = "ROOT"

#   # Skip TLS verification (use only in development/testing environments)
#   insecure = true

#   private_cloud_gateway_id = var.private_cloud_gateway_id
# }

# # Example 3: Apache CloudStack Account with Specific Domain
# resource "spectrocloud_cloudaccount_apache_cloudstack" "cloudstack_account_domain" {
#   name = "apache-cloudstack-account-custom-domain"

#   api_url    = var.cloudstack_api_url
#   api_key    = var.cloudstack_api_key
#   secret_key = var.cloudstack_secret_key

#   # Specify a custom CloudStack domain
#   domain = "Production"

#   private_cloud_gateway_id = var.private_cloud_gateway_id
# }



data "spectrocloud_private_cloud_gateway" "pcg" {
  name = "System Private Gateway"
}

# Apache CloudStack Cloud Account
resource "spectrocloud_cloudaccount_apache_cloudstack" "cloudstack_account" {
  name                     = "ran-tf-cloudstack-account"
  context                  = "project"                                      # Allowed values: "project" or "tenant". Default is "project"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id # Required: ID of the private cloud gateway

  # CloudStack API Configuration
  api_url    = var.cloudstack_api_url    # Required: e.g., https://cloudstack.example.com:8080/client/api
  api_key    = var.cloudstack_api_key    # Required: API key for CloudStack authentication # gitleaks:allow
  secret_key = var.cloudstack_secret_key # Required: Secret key for CloudStack authentication # gitleaks:allow

  # Optional Configuration
  domain   = "ROOT" # Optional: Domain for multi-domain CloudStack environments. Default is empty (ROOT domain)
  insecure = true   # Optional: Skip SSL certificate verification. Default is false
}
