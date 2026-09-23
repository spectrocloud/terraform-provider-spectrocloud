# Creates and manages a dedicated Apache CloudStack cloud account for this end-to-end example,
# rather than looking up one that already exists - so this folder is fully self-contained and
# every attribute of spectrocloud_cloudaccount_apache_cloudstack is exercised in one place.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute updates in place.
#
# Prerequisite: a Private Cloud Gateway (PCG) must already be installed and reachable from this
# CloudStack management server before this account can be validated.
resource "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  # Required. Display name shown in the Palette UI.
  name = "e2e-apache-cloudstack-account"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required. UID of the Private Cloud Gateway this account routes CloudStack traffic through.
  private_cloud_gateway_id = var.private_cloud_gateway_id
  # Required. The CloudStack management server's API endpoint.
  api_url = var.cloudstack_api_url
  # Required, sensitive. CloudStack API key.
  api_key = var.cloudstack_api_key
  # Required, sensitive. CloudStack secret key.
  secret_key = var.cloudstack_secret_key
  # Optional, default "" (the ROOT domain). Set for multi-domain CloudStack environments.
  domain = "ROOT"
  # Optional, default false. Skips SSL certificate verification - only use this for
  # development/testing; CloudStack must have valid CA-signed certificates otherwise.
  insecure = false
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
#   name = "some-existing-account-name"
# }
