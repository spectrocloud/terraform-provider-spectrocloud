# Nothing on this resource is ForceNew - every attribute below updates in place.
#
# Attributes:
#   name                     - Required.
#   context                  - Optional, default "project". Allowed: "project", "tenant".
#   private_cloud_gateway_id - Required. The ID of the Private Cloud Gateway used to reach this
#                              CloudStack environment.
#   api_url                  - Required. The CloudStack management server's API endpoint.
#   api_key                  - Required, sensitive. CloudStack API key.
#   secret_key               - Required, sensitive. CloudStack secret key.
#   domain                   - Optional, default "" (the ROOT domain). Set for multi-domain
#                              CloudStack environments.
#   insecure                 - Optional, default false. Skips SSL certificate verification - only
#                              use this for development/testing; CloudStack must have valid
#                              CA-signed certificates otherwise.

data "spectrocloud_private_cloud_gateway" "pcg" {
  name = "System Private Gateway"
}

resource "spectrocloud_cloudaccount_apache_cloudstack" "cloudstack_account" {
  name                     = "ran-tf-cloudstack-account"
  context                  = "project"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id

  api_url    = var.cloudstack_api_url    # gitleaks:allow
  api_key    = var.cloudstack_api_key    # gitleaks:allow
  secret_key = var.cloudstack_secret_key # gitleaks:allow

  domain   = "ROOT"
  insecure = true
}
